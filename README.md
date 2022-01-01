# FWENDS 2

## Development Environment (Required)

### Install Minikube

https://minikube.sigs.k8s.io/docs/start/

### Install Skaffold

https://skaffold.dev/docs/quickstart/

### Run services with Skaffold

```shell
skaffold dev -p dev
```

### Access on port 8080

http://localhost:8080/

## Development Environment (Optional)

### Install Go

https://go.dev/doc/install

### Install Node.js

https://nodejs.org/en/download/

### Install eslint

```shell
npm install eslint --global
```

### Install staticcheck

```shell
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### Install pylint

https://pylint.org/#install

### Enable pre-commit hooks

```shell
chmod +x git-hooks/pre-commit
ln -s -f ../../git-hooks/pre-commit .git/hooks/pre-commit
```

## Useful commands

### Delete persistent Postgres volume

```shell
kubectl delete pvc -l app=fwends-postgres
```
