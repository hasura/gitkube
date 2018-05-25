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

## Elastic Container Registry

1. Use the `aws` cli to generate the registry docker login command:
```
aws ecr get-login
> docker login -u AWS -p <password> -e none https://<aws_account_id>.dkr.ecr.<region>.amazonaws.com
```

Note the output generated. This will be used in subsequent steps.

2. Create a docker-registry secret with the following command:

```
kubectl create secret docker-registry regsecret \
--docker-server=https://<aws_account_id>.dkr.ecr.<region>.amazonaws.com \
--docker-username="AWS" \
--docker-password=<password> \
--docker-email=<email>
```
3. In your remote.yaml, configure the `registry` section as follows

```
registry:
  url: <aws_account_id>.dkr.ecr.<region>.amazonaws.com
  credentials:
    secretRef: regsecret
```

Note 1: ECR does not automatically create repositories on `docker push`, therefore you must create a repository for all containers that are built in gitkube beforehand. The name of the repository is important and must follow the scheme: `<remote-ns>-<remote-name>-<remote-ns>.<deployment-name>-<container-name>`  For instance, consider the following `remote.yaml`

```
apiVersion: gitkube.sh/v1alpha1
kind: Remote
metadata:
  name: dev
  namespace: default
spec:
  authorizedKeys:
  - "ssh-rsa your-ssh-public-key"
  registry:
    url: 27xyz662807.dkr.ecr.us-west-2.amazonaws.com
    credentials:
      secretRef: regsecret
  deployments:
  - name: www
    containers: 
    - name: www
      path: example/www
      dockerfile: example/www/Dockerfile
```

You must create a repository with the name `default-dev-default.www-www` in your registry. 

Note 2: The docker authentication credentials for ECR last only for 12 hours. You must ensure to renew the credentials by running Step 1) repeatedly and updating the docker-registry secret created in Step 2). This can be automated by running a CronJob as described in this [blog post](https://medium.com/@xynova/keeping-aws-registry-pull-credentials-fresh-in-kubernetes-2d123f581ca6).
