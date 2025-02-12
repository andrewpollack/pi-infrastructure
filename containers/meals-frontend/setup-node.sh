#!/usr/bin/env bash

set -e

# Installing nvm...
. ~/.nvm/nvm.sh
if ! nvm use > /dev/null 2>&1; then
  echo "Installing Node version from .nvmrc..."
  nvm install
  nvm use
fi

# Installing npm...
desired_npm_version=$(cat .npm_version)
if npm --version >/dev/null 2>&1; then
  current_npm_version=$(npm --version)
else
  current_npm_version=""
fi
if [ "$current_npm_version" != "$desired_npm_version" ]; then
  echo "Installing npm@$desired_npm_version ..."
  npm install -g "npm@$desired_npm_version"
fi

echo "Node version: $(node -v)"
echo "npm version:  $(npm -v)"
