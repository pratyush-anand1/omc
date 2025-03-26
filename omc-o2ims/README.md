#  omc-o2ims

A scalable, highly available microservice acting as an interface between SMO/FOCOM and OMC systems. It facilitates seamless communication and data flow between these components while being cloud-native for easy deployment across various environments. The stateless design ensures scalability, fault tolerance, and simple deployment.

# Project Structure Detailed Explanation

## Root Level Components

### `cmd/`
- Primary directory for executable applications
- `server/main.go`: Main entry point of the application that initializes and starts all components (controller, operator, services)

### `config/`
- Handles all configuration management
- `config.yaml`: Configuration file
- `crd-list/`: Directory containing CRD instances list
- `crds/`: Directory containing CRDs
- `kubeconfig`: Automatically generated when we run setup-dev-env.sh, to ensurethe applcaition can access k8s

### `coverage/`
- Generated Directory containing coverage reports 

### `deployments/`
- Contains all deployment-related configurations:

#### `helm/` 
- (TBA) Helm chart for Kubernetes deployment
- `Chart.yaml`: Metadata about the Helm chart
- `templates/`: Kubernetes resource templates

#### `kubernetes/`
- (TBA) Raw Kubernetes manifests:
- `configmap.yaml`: Configuration data
- `deployment.yaml`: Kubernetes deployment configuration
- `service.yaml`: Kubernetes service definition

### `docs/`
- Project documentation:
- `api.md`: API specifications and endpoints
- `design.md`: Architecture and design decisions

### `go.mod` and `go.sum`
- Go module definition and dependency tracking

### `internal/`
- Private application code not meant to be imported by other projects:

#### `config/`
- `config.go`: Defines configuration structures and loading logic
- `config_test.go`: Tests for configuration validation and loading
- `test_data/`: Test data for configuration

#### `operator/`
- `remote/`: Remote service implementation
- `resource/`: Resource implementation
- `store/`: Store implementation
- `watcher/`: Watcher implementation

#### `service/`
- `omc_rest/`: OMC REST implementation

### `pkg/`
- Public code that can be imported by other projects:

#### `apis/o2ims/v1alpha1/`
- Custom Resource Definitions (CRDs):

### `scripts/`
- Utility scripts:
- `build.sh`: Build automation script
- `setup-dev-env.sh`: Setup development environment script
- `test.sh`: Test automation script

### `test/`
- Complex tests that require special setup:

#### `e2e/`
- End-to-end tests simulating real user scenarios

#### `integration/`
- Integration tests between components

### `Dockerfile`
- Multi-stage build configuration for creating the container image

### `go.mod` and `go.sum`
- Go module definition and dependency tracking

### `Makefile`
- Build automation and common development tasks

### `README.md`
- Project documentation, setup instructions, and usage guidelines

## Developers

# `cloning`
git clone git@git.rosetta.ericssondevops.com:gyan.a.ranjan/omc-o2ims.git


# NOTE
Make sure that you have Rosetta setup in your system to build and run this project.
https://rosetta.pages.rosetta.ericssondevops.com/rosetta-dashboard/login-support/

Make sure that Minikube with Docker is properly configured and accessible on your development environment.


## Build

Build the project using the Makefile.

## Prerequisites


- Ubuntu 22.04.5 LTS: https://releases.ubuntu.com/22.04/  in WSL2
- Minikube with Docker diriver for CRD related testing
- Trivy
- Go
- Govulncheck


### Building the project

To build the project, run the following commands:
first time g to the scripts dir and run setup-dev-env.sh, this will ensure your current kubeconfigs info is copied in config directorry 

- `make test` runs all the tests and calculates test coverage
- `make lint` ensures that the code is properly formatted and lints are executed
- `make build` compiles the application and generates the executable
- `make docker-build` builds the Docker image for the application
- `make docker-run` starts the application in a Docker container
- `make docker-stop` stops the application running in a Docker container
