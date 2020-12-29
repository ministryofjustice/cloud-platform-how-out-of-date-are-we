#!/bin/sh

set -euo pipefail

curl -H "X-API-KEY: ${HOODAW_API_KEY}" --header 'Content-Type: application/json' -d "$(/app/bin/hosted_services.rb)" ${HOODAW_HOST}/hosted_services