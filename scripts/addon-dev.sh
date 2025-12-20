#!/usr/bin/env bash

# Change to project root directory
cd "$(dirname "$0")/.."

plugin_name=LavernaAddonDev
plugin_path_linux=~/.local/share/Anki2/addons21
plugin_path_mac=~/Library/Application\ Support/Anki2/addons21

if [ -d "$plugin_path_linux" ]; then
    ln -s -f "$(pwd)/addon" "$plugin_path_linux/$plugin_name"
fi

if [ -d "$plugin_path_mac" ]; then
    ln -s -f "$(pwd)/addon" "$plugin_path_mac/$plugin_name"
fi
