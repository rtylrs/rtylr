#!/usr/bin/env sh
set -eu
exec pnpm exec protoc-gen-ts_proto "$@"
