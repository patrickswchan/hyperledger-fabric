#!/usr/bin/env bash
# The test network normally puts peers and the orderer on one Docker bridge.
# Some environments accept traffic to published ports but drop packets between
# containers. This script detects that and switches those services to the host
# network, which still publishes the same localhost ports.
set -euo pipefail

source "$(dirname "$0")/env.sh"

probe_bridge() {
  local ip
  docker network rm fabric-icc-probe >/dev/null 2>&1 || true
  docker network create fabric-icc-probe >/dev/null
  docker run -d --name fabric-icc-a --network fabric-icc-probe busybox sleep 20 >/dev/null
  ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' fabric-icc-a)
  if docker run --rm --network fabric-icc-probe busybox ping -c 1 -W 2 "$ip" >/dev/null 2>&1; then
    docker rm -f fabric-icc-a >/dev/null 2>&1 || true
    docker network rm fabric-icc-probe >/dev/null 2>&1 || true
    return 0
  fi
  docker rm -f fabric-icc-a >/dev/null 2>&1 || true
  docker network rm fabric-icc-probe >/dev/null 2>&1 || true
  return 1
}

if probe_bridge; then
  echo "Docker bridge forwarding works. Leaving the test network on its bridge."
  exit 0
fi

echo "Docker cannot forward between containers. Using host networking for peers, the orderer, and chaincode."
cp "$ROOT_DIR/network/hostnet/compose-test-net.yaml" "$TEST_NETWORK/compose/compose-test-net.yaml"
cp "$ROOT_DIR/network/hostnet/docker-compose-test-net.yaml" "$TEST_NETWORK/compose/docker/docker-compose-test-net.yaml"

if ! grep -q "peer0.org1.example.com" /etc/hosts; then
  echo "127.0.0.1 orderer.example.com peer0.org1.example.com peer0.org2.example.com" | sudo tee -a /etc/hosts >/dev/null
fi
