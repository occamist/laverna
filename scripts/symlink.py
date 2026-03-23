#!/usr/bin/env python3
"""Creates a symlink in Anki's addons folder for the addon development and testing"""

from pathlib import Path

PLUGIN_NAME = "LavernaAddonDev"

def main():
    project_root = Path(__file__).parent.parent.resolve()
    addon_dir = project_root / "addon"

    plugin_paths = [
        Path.home() / ".local/share/Anki2/addons21",   # Linux
        Path.home() / "Library/Application Support/Anki2/addons21",  # macOS
    ]

    for plugin_path in plugin_paths:
        if plugin_path.is_dir():
            symlink = plugin_path / PLUGIN_NAME
            symlink.unlink(missing_ok=True)  # remove existing symlink/file first
            symlink.symlink_to(addon_dir)
            print(f"Symlink created: {symlink} → {addon_dir}")

if __name__ == "__main__":
    main()