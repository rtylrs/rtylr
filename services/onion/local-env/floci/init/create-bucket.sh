#!/usr/bin/env bash
set -euo pipefail

bucket="${FLOCI_BUCKET:-onion-dev}"

aws s3 mb "s3://${bucket}" || true
