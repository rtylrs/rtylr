# ADR-003: neverthrow for Type-Safe API Error Handling

**Status:** Accepted  
**Date:** 2026-03-01  
**Decision:** Wrap all API calls in `neverthrow` `ResultAsync` via `safeFetch()`. Use `unwrap()` to bridge to React Query.

## Context

API calls can fail in many ways: network errors, 401s, 500s, validation errors. Without explicit error handling, these failures are silent or produce cryptic errors. Options considered:

1. **Try/catch everywhere** - verbose, easy to forget, error types are `unknown`
2. **Axios interceptors only** - handles auth redirects but doesn't give callers typed errors
3. **neverthrow Result type** - forces callers to handle both success and error paths

## Decision

`safeFetch(axiosClient)` returns a `SafeApi` object where every method (get/post/put/patch/delete) returns `ResultAsync<T, ApiError>`. The `ApiError` type has structured fields: `status`, `code` (enum), `message`.

```typescript
// In service layer
const result = await taskService.create(data);
// result is ResultAsync<ApiTask, ApiError>

// In React Query hook (via unwrap helper)
useMutation({ mutationFn: (data) => unwrap(taskService.create(data)) });
```

## Rationale

- **Compile-time safety**: You can't accidentally ignore an error. `ResultAsync` forces you to handle `.match()`, `.map()`, or `unwrap()`.
- **Structured errors**: `ApiError.code` is an enum (`NETWORK_ERROR`, `UNAUTHORIZED`, `VALIDATION_ERROR`, `SERVER_ERROR`). No more `if (err.response?.status === 401)` scattered everywhere.
- **Composable**: Results chain with `.andThen()`, `.map()`, `.orElse()`. Complex flows (fetch user, then fetch their projects) compose cleanly.
- **React Query bridge**: `unwrap()` converts `ResultAsync<T, ApiError>` to `Promise<T>` that throws on error - exactly what React Query's `queryFn`/`mutationFn` expects.

## Trade-offs

- **Learning curve**: Developers must understand Result types. Not standard JavaScript practice.
- **Bundle size**: neverthrow adds ~2KB gzipped.
- **Verbosity**: Simple calls need `unwrap()` wrapping. `await unwrap(api.get('/foo'))` vs `await axios.get('/foo')`.

## Consequences

- `packages/api/src/result.ts` exports `safeFetch`, `unwrap`, `ApiError`
- All service files use `safe.get/post/put/delete` (the safeFetch instance)
- React Query hooks use `unwrap()` in `queryFn`/`mutationFn`
- Error boundaries catch structured `ApiError` objects, not raw Axios errors
