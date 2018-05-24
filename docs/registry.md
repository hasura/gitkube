# Registry configuration

Gitkube builds docker images in the cluster and pushes it to a registry. You need to provide registry credentials to gitkube so that it has access to push and pull to/from the registry. You can do this by creating a K8s secret of type `docker-registry` and giving its name in the `Remote` resource. Following are detailed instructions for few registry providers:

## Docker Hub

1. Create a `docker-registry` secret with the following command:

```
kubectl create secret docker-registry regsecret \
--docker-server=https://index.docker.io/v1/ \
--docker-username='<username>' \
--docker-password='<password>' \
--docker-email='<email>'
```
2. In your remote.yaml, configure the `registry` section as follows

```
registry:
  url: docker.io/<user>
  credentials:
    secretRef: regsecret
```
## Google Container Registry 

1. Download the JSON key for your Container Registry service account : https://support.google.com/cloud/answer/6158849#serviceaccounts (assume the json file is called `keyfile.json`)

2. Create a docker-registry secret with the following command:

```
kubectl create secret docker-registry regsecret \
--docker-server=https://gcr.io \
--docker-username="_json_key" \
--docker-password="$(cat keyfile.json)" \
--docker-email=<email>
```
3. In your remote.yaml, configure the `registry` section as follows

```
registry:
  url: gcr.io/<project-id>
  credentials:
    secretRef: regsecret
```

