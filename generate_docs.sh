#!/bin/bash

# Check if LDoc is installed
if ! command -v ldoc &> /dev/null; then
    echo "LDoc is not installed. Please install it with one of the following commands:"
    echo "  luarocks install ldoc  # If you have LuaRocks installed"
    echo "  apt-get install lua-ldoc  # On Debian/Ubuntu"
    echo "  brew install lua-ldoc  # On macOS with Homebrew"
    echo ""
    echo "After installing LDoc, run this script again."
    exit 1
fi

# Generate documentation using LDoc
ldoc .

echo "Documentation generated in docs/index.html"
