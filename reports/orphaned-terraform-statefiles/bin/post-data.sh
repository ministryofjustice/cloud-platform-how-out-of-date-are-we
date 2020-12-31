#!/bin/sh

set -euo pipefail

bin/list-statefiles.rb > data.json

curl -vvv \
  -H "Content-Type: application/json" \
  -H "X-API-KEY: ${HOODAW_API_KEY}" \
  -d @data.json \
  ${HOODAW_HOST}/orphaned_statefiles
