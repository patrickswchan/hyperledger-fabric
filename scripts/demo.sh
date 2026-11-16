#!/usr/bin/env bash
# Install the Gateway client dependencies and run the invoice walkthrough.
set -euo pipefail

source "$(dirname "$0")/env.sh"

cd "$ROOT_DIR/application"
if [[ ! -d node_modules ]]; then
  npm install
fi

exec node app.js
