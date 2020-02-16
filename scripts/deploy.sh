#!/usr/bin/env bash

set -e
set -o pipefail # Only exit with zero if all commands of the pipeline exit successfully

[[ -z ${BOT_TOKEN} ]] && echo "BOT_TOKEN not defined" && exit 1
[[ -z ${DISCORDRUS_WEBHOOK_URL} ]] && echo "DISCORDRUS_WEBHOOK_URL not defined" && exit 1
[[ -z ${DISCORDBOTS_ORG_BOT_ID} ]] && echo "DISCORDBOTS_ORG_BOT_ID not defined" && exit 1
[[ -z ${DISCORDBOTS_ORG_TOKEN} ]] && echo "DISCORDBOTS_ORG_TOKEN not defined" && exit 1

SCRIPT_PATH=$(readlink -f "${0}")
SCRIPT_DIR=$(dirname "${SCRIPT_PATH}")

COMMIT=$(git rev-parse --short HEAD)
SHARD_COUNT="3"

REPO_YMLS="${SCRIPT_DIR}/../deployments/kubernetes"

NAMESPACE_YML="${REPO_YMLS}/namespace.yml"
SERVICE_YML="${REPO_YMLS}/service.yml"

DEPLOYMENT_YML="${REPO_YMLS}/statefulset.yml"
VARIABLIZED_DEPLOYMENT_YML="/tmp/statefulset.yml"

setup() {
  cp "${DEPLOYMENT_YML}" "${VARIABLIZED_DEPLOYMENT_YML}"
}

applyValues() {
  sed -i "s|{COMMIT}|${COMMIT}|g" "${VARIABLIZED_DEPLOYMENT_YML}"
  sed -i "s|{BOT_TOKEN}|${BOT_TOKEN}|g" "${VARIABLIZED_DEPLOYMENT_YML}"
  sed -i "s|{DISCORDRUS_WEBHOOK_URL}|${DISCORDRUS_WEBHOOK_URL}|g" "${VARIABLIZED_DEPLOYMENT_YML}"
  sed -i "s|{DISCORDBOTS_ORG_BOT_ID}|${DISCORDBOTS_ORG_BOT_ID}|g" "${VARIABLIZED_DEPLOYMENT_YML}"
  sed -i "s|{DISCORDBOTS_ORG_TOKEN}|${DISCORDBOTS_ORG_TOKEN}|g" "${VARIABLIZED_DEPLOYMENT_YML}"
  sed -i "s|{SHARD_COUNT}|${SHARD_COUNT}|g" "${VARIABLIZED_DEPLOYMENT_YML}"
}

deploy() {
  kubectl apply -f "${NAMESPACE_YML}"
  kubectl apply -f "${SERVICE_YML}"
  kubectl apply -f "${VARIABLIZED_DEPLOYMENT_YML}"
  kubectl -n ephemeral-roles rollout status --timeout 60s statefulset/ephemeral-roles
}

cleanup() {
  rm -f "${VARIABLIZED_DEPLOYMENT_YML}"
}

trap cleanup EXIT

setup
applyValues
deploy
