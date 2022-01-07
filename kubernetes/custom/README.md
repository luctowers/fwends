# Custom Config

Add custom ConfigMaps and Secrets to this directory, they will be applied after
those in [config](../config). They will be ignored by git.

## Example development config

`fwends-admin.yaml`

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: fwends-admin
data:
  ADMIN_EMAILS: myemail@gmail.com,myfriend@gmail.com
```

`fwends-auth.yaml`

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: fwends-auth
data:
  AUTH_ENABLE: "true"
  GOOGLE_CLIENT_ID: "yourgoogleclientidhere.apps.googleusercontent.com"
```
