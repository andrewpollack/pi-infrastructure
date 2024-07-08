#!/usr/bin/bash

curl -sfL https://get.k3s.io | K3S_NODE_NAME="{{ inventory_hostname }}" K3S_URL="{{ k3s_url }}" K3S_TOKEN="{{ k3s_token }}" sh -
