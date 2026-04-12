# ADR-001: Proto Contracts via Git Submodule (not npm package)

**Status:** Accepted  
**Date:** 2026-03-15  
**Decision:** Use `voxie` as a git submodule shared between frontend and backend, not as a published npm/Go package.

## Context

We have a Go backend (`onion`) and 14 React frontend apps that need to share type definitions for API request/response shapes. Options considered:

1. **npm package** — publish generated TypeScript types to a private npm registry
2. **Go module** — publish generated `.pb.go` files as a Go module
3. **Git submodule** — single repo (`voxie`) containing `.proto` files + generated code for both languages, included as a submodule in both repos

## Decision

Git submodule. The `voxie` repo contains:

- `proto/{app}/v1/{concern}.proto` — source of truth
- `gen/go/{app}/v1/*.pb.go` — generated Go types (committed)
- `gen/ts/{app}/v1/*.ts` — generated TypeScript types (committed)
- `buf.gen.yaml` — buf generation config

Both `rtylr` (frontend) and `onion` (backend) include `voxie` as a submodule. Frontend re-exports via `@voxire/contracts` package.

## Rationale

- **Atomic updates**: A single commit in `voxie` updates both frontend and backend types simultaneously. No version drift between a Go module tag and an npm package version.
- **No registry needed**: No private npm registry or Go module proxy to maintain. Reduces infrastructure.
- **CI simplicity**: `git submodule update --init --recursive` in CI gives both repos the exact same generated code.
- **Protobuf as source of truth**: `.proto` files are language-agnostic. `buf generate` produces both Go and TypeScript from the same definition. Adding a third language (e.g., Swift for mobile) is just another `buf.gen.yaml` entry.

## Trade-offs

- **Submodule pain**: Developers must remember `git submodule update`. Force-pushes to `voxie` can break CI if the submodule ref points to a non-existent commit.
- **CI tokens**: HTTPS submodule checkout requires a PAT_TOKEN in CI (SSH URLs don't work in GitHub Actions HTTPS context).
- **Generated code in git**: Committing generated `.pb.go` and `.ts` files is unconventional but avoids requiring `protoc`/`buf` on every developer's machine.

## Consequences

- Every handler in `onion` imports from `github.com/voxire/voxie/gen/go/{app}/v1`
- Every frontend hook imports from `@voxire/contracts/{app}/{concern}`
- Adding a new API endpoint requires: 1) add to `.proto`, 2) `buf generate`, 3) commit to `voxie`, 4) update submodule ref in both repos
