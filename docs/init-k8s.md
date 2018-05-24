# Initialisation options

Gitkube requires K8s resources to be initialised once. K8s resources like deployments can be initialised with some base or dummy image. These images then updated by Gitkube on each git push. 

Gitkube can automatically initialise K8s resources if the manifests are kept in a single directory in the source repo. There are 2 types of manifests which can be initialised:

## K8s yaml

K8s yamls (K8s resource specs) can be initialised by Gitkube. Assuming the yamls are stored in a directory called `k8s` in the base of the repo, the `Remote` CR should define a `manifests` spec as follows:

```yaml
apiVersion: gitkube.sh/v1alpha1
kind: Remote
metadata:
  name: myremote
  namespace: default
spec:
  authorizedKeys:
  - "ssh-rsa <key>"
  manifests:
    path: “k8s”              # Path of the manifests directory relative to the repo
  deployments:
  - name: www                            
    containers: 
    - name: www                          
      path: example/www                  
      dockerfile: example/www/Dockerfile 
```

Now during every `git push` the manifests are applied first before proceeding with the build.

## Helm

Gitkube can also install a Helm chart during initialisation. Refer to the [helm docs](helm.md) for details.

