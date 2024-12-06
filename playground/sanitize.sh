#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

echo "- First attempt with grep:"
cat "$SCRIPT_DIR/sample.txt" | grep -Eo "\w+\.com" | tr '[:upper:]' '[:lower:]' | sort | uniq

echo ""
echo "- Second attempt with awk:"
cat "$SCRIPT_DIR/sample.txt" | awk  -v FPAT='\\w+\\.com' '{print $1}' | awk '{print tolower($0)}' | sort | uniq
