# Fanet-operator

The FANET Kubernetes Operator is a specialized controller that extends the Kubernetes API to manage Flying Ad Hoc Networks (FANETs). 
A FANET is a network of Unmanned Aerial Vehicles (UAVs), or drones, equipped with computing capabilities. 
This operator introduces the concept of UAVs as first-class citizens within a Kubernetes cluster and provides a mechanism to manage the deployment of network services on them.

## Description

The operator provides Custom Resource Definitions (CRDs) for managing:
- UAVs (Unmanned Aerial Vehicles)
- Virtual Functions (VFs)
- Service Chains (SCs)

## Table of Contents

- [Fanet-operator](#fanet-operator)
  - [Description](#description)
  - [Table of Contents](#table-of-contents)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Deploy on a local testing cluster](#deploy-on-a-local-testing-cluster)
    - [To Uninstall](#to-uninstall)
  - [Running Simulations](#running-simulations)
    - [Example Simulation Scenarios](#example-simulation-scenarios)
  - [Project Distribution](#project-distribution)
    - [By providing a bundle with all YAML files](#by-providing-a-bundle-with-all-yaml-files)
    - [By providing a Helm Chart](#by-providing-a-helm-chart)
  - [Contributing](#contributing)
  - [License](#license)

## Getting Started

### Prerequisites
- go version v1.23.0+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.
- Kind (Kubernetes in Docker) of local testing

### Deploy on a local testing cluster

**Make a kind cluster to test the operator locally:**
```sh
make kind-cluster
```

**Build your image to the location specified by `IMG`:**
> [!NOTE]
> Building stage may require up to 10 minutes

The `IMG` parameter is optional but required if you want to push your version of the opereator on your dockerhub
```sh
make docker-build IMG=<some-registry>/fanet-operator:tag
# optional
# make docker-push IMG=<some-registry>/fanet-operator:tag
```

> [!NOTE]
> This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don't work.

**Load the images on Kind**
```sh
make kind-load
```

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
# default IMG massigollo/fanet-operator
make deploy IMG=<some-registry>/fanet-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Running Simulations

The operator includes a simulation system to test UAV behavior, energy consumption, and network metrics.
Run a reproducible simulation:

1. Deploy the simulator:
```bash
./hack/simulate.sh deploy
```

2. Create a test UAV with initial parameters:
```bash
./hack/simulate.sh create test-uav 100 10.5 5.2
```
This creates a UAV with:
- 100% battery level
- 10.5W motor idle consumption
- 5.2W CE idle consumption

3. Watch the UAV status changes:
```bash
./hack/simulate.sh watch
```

The simulator will:
- Gradually decrease battery level
- Add random fluctuations to energy consumption
- Update metrics every second
- Handle multiple UAVs if created

4. Clean up resources when done:
```bash
./hack/simulate.sh cleanup
```

### Example Simulation Scenarios

1. **Basic UAV Monitoring**
```bash
# Deploy simulator
./hack/simulate.sh deploy

# Create a UAV with high battery and low consumption
./hack/simulate.sh create uav1 100 5.0 2.5

# Create another UAV with low battery and high consumption
./hack/simulate.sh create uav2 30 15.0 8.0

# Watch the status changes
./hack/simulate.sh watch
```

2. **Multiple UAVs with Different Configurations**
```bash
# Deploy simulator
./hack/simulate.sh deploy

# Create UAVs with different configurations
./hack/simulate.sh create uav-efficient 100 8.0 4.0
./hack/simulate.sh create uav-standard 100 12.0 6.0
./hack/simulate.sh create uav-powerful 100 15.0 8.0

# Watch the status changes
./hack/simulate.sh watch
```

The simulator will automatically:
- Decrease battery levels over time
- Add random fluctuations to energy consumption
- Update network metrics
- Handle multiple UAVs simultaneously


## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/fanet-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/fanet-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
operator-sdk edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.