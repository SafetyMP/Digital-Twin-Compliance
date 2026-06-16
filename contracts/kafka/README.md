# Kafka payload contracts

Every message shape on a Kafka topic consumed across services is a **published API**.
Changes require updating the golden fixture here and **both** sides of the boundary:

| Topic | Publisher | Consumer | Golden fixture(s) |
|-------|-----------|----------|-------------------|
| `twin.state.updated` | state-service (outbox) | Flink CEP | `twin.state.updated/*.payload.json` |
| `domain.events.public.payments` | Debezium (core banking) | Flink CEP | `domain.events.public.payments/*.cdc.json` |
| `compliance.alerts` | Flink CEP | alert-service | `compliance.alerts/*.envelope.json` |

## Layout

- `*.payload.json` — inner twin-state body Flink parses (`personaId`, `personaType`, `stateVersion`, `currentState`)
- `cdc/*.after.json` — Debezium `after` row used by Go publisher contract tests to **produce** the payload golden
- `*.cdc.json` — full Debezium envelope for payment CDC
- `*.envelope.json` — full envelope (`TwinStateUpdated` or `ComplianceAlertRaised`) for cross-service parsing tests

## When you change a payload

1. Edit the golden JSON in this directory (deliberate contract change).
2. Update publisher test (Go or Java) so generation matches golden.
3. Update consumer test (Java or Go) so parsing still passes.
4. Run `./scripts/check-kafka-contracts.sh` (also runs in CI via `go test` / `mvn test`).

Do not duplicate fixtures under `services/` or `jobs/` — import from here.
