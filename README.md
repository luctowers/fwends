# FWENDS

[![CI](https://github.com/luctowers/fwends/actions/workflows/ci.yaml/badge.svg)](https://github.com/luctowers/fwends/actions/workflows/ci.yaml)
[![Lint](https://github.com/luctowers/fwends/actions/workflows/linting.yaml/badge.svg)](https://github.com/luctowers/fwends/actions/workflows/linting.yaml)

## Development Setup

### Minimal Setup

1. install skaffold https://skaffold.dev/docs/quickstart/
2. run services with dev profile `skaffold dev -p dev`
3. open a browser http://localhost:8080/

### Optional Setup

- setup pre-commit git hook linting [git-hooks/README.md](./git-hooks/README.md)
- learn how to run integration and failure tests [fwends-test/README.md](./fwends-test/README.md)
- set custom kubernetes configmaps and secrets [kubernetes/custom/README.md](./kubernetes/custom/README.md)
- read useful commands reference [docs/commands.md](./docs/commands.md)
