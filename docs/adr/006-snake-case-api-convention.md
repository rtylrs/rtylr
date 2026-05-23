# ADR-006: snake_case API with Automatic camelCase Conversion

**Status:** Accepted  
**Date:** 2026-03-25  
**Decision:** All API JSON uses snake_case. The Axios client automatically converts between camelCase (frontend) and snake_case (backend).

## Context

Go uses PascalCase for exported fields with `json:"snake_case"` tags. TypeScript/JavaScript conventions use camelCase. This creates a mismatch. Options:

1. **camelCase everywhere** - Go JSON tags use camelCase (non-standard for Go APIs)
2. **snake_case everywhere** - TypeScript code uses snake_case (non-standard for JS)
3. **Automatic conversion** - Backend uses snake_case, frontend uses camelCase, interceptor converts

## Decision

Option 3. All Go struct tags are `json:"snake_case"`. The Axios request interceptor converts camelCase request bodies to snake_case. The response interceptor converts snake_case response bodies to camelCase.

```typescript
// Request interceptor
function camelToSnake(str: string): string {
  return str.replace(/[A-Z]/g, (c) => '_' + c.toLowerCase());
}

// Response interceptor
function snakeToCamel(str: string): string {
  return str.replace(/_([a-z])/g, (_, c) => c.toUpperCase());
}
```

## Rationale

- **Go conventions**: Go APIs universally use snake_case JSON. Tools like `golint` flag camelCase JSON tags.
- **JS conventions**: TypeScript code reads naturally with camelCase. `task.dueDate` not `task.due_date`.
- **Single conversion point**: The interceptor is one place. Every API call goes through it. No per-endpoint conversion logic.
- **Proto compatibility**: Protobuf field names are snake_case by convention. Generated Go code uses snake_case JSON tags. This aligns.

## Trade-offs

- **Debugging confusion**: Network tab shows snake_case, TypeScript types show camelCase. Must mentally translate.
- **Interceptor complexity**: The recursive `convertKeys()` function handles nested objects, arrays, and edge cases. A bug here affects every API call.
- **Performance**: Deep object conversion on every request/response adds CPU time. Negligible for typical payloads but worth noting.

## Consequences

- 2,411 Go JSON tags across 246 files were standardized to snake_case
- `packages/api/src/client.ts` has the conversion interceptors
- Frontend TypeScript types use camelCase field names
- Proto-generated types are already snake_case, matching the Go convention
