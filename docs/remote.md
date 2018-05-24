# Remote spec

Here is the complete spec for a Remote custom resource. The definitions are inlined as comments:

```yaml
apiVersion: gitkube.sh/v1alpha1
kind: Remote
metadata:
  name: sampleremote
  namespace: default
spec:

# Insert ssh-keys for allowing users to git push
  authorizedKeys:
  - "ssh-rsa your-ssh-public-key"

# Specify manifests for initialisation
  manifests:
    path: mymanifests                     # Path of the manifests/chart directory in the repo
    helm:
      release: myapp                      # Release name, if using helm initialisation
      values:                             # Set helm values, if using helm initialisation
        name: username
        value: janedoe

# Provide registry details: https://github.com/hasura/gitkube/blob/master/docs/registry.md
  registry:
    url: "docker.io/user"
    credentials:
      secretRef: regsecret                # Name of docker-registry secret

# Define deployment rules
  deployments:
  - name: www                             # Name of K8s deployment which is updated on git push
    containers: 
    - name: www                           # Name of container in the deployment which is built during git push
      path: example/www                   # Docker build context path in the git repo
      dockerfile: example/www/Dockerfile  # Location of Dockerfile for the source code
      buildArgs:                          # Any --build-args during docker build
      - name: SOME_BUILD_ARG
        value: something
```

