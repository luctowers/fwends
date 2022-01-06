# FWENDS Useful commands

## Delete persistent Postgres volume

```shell
kubectl delete pvc -l app=fwends-postgres
```

## Delete persistent Minio volume

```shell
kubectl delete pvc -l app=fwends-minio
```
