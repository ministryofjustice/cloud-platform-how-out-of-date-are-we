#!/bin/sh

set -euo pipefail

curl -H "X-API-KEY: ${HOODAW_API_KEY}" -d "$(./bin/list_orphaned_resources.rb)" ${HOODAW_HOST}/orphaned_resources
