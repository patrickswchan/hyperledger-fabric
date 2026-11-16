#!/usr/bin/env bash
# Stop the test network and delete its containers, crypto material, and chaincode images.
set -euo pipefail

source "$(dirname "$0")/env.sh"

if [[ ! -x "$TEST_NETWORK/network.sh" ]]; then
  echo "Nothing to stop: ${TEST_NETWORK} is not installed."
  exit 0
fi

cd "$TEST_NETWORK"
./network.sh down
