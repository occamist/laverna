#!/usr/bin/env python3
"""Creates a .ankiaddon package for distribution on AnkiWeb"""

import shutil
import zipfile
from pathlib import Path

def main():
    script_dir = Path(__file__).parent.resolve()
    project_root = script_dir.parent.resolve()
    addon_dir = project_root / "addon"

    print(f"Working directory: {addon_dir}")

    # Clean cache directories and files
    for pattern in ["__pycache__", ".mypy_cache", ".ruff_cache"]:
        for path in addon_dir.rglob(pattern):
            if path.is_dir():
                shutil.rmtree(path, ignore_errors=True)

    for pyc in addon_dir.rglob("*.pyc"):
        pyc.unlink(missing_ok=True)

    (addon_dir / "addon.log").unlink(missing_ok=True)

    output_path = project_root / "laverna-addon.ankiaddon"
    files_to_package = ["__init__.py", "config.json", "README.md"]

    with zipfile.ZipFile(output_path, "w", zipfile.ZIP_DEFLATED) as zf:
        for filename in files_to_package:
            src = addon_dir / filename
            if src.exists():
                zf.write(src, arcname=filename)
            else:
                print(f"Warning: {filename} not found, skipping")

    print(f"Done! Package created at: {output_path}")

if __name__ == "__main__":
    main()