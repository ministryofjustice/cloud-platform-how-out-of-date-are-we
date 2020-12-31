#!/bin/sh

set -euo pipefail

curl -H "X-API-KEY: ${HOODAW_API_KEY}" -d "$(bin/documentation-pages-to-review.rb)" ${HOODAW_HOST}/documentation
