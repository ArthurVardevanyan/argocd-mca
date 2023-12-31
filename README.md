# ArgoCD MCA (Multi Cluster Authentication) (WIP)

The idea is to automatically provide a central ArgoCD Instance Access to Multiple Namespace in a Remote Cluster using Zero Trust Methods

## Incepting Controller

How to Repo was setup

- <https://book.kubebuilder.io/cronjob-tutorial/cronjob-tutorial.html>
- <https://book.kubebuilder.io/cronjob-tutorial/new-api.html>

```bash
kubebuilder init --domain arthurvardevanyan.com --repo github.com/ArthurVardevanyan/argocd-mca
kubebuilder create api --group argocd --version v1beta1 --kind ServiceAccount --namespaced=true
kubebuilder create api --group argocd --version v1beta1 --kind Cluster --namespaced=true
```

## Getting Started

### Building Image

Build and push your image to the location specified by `IMG`:

```bash
go get -u
go mod tidy
```

```bash
# https://catalog.redhat.com/software/containers/ubi9/ubi-minimal/
export KO_DEFAULTBASEIMAGE=registry.access.redhat.com/ubi9-minimal:9.2-691
export DATE=$(date --utc '+%Y%m%d-%H%M')

# For OpenShift
export KO_DOCKER_REPO=registry.arthurvardevanyan.com/homelab/argocd-mca
# Test Builds / Release Builds
export EXPIRE=1w # 26w
ko build --platform=linux/amd64 --bare --sbom none --image-label quay.expires-after="${EXPIRE}" --tags "${DATE}" # Quay Doesn't Support SBOM KO Yet
```

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/argocd-mca:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/argocd-mca:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
