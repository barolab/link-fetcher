#!/bin/sh

echo "- First attempt with grep:"
cat sample.txt | grep -Eo "\w+\.com" | tr '[:upper:]' '[:lower:]' | sort | uniq

echo ""
echo "- Second attempt with awk:"
cat sample.txt | awk  -v FPAT='\\w+\\.com' '{print $1}' | awk '{print tolower($0)}' | sort | uniq
