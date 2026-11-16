#!/usr/bin/env bash
# Make sure a local Docker daemon is reachable. Fabric peers launch chaincode
# containers through the daemon, so the socket has to be usable by this user.
set -euo pipefail

if docker info >/dev/null 2>&1; then
  exit 0
fi

if [[ ! -S /var/run/docker.sock ]] && command -v sudo >/dev/null 2>&1; then
  echo "Docker socket is missing. Starting dockerd."
  sudo mkdir -p /var/log
  if ! pgrep -x dockerd >/dev/null 2>&1; then
    sudo sh -c 'nohup dockerd --host=unix:///var/run/docker.sock --storage-driver=overlay2 >/tmp/dockerd.log 2>&1 &'
    for _ in $(seq 1 30); do
      [[ -S /var/run/docker.sock ]] && break
      sleep 1
    done
  fi
fi

if [[ -S /var/run/docker.sock ]] && command -v sudo >/dev/null 2>&1; then
  sudo chmod 666 /var/run/docker.sock || true
fi

if ! docker info >/dev/null 2>&1; then
  echo "Docker is not available. Install Docker and start the daemon, then rerun this script." >&2
  echo "On a machine without systemd: sudo dockerd --host=unix:///var/run/docker.sock" >&2
  exit 1
fi
