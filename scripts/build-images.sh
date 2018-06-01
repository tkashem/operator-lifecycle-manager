#!/usr/bin/env bash

# Note: run from root
# This is used to start and build services for running e2e tests

set -e

REPOSITORY=docker.io/tohinkashem
TAG=local

docker build -f e2e.Dockerfile .
docker tag $(docker images --filter 'label=broker=true' --format '{{.CreatedAt}}\t{{.ID}}' | sort -nr | head -n 1 | cut -f2) ${REPOSITORY}/olm-service-broker:${TAG}
docker tag $(docker images --filter 'label=catalog=true' --format '{{.CreatedAt}}\t{{.ID}}' | sort -nr | head -n 1 | cut -f2) ${REPOSITORY}/catalog:${TAG}
docker tag $(docker images --filter 'label=e2e=true' --format '{{.CreatedAt}}\t{{.ID}}' | sort -nr | head -n 1 | cut -f2) ${REPOSITORY}/olm-e2e:${TAG}
docker tag $(docker images --filter 'label=olm=true' --format '{{.CreatedAt}}\t{{.ID}}' | sort -nr | head -n 1 | cut -f2) ${REPOSITORY}/olm:${TAG}

docker push ${REPOSITORY}/olm-service-broker:${TAG}
docker push ${REPOSITORY}/catalog:${TAG}
docker push ${REPOSITORY}/olm:${TAG}