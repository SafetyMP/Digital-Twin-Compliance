#!/usr/bin/env bash
# Shared Phase 2 smoke helpers — source from smoke-test-phase2.sh / verify-state-twin-pipeline.sh
# Expects: CORE_URL, STATE_URL, CONNECT_URL, REDIS_CONTAINER, KAFKA_CONTAINER, TENANT_ID

smoke_twin_pipeline_debug() {
  psql "${STATE_URL}" -tA -c "SELECT COUNT(*) FROM outbox WHERE published_at IS NULL;" 2>/dev/null \
    | awk '{print "unpublished outbox rows: " $1}' >&2 || true
  curl -sf "${CONNECT_URL}/connectors/core-banking-cdc/status" 2>/dev/null \
    | jq '{connector:.connector.state, task:(.tasks[0].state // "unknown")}' >&2 || true
}

redis_get() {
  local key="$1"
  if ! docker ps --format '{{.Names}}' | grep -qx "${REDIS_CONTAINER}"; then
    return 1
  fi
  docker exec "${REDIS_CONTAINER}" redis-cli GET "$key" 2>/dev/null || true
}

wait_redis_gte() {
  local key="$1" min="$2" timeout="${3:-30}" label="${4:-$key}"
  for _ in $(seq 1 "$timeout"); do
    local val
    val="$(redis_get "$key")"
    if [[ -n "$val" && "$val" != "(nil)" ]]; then
      if python3 -c "import sys; sys.exit(0 if float(sys.argv[1]) >= float(sys.argv[2]) else 1)" "$val" "$min"; then
        echo "$label ready: $val (>= $min)"
        return 0
      fi
    fi
    sleep 1
  done
  echo "$label not >= $min within ${timeout}s (key=$key)" >&2
  return 1
}

wait_redis_lt() {
  local key="$1" max="$2" timeout="${3:-30}" label="${4:-$key}"
  for _ in $(seq 1 "$timeout"); do
    local val
    val="$(redis_get "$key")"
    if [[ -n "$val" && "$val" != "(nil)" ]]; then
      if python3 -c "import sys; sys.exit(0 if float(sys.argv[1]) < float(sys.argv[2]) else 1)" "$val" "$max"; then
        echo "$label ready: $val (< $max)"
        return 0
      fi
    fi
    sleep 1
  done
  echo "$label not < $max within ${timeout}s (key=$key)" >&2
  return 1
}

twin_state_version() {
  local persona_id="$1"
  psql "${STATE_URL}" -tA -c "
    SELECT COALESCE(state_version::text, '')
    FROM twin_personas
    WHERE persona_id = '$persona_id';
  " 2>/dev/null | tr -d '[:space:]'
}

wait_twin_state_version_gt() {
  local persona_id="$1" min_version="$2" timeout="${3:-45}" label="${4:-twin $persona_id}"
  for _ in $(seq 1 "$timeout"); do
    local ver
    ver="$(twin_state_version "$persona_id")"
    if [[ -n "$ver" ]] && python3 -c "import sys; sys.exit(0 if int(sys.argv[1]) > int(sys.argv[2]) else 1)" "$ver" "$min_version" 2>/dev/null; then
      echo "$label state_version=$ver (>$min_version)"
      return 0
    fi
    sleep 1
  done
  echo "$label state_version not > $min_version within ${timeout}s (persona=$persona_id, got=${ver:-empty})" >&2
  return 1
}

twin_notional() {
  local persona_id="$1"
  psql "${STATE_URL}" -tA -c "
    SELECT COALESCE(current_state->>'notional_amount', '')
    FROM twin_personas
    WHERE persona_id = '$persona_id' AND persona_type = 'Instrument';
  " 2>/dev/null | tr -d '[:space:]'
}

wait_twin_notional() {
  local persona_id="$1" expected="$2" timeout="${3:-45}" label="${4:-twin $persona_id}"
  for _ in $(seq 1 "$timeout"); do
    local val
    val="$(twin_notional "$persona_id")"
    if [[ -n "$val" ]]; then
      if python3 -c "import sys; sys.exit(0 if abs(float(sys.argv[1]) - float(sys.argv[2])) < 0.01 else 1)" "$val" "$expected" 2>/dev/null; then
        echo "$label mirrored notional=$val"
        return 0
      fi
    fi
    sleep 1
  done
  echo "$label notional != $expected within ${timeout}s (persona=$persona_id, got=${val:-empty})" >&2
  return 1
}

twin_institution_lcr() {
  local persona_id="$1"
  psql "${STATE_URL}" -tA -c "
    SELECT COALESCE(current_state->'liquidity'->>'lcr', '')
    FROM twin_personas
    WHERE persona_id = '$persona_id' AND persona_type = 'Institution';
  " 2>/dev/null | tr -d '[:space:]'
}

wait_twin_lcr_at_most() {
  local persona_id="$1" max_lcr="$2" timeout="${3:-45}" label="${4:-twin $persona_id LCR}"
  for _ in $(seq 1 "$timeout"); do
    local val
    val="$(twin_institution_lcr "$persona_id")"
    if [[ -n "$val" ]]; then
      if python3 -c "import sys; sys.exit(0 if float(sys.argv[1]) <= float(sys.argv[2]) + 1e-6 else 1)" "$val" "$max_lcr" 2>/dev/null; then
        echo "$label mirrored lcr=$val (<= $max_lcr)"
        return 0
      fi
    fi
    sleep 1
  done
  echo "$label lcr not <= $max_lcr within ${timeout}s (persona=$persona_id, got=${val:-empty})" >&2
  return 1
}

wait_twin_lcr_below() {
  local persona_id="$1" max_lcr="$2" timeout="${3:-45}" label="${4:-twin $persona_id LCR}"
  for _ in $(seq 1 "$timeout"); do
    local val
    val="$(twin_institution_lcr "$persona_id")"
    if [[ -n "$val" ]]; then
      if python3 -c "import sys; sys.exit(0 if float(sys.argv[1]) < float(sys.argv[2]) else 1)" "$val" "$max_lcr" 2>/dev/null; then
        echo "$label mirrored lcr=$val (< $max_lcr)"
        return 0
      fi
    fi
    sleep 1
  done
  echo "$label lcr not < $max_lcr within ${timeout}s (persona=$persona_id, got=${val:-empty})" >&2
  return 1
}

wait_outbox_drained() {
  local timeout="${1:-30}"
  for _ in $(seq 1 "$timeout"); do
    local pending
    pending="$(psql "${STATE_URL}" -tA -c "SELECT COUNT(*) FROM outbox WHERE published_at IS NULL;" 2>/dev/null | tr -d '[:space:]')"
    if [[ "${pending:-}" == "0" ]]; then
      echo "outbox drained"
      return 0
    fi
    sleep 1
  done
  echo "outbox not drained within ${timeout}s (pending=${pending:-?})" >&2
  return 1
}
