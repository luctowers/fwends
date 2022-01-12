# FWENDS Useful commands

## Delete persistent Postgres volume

```shell
kubectl delete pvc -l app=fwends-postgres
```

## Delete persistent Minio volume

```shell
kubectl delete pvc -l app=fwends-minio
```

## Prune images when using minikube docker backend

```shell
minikube ssh -- docker system prune
```
