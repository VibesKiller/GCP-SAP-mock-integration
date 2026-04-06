#!/usr/bin/env bash
set -euo pipefail

find . -maxdepth 3 \
  -not -path './.git*' \
  | sort
