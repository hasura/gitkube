# e2e test

A simple e2e test is defined in `test.bash`. It creates two namespaces, one for `gitkube` and another one for test deployment, installs gitkube, creates and pushes a test deployment.

## Running the test

```bash
$ export KUBECONFIG="[kubeconfig file]"
$ ./test.bash
```

## Tearing down

```bash
$ ./teardown.bash
```
