#!/bin/sh

set -euo pipefail

curl -H "X-API-KEY: ${HOODAW_API_KEY}" --header 'Content-Type: application/json' -d "$(/app/bin/ingress_weighting.rb)" ${HOODAW_HOST}/ingress_weighting
