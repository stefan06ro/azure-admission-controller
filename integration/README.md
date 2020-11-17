# Integration tests

Here we have a basic integration test for the happy path.

## Requirements

You need these tools installed locally:

- [kind](https://kind.sigs.k8s.io/)
- [apptestctl](https://github.com/giantswarm/apptestctl/)

## Running the tests

Create the `kind` cluster and use `apptestctl` to bootstrap GiantSwarm's App platform.

```bash
kind delete cluster --name kind-admission-test && kind create cluster --name kind-admission-test && apptestctl bootstrap --kubeconfig="$(kind get kubeconfig --name kind-admission-test)"
```

You can now run the tests.

```bash
E2E_KUBECONFIG=~/.kube/config CIRCLE_SHA1=$(git rev-parse HEAD) AZURE_CLIENTID="${AZURE_CLIENTID}" AZURE_CLIENTSECRET="${AZURE_CLIENTSECRET}" AZURE_TENANTID="${AZURE_TENANTID}" AZURE_SUBSCRIPTIONID="${AZURE_SUBSCRIPTIONID}" go test -tags=k8srequired ./integration/test/... -count=1
```

You can run the tests several times, but the apps won't be re-deployed. This is useful when editing the CR yaml files.
If you need to re-deploy an app, remove it with Helm or start over re-creating the `kind` cluster from scratch.
