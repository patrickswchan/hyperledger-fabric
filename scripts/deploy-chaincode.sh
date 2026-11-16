#!/usr/bin/env bash
# Package this repo's invoice chaincode, install it on both peers,
# approve it for both organizations, and commit it to the channel.
set -euo pipefail

source "$(dirname "$0")/env.sh"

if [[ ! -d "$TEST_NETWORK/organizations/peerOrganizations" ]]; then
  echo "Test network crypto material is missing. Run ./scripts/up.sh first." >&2
  exit 1
fi

cd "$TEST_NETWORK"
./network.sh deployCC \
  -c "$CHANNEL_NAME" \
  -ccn "$CC_NAME" \
  -ccp "$CC_SRC_PATH" \
  -ccl "$CC_LANG"

echo
echo "Chaincode ${CC_NAME} is committed on ${CHANNEL_NAME}."
echo "Next: ./scripts/demo.sh"
