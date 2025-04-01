#!/bin/bash
set -euo pipefail

# Check for required commands
for cmd in kubectl curl awk grep; do
  command -v "$cmd" >/dev/null 2>&1 || { echo "$cmd is required but not installed."; exit 1; }
done

# Usage: ./update-pods.sh <pod_name> <registry_url>
POD_NAME=${1:-""}
REGISTRY=${2:-""}
if [ -z "${POD_NAME}" ] || [ -z "${REGISTRY}" ]; then
  echo "Usage: $0 <pod_name> <registry_url>"
  exit 1
fi

# Function to get a matching pod name (first match)
get_pod() {
  kubectl get pods --no-headers | grep -F "${POD_NAME}" | awk '{print $1}' | head -n 1
}

POD=$(get_pod)
if [ -z "${POD}" ]; then
  echo "No pod matching '${POD_NAME}' found."
  exit 1
fi

# Retrieve the image SHA from the pod
POD_IMAGE_SHA=$(kubectl get pod "${POD}" -o jsonpath='{.status.containerStatuses[0].imageID}' | cut -d'@' -f2)
if [ -z "${POD_IMAGE_SHA}" ]; then
  echo "Could not determine image SHA for pod ${POD}."
  exit 1
fi

# Get the digest from the registry
REGISTRY_CURL_RESULT=$(curl -sI --fail -H "Accept: application/vnd.docker.distribution.manifest.v2+json" "${REGISTRY}")
DIGEST=$(echo "${REGISTRY_CURL_RESULT}" | grep -i "Docker-Content-Digest:" | awk '{print $2}' | tr -d '\r')
if [ -z "${DIGEST}" ]; then
  echo "Failed to retrieve digest from the registry."
  exit 1
fi

if [ "${DIGEST}" != "${POD_IMAGE_SHA}" ]; then
  echo "🔃 Image SHA for ${POD_NAME} has changed. Deleting pod so new image will pull..."
  kubectl delete pod "${POD}" > /dev/null 2>&1

  # Wait for a new pod to be created and become running.
  echo "⌛ Waiting for pod to be running..."
  TIMEOUT=60
  elapsed=0
  while [ $elapsed -lt $TIMEOUT ]; do
    NEW_POD=$(get_pod)
    if [ -n "${NEW_POD}" ]; then
      POD_PHASE=$(kubectl get pod "${NEW_POD}" -o jsonpath='{.status.phase}')
      if [ "${POD_PHASE}" == "Running" ]; then
        echo "✅ Pod ${NEW_POD} is running."
        exit 0
      else
        echo "⌛ Pod ${NEW_POD} is in ${POD_PHASE} phase. Waiting..."
      fi
    else
      echo "⌛ Waiting for pod to be created..."
    fi
    sleep 2
    elapsed=$((elapsed + 2))
  done

  echo "⏰ Timeout waiting for pod to become running."
  exit 1
else
  echo "✅ Image SHA for ${POD_NAME} has not changed. No action needed."
fi
