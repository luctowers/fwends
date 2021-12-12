# FWENDS

## Development Environment (Required)

### Install Docker

https://docs.docker.com/get-docker/

### Run services with docker compose, auto-reload configured

```shell
docker compose up
```

## Development Environment (Optional)

### Install Rust

https://www.rust-lang.org/tools/install

### Install RustFmt

```shell
rustup component add rustfmt
```

### Install Node.js

https://nodejs.org/en/download/

### Install eslint

```shell
npm install eslint --global
```

### Enable pre-commit hooks

```shell
chmod +x git-hooks/pre-commit
ln -s -f ../../git-hooks/pre-commit .git/hooks/pre-commit
```
