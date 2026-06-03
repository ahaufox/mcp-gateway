#!/bin/bash
set -e

export NODE_OPTIONS="--max-old-space-size=4096"

echo "=== Updating submodules ==="
git submodule update --init --recursive

echo "=== Initializing mcp-server-chart (Node.js) ==="
cd mcp-server-chart
npm install
cd ..

echo "=== Initializing douyin-mcp (Python) ==="
cd douyin-mcp
uv sync
cd ..

echo "=== Initializing jules-mcp-server (Python) ==="
cd jules-mcp-server
uv sync
cd ..

echo "=== Initializing PyMCPAutoGUI (Python) ==="
cd PyMCPAutoGUI
uv sync
cd ..

echo "=== Initializing mcp-proxy (Go) ==="
cd mcp-proxy
go mod download
cd ..

echo "=== Environment setup complete! ==="
