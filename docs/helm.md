# Helm

## Use Helm for initialisation

Gitkube can install a Helm chart as an initialisation step. This step runs before the build process. To install a Helm chart, define the `manifests` spec in the `Remote` CR as below. Assume the chart directory is called `mychart` and stored in the base of the repo.

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
    path: “mychart”              # Path of the chart directory relative to the repo
    helm:
      release: myapp             # (Optional) Name of the release (recommended)
      values:                    # (Optional) Override values during install
      - name: runAsUser
        value: janedoe
  deployments:
  - name: myapp-www                            
    containers: 
    - name: www                          
      path: example/www                  
      dockerfile: example/www/Dockerfile 
```

You can specify the release name and set values for the chart in the `manifests` spec. 

Note that you must define the full name of the deployment in the `deployments` spec of the `Remote` CR. This means that you must include the release-name if its part of your deployment name. If you don't specify a release name in the `Remote` CR, then you may not be able to specify the full deployment name (if release-name is part of the deployment name). Hence, it is recommended to define the name of the release in the `manifests` spec.


