# Site harness wrappers

Corporate site delivery binds **only** this directory as `verification_scripts`:

- `verify.sh` → execs `../../scripts/verify.sh`
- `adversarial.sh` → execs `../../scripts/adversarial.sh`

Do not point gates at the whole `scripts/` tree.
