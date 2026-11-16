#!/usr/bin/env bash
# Shared paths for the local Fabric walkthrough.
# shellcheck disable=SC2034

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
FABRIC_HOME="${FABRIC_HOME:-$ROOT_DIR/.fabric}"
FABRIC_SAMPLES="$FABRIC_HOME/fabric-samples"
TEST_NETWORK="$FABRIC_SAMPLES/test-network"

FABRIC_VERSION="${FABRIC_VERSION:-3.1.5}"
FABRIC_CA_VERSION="${FABRIC_CA_VERSION:-1.5.22}"
CHANNEL_NAME="${CHANNEL_NAME:-mychannel}"
CC_NAME="${CC_NAME:-invoice}"
CC_SRC_PATH="$ROOT_DIR/chaincode/invoice"
CC_LANG="${CC_LANG:-go}"

export PATH="$FABRIC_SAMPLES/bin:${PATH:-}"
export FABRIC_CFG_PATH="$FABRIC_SAMPLES/config"
