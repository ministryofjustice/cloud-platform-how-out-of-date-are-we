#!/bin/sh

set -euo pipefail

git clone --depth 1 https://github.com/ministryofjustice/cloud-platform-environments.git

curl -H "X-API-KEY: ${HOODAW_API_KEY}" -d "$(bin/module-versions.rb)" ${HOODAW_HOST}/terraform_modules
