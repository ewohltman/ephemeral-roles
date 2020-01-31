#!/usr/bin/env bash

[[ -z ${BOT_TOKEN} ]] && echo "BOT_TOKEN not defined" && exit 1
[[ -z ${DISCORDRUS_WEBHOOK_URL} ]] && echo "DISCORDRUS_WEBHOOK_URL not defined" && exit 1
[[ -z ${DISCORDBOTS_ORG_BOT_ID} ]] && echo "DISCORDBOTS_ORG_BOT_ID not defined" && exit 1
[[ -z ${DISCORDBOTS_ORG_TOKEN} ]] && echo "DISCORDBOTS_ORG_TOKEN not defined" && exit 1

SCRIPT_PATH=$(readlink -f "${0}")
SCRIPT_DIR=$(dirname "${SCRIPT_PATH}")

REPO_YMLS="${SCRIPT_DIR}/../deployments/kubernetes"

SERVICE_YML="${REPO_YMLS}/service.yml"
INGRESS_YML="${REPO_YMLS}/ingress.yml"

TEMPLATE_DEPLOYMENT_YML="${REPO_YMLS}/deployment.yml"
VARIABLIZED_DEPLOYMENT_YML="/tmp/deployment.yml"

COMMIT=$(git rev-parse --short HEAD)

setup() {
  cp "${TEMPLATE_DEPLOYMENT_YML}" "${VARIABLIZED_DEPLOYMENT_YML}"
}

applyValues() {
  sed -i "s/{COMMIT}/${COMMIT}/g" "${VARIABLIZED_DEPLOYMENT_YML}"
  sed -i "s/{BOT_TOKEN}/${BOT_TOKEN}/g" "${VARIABLIZED_DEPLOYMENT_YML}"
  sed -i "s/{DISCORDRUS_WEBHOOK_URL}/${DISCORDRUS_WEBHOOK_URL}/g" "${VARIABLIZED_DEPLOYMENT_YML}"
  sed -i "s/{DISCORDBOTS_ORG_BOT_ID}/${DISCORDBOTS_ORG_BOT_ID}/g" "${VARIABLIZED_DEPLOYMENT_YML}"
  sed -i "s/{DISCORDBOTS_ORG_TOKEN}/${DISCORDBOTS_ORG_TOKEN}/g" "${VARIABLIZED_DEPLOYMENT_YML}"
}

deploy() {
  kubectl apply -f "${SERVICE_YML}"
  kubectl apply -f "${INGRESS_YML}"
  kubectl apply -f "${VARIABLIZED_DEPLOYMENT_YML}"
}

cleanup() {
  rm -f "${VARIABLIZED_DEPLOYMENT_YML}"
}

trap cleanup EXIT

setup
applyValues
deploy
