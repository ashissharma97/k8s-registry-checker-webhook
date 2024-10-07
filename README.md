# Kubernetes Admission Webhook for Image Registry Validation

This project implements a Kubernetes admission webhook that validates container images in pods to ensure they come from an allowed registry.

## Overview

The webhook server is written in Go and performs the following functions:

1. Listens for admission requests from the Kubernetes API server.
2. Validates that all container images in a pod specification start with an allowed registry name.
3. Allows or denies the admission request based on the validation result.

## Prerequisites

- Go 1.x (where x is the version used in your project)
- Access to a Kubernetes cluster
- `kubectl` configured to communicate with your cluster
- Openssl should be installed
- Optional: Kind also should be installed or you can use any Kubernetes Cluster.

## Installation

1. Clone this repository:
   ```
   git clone <repository-url>
   cd <repository-directory>
   ```

2. Generate Self Signed Certificate:
   ```
   chmod +x ./config/genCerts.sh
   ./config/genCerts.sh
   ```

3. Create a Kind Cluster not required if you are using any Kubernetes :
   ```
   chmod +x ./config/kind.sh
   ./config/kind.sh
   ```
4. Build Docker image of webhook server.
   ```
   docker build -t [registry_name]:[tag] .
   ```
   Mention image name in manifest/deployment.yaml

5. Deploy the webhook server, webhook service and validation webhook to your Kubernetes cluster (manifests files are in manifest folder).
   ```
   kubectl apply -f ./manifest/deployment.yaml -f ./manifest/service.yaml -f ./manifest/webhook.yaml
   ```
6. Test the Webhook whether it's working or not.
   ```
   kubectl apply -f ./manifest/test-pod.yaml
   ```