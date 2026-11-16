#!/usr/bin/env bash
# Start the two-organization test network and create a channel.
#
# Topology (all containers on this machine):
#   peer0.org1.example.com   localhost:7051   Org1MSP
#   peer0.org2.example.com   localhost:9051   Org2MSP
#   orderer.example.com      localhost:7050   Raft orderer
#   ca_org1 / ca_org2                         Fabric CAs that enroll users
set -euo pipefail

source "$(dirname "$0")/env.sh"
"$(dirname "$0")/install-fabric.sh"
"$(dirname "$0")/configure-runtime-network.sh"

cd "$TEST_NETWORK"
./network.sh up createChannel -c "$CHANNEL_NAME" -ca

echo
echo "Channel ${CHANNEL_NAME} is ready."
echo "User1@org1.example.com and User1@org2.example.com were enrolled by the Fabric CAs."
echo "Next: ./scripts/deploy-chaincode.sh"
