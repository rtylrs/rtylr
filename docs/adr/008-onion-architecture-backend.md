# ADR-008: Onion Architecture for Go Backend

**Status:** Accepted  
**Date:** 2026-01-20  
**Decision:** Each backend module follows onion/clean architecture with strict layer dependencies.

## Context

With 100+ modules, the backend needs a consistent structure that prevents coupling. Options:

1. **Flat handlers** — routes directly call database
2. **MVC** — models, views, controllers
3. **Onion/Clean Architecture** — domain, application, infrastructure layers with inward dependencies

## Decision

Each module follows this structure:

```
internal/modules/{name}/
  internal/
    domain/          # Entities, value objects, interfaces
    application/     # Service layer (business logic)
    repository/      # Database access (sqlc queries)
    routes/
      handlers/      # HTTP handlers (request/response only)
      routes.go      # Route registration
  module.go          # Module entry point, dependency wiring
```

Dependencies flow inward: handlers -> application -> repository. Domain has no external dependencies.

## Rationale

- **Testability**: Services can be tested with mock repositories. Handlers can be tested with mock services. Each layer is independently testable.
- **Business logic isolation**: The application layer contains ALL business rules. Handlers only parse HTTP requests and format responses. Repositories only execute SQL.
- **Enforced by convention**: "Handler should not talk to a DB. That should happen only on repo." — this rule prevents data access logic from leaking into HTTP handlers.
- **Module independence**: Each module is self-contained. Module A doesn't import Module B's internals. Cross-module communication goes through service interfaces.

## Trade-offs

- **Boilerplate**: A simple CRUD endpoint requires files in handlers/, application/, repository/, and domain/. More code than a flat handler.
- **Indirection**: A request passes through 3+ layers before reaching the database. More stack frames to debug.
- **Learning curve**: New developers must understand which code goes in which layer.

## Consequences

- Raw SQL is NEVER in handlers (enforced by code review)
- Business validation is in the Service layer, not in handlers
- Domain types are separate from database models (sqlc types) and API types (proto types)
- Handlers convert between proto request types and domain types
