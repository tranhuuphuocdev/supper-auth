#!/usr/bin/env sh
set -eu

DOCKER_USER=tranhuuphuoc22
TAG=auth-service-1.0
IMAGE="${DOCKER_USER}/hthouse:${TAG}"

echo "Building image: ${IMAGE} (linux/amd64)"
docker buildx build --platform linux/amd64 -t "${IMAGE}" --push .

echo "Done: ${IMAGE}"
