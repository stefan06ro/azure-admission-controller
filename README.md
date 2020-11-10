[![CircleCI](https://circleci.com/gh/giantswarm/azure-admission-controller.svg?style=svg)](https://circleci.com/gh/giantswarm/azure-admission-controller)

# Azure Admission Controller

Giant Swarm Control Plane admission controller for Azure that implements the following rules:

- Check for TC upgrades to avoid skipping major or minor releases. 

The certificates for the webhook are created with CertManager and injected through the CA Injector.

## Ownership

Team Celestial

## CRs and fields managed

See [docs/mutating.md](https://github.com/giantswarm/azure-admission-controller/blob/master/docs/mutating.md) and [docs/validating.md](https://github.com/giantswarm/azure-admission-controller/blob/master/docs/validating.md)

### Local Development

Testing the azure-admission-controller in a kind cluster on your local machine:

```nohighlight
kind create cluster

# Build a linux image
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
docker build . -t azure-admission-controller:dev
kind load docker-image azure-admission-controller:dev

# Make sure the Custom Resource Definitions are in place
opsctl ensure crds -k "$(kind get kubeconfig)" -p azure

# Insert the certificate
kubectl apply --context kind-kind -f local_dev/certmanager.yml

## Wait until certmanager is up

kubectl apply --context kind-kind -f local_dev/clusterissuer.yml
helm template azure-admission-controller -f helm/azure-admission-controller/ci/default-values.yaml helm/azure-admission-controller > local_dev/deploy.yaml

## Replace image name with azure-admission-controller:dev
kubectl apply --context kind-kind -f local_dev/deploy.yaml
kind delete cluster
```

## Changelog

See [Releases](https://github.com/giantswarm/azure-admission-controller/releases)

## Contact

- Bugs: [issues](https://github.com/giantswarm/azure-admission-controller/issues)
- Please visit https://www.giantswarm.io/responsible-disclosure for information on reporting security issues.

## Contributing, reporting bugs

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches, the
contribution workflow as well as reporting bugs.

## Publishing a release

See [docs/Release.md](https://github.com/giantswarm/azure-admission-controller/blob/master/docs/release.md)

## Add a new webhook

See [docs/webhook.md](https://github.com/giantswarm/azure-admission-controller/blob/master/docs/webhook.md)

## Writing tests

See [docs/tests.md](https://github.com/giantswarm/azure-admission-controller/blob/master/docs/tests.md)
