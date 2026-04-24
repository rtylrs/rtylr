#!/usr/bin/env bash
set -euo pipefail

bucket="${LOCAL_STACK_BUCKET:-onion-dev}"

awslocal s3 mb "s3://${bucket}" || true
