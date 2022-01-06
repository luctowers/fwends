# FWENDS Git Hooks

## Prerequisites

1. install go https://go.dev/doc/install
2. install go dependencies `cd fwends-backend && go install`
3. install staticheck `go install honnef.co/go/tools/cmd/staticcheck@latest`
4. install nodejs https://nodejs.org/en/download/
5. install nodejs dependencies `cd fwends-frontend && npm install`
6. install python https://www.python.org/downloads/
7. install python poetry https://python-poetry.org/docs/#installation
8. install python dependencies `cd fwends-test && poetry install`

## Enable git hooks with symlink

```shell
chmod +x git-hooks/pre-commit
ln -s -f ../../git-hooks/pre-commit .git/hooks/pre-commit
```
