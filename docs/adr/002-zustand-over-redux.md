# ADR-002: Zustand for Client State (not Redux)

**Status:** Accepted  
**Date:** 2026-02-20  
**Decision:** Use Zustand for client-side state management. Server state uses TanStack Query.

## Context

With 14 frontend apps, we need a state management solution that's lightweight, TypeScript-native, and doesn't impose boilerplate. Options considered:

1. **Redux Toolkit** - industry standard, large ecosystem, significant boilerplate
2. **Zustand** - minimal API, TypeScript-first, no providers/context
3. **Jotai** - atomic state model, good for fine-grained reactivity
4. **React Context** - built-in, no dependencies, but re-render issues at scale

## Decision

Zustand for client/UI state (cart, theme, selected workspace, UI toggles). TanStack Query for all server state (API data fetching, caching, mutations).

## Rationale

- **Zero boilerplate**: A Zustand store is a single function call. No actions, reducers, selectors, providers, or middleware setup.
- **No context providers**: Stores are module-level singletons. No `<Provider>` wrapping needed, which matters when you have 14 apps.
- **TypeScript inference**: Full type inference without manual typing of actions/selectors. `useFlowStore((s) => s.tasks)` is fully typed.
- **Bundle size**: ~1.2KB gzipped vs Redux Toolkit's ~11KB. Matters when apps load in POS terminals.
- **Server state separation**: TanStack Query handles fetching, caching, background refetching, optimistic updates. Zustand only holds things React Query can't - UI preferences, cart state, selected IDs.

## Trade-offs

- **DevTools**: Redux DevTools are more mature. Zustand has a devtools middleware but it's less feature-rich.
- **Middleware ecosystem**: Redux has a vast middleware ecosystem (saga, thunk, persist). Zustand middleware is simpler but covers core needs (persist, devtools, immer).
- **Team familiarity**: Redux is more commonly known. New hires may need to learn Zustand patterns.

## Consequences

- Each app has 1-3 Zustand stores (e.g., `flowStore`, `themeStore`, `userStore`)
- API data flows through TanStack Query hooks, NOT through Zustand
- Zustand stores are for: UI state, shopping cart, selected items, preferences
- The `flowStore` is being migrated: reads move to TanStack Query, writes move to useMutation hooks
