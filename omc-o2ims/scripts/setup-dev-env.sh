#!/bin/bash


# Get the directory where this script is located
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Get the config directory
CONFIG_DIR="$DIR/../config"

echo "Installing dependencies..."

echo "Checking if Docker is installed and running..."
if ! docker stats --no-stream > /dev/null 2>&1; then
    echo "Docker is not installed or not running. Please install and start Docker."
    exit
fi

echo "Getting current kubernetes context and saving it to ${CONFIG_DIR}/kubeconfig"
kubectl config view --minify --flatten --context=$(kubectl config current-context)   > ${CONFIG_DIR}/kubeconfig

echo "All set up!"
