# Load profile (Phase 6 / UR-CECT-002)

## Target

10K evt/s CDC → twin → CEP path (reference ambition).

## CI-sized substitute (default honesty)

Default CI/laptop hosts run existing smoke suites only. They do **not** prove
10K evt/s. Residual is documented in `docs/deployment.md`.

Optional scaled script (when hardware allows):

```bash
./scripts/load-profile/run-ci-sized.sh
```

This emits a small burst and writes `evidence/phase6-load-*.txt` with measured
throughput — never rewrite the evidence to claim 10K without numbers.
