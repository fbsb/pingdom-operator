
# Image URL to use all building/pushing image targets
IMG ?= fbsb/pingdom-operator:latest

all: vendor test manager

# Vendor dependencies
vendor: go.mod go.sum
	GO111MODULE=on go mod vendor

# Run tests
test: generate fmt vet manifests
	go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager github.com/fbsb/pingdom-operator/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crds

# Make secrets
config/secret/pingdom-credentials.env: config/secret/pingdom-credentials.env.dist
	cp config/secret/pingdom-credentials.env.dist config/secret/pingdom-credentials.env

secrets: config/secret/pingdom-credentials.env
	@echo "Add your pingdom api credentials to config/secret/pingdom-credentials.env before deploying"

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy:
	kubectl apply -f config/namespace.yaml
	kustomize build config | kubectl apply -f -

remove:
	kustomize build config | kubectl delete -f -
	kubectl delete -f config/namespace.yaml

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

# Generate code
generate:
ifndef GOPATH
	$(error GOPATH not defined, please define GOPATH. Run "go help gopath" to learn more about GOPATH)
endif
	go generate ./pkg/... ./cmd/...

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go all

# Build the docker image
docker-build: test
	docker build . -t ${IMG}
	cd config && kustomize edit set image manager=${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}
