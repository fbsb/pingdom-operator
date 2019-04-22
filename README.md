# pingdom-operator
A kubernetes operator used to manage [pingdom](https://www.pingdom.com/) http checks.

# Deploy

First we need to create a file for storing our pingdom credentials. 
The file is used by [kustomize](https://kustomize.io/) when generating the kubernetes secret resource.

```
make secrets
```

After creating the file add your pingdom credentials to `config/secret/pingdom-credentials.env`

```
make docker-build deploy
```