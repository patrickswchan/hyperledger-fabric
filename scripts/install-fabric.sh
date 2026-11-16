#!/usr/bin/env bash
# Download Fabric binaries, the fabric-samples repository, and the runtime images.
# Those artifacts stay in .fabric/ and are not committed.
set -euo pipefail

source "$(dirname "$0")/env.sh"
source "$(dirname "$0")/ensure-docker.sh"

mkdir -p "$FABRIC_HOME"
cd "$FABRIC_HOME"

if [[ ! -x install-fabric.sh ]]; then
  curl -fsSL -o install-fabric.sh https://raw.githubusercontent.com/hyperledger/fabric/main/scripts/install-fabric.sh
  chmod +x install-fabric.sh
fi

if [[ ! -x "$FABRIC_SAMPLES/bin/peer" ]]; then
  ./install-fabric.sh -f "$FABRIC_VERSION" -c "$FABRIC_CA_VERSION" binary docker samples
fi

peer version
docker image inspect "hyperledger/fabric-peer:${FABRIC_VERSION}" >/dev/null
echo "Fabric ${FABRIC_VERSION} is installed under ${FABRIC_HOME}"
