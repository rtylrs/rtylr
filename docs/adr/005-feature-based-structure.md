# ADR-005: Feature-Based Directory Structure (not Flat)

**Status:** Accepted  
**Date:** 2026-03-20  
**Decision:** Organize each app by feature, not by file type.

## Context

With 14 apps, some having 15+ screens, the codebase was becoming hard to navigate. Options considered:

1. **Flat by type**: `screens/`, `hooks/`, `services/`, `components/` — all screens in one folder, all hooks in another
2. **Feature-based**: `features/{name}/Screen.tsx + useX.ts` — co-locate screen, hooks, and local components

## Decision

Feature-based. Each feature gets its own directory containing its screen component and hooks:

```
apps/flow/src/
  features/
    tasks/
      MyIssues.tsx      # Screen
      TaskDetail.tsx     # Screen
      useMyIssues.ts     # Hook
      useTasks.ts        # Hooks (CRUD mutations)
      useTask.ts         # Hook (single task)
    projects/
      ProjectList.tsx
      ProjectBoard.tsx
      ProjectSettings.tsx
      useProjects.ts
    calendar/
      Calendar.tsx
  shared/
    components/          # Cross-feature components
    services/            # API service modules
    stores/              # Zustand stores
    lib/                 # Utilities
    hooks/               # Cross-feature hooks
```

## Rationale

- **Colocation**: When working on the task feature, everything you need is in `features/tasks/`. No jumping between `screens/TaskDetail.tsx`, `hooks/useTasks.ts`, `services/taskService.ts`.
- **Deletion test**: To remove a feature, delete its folder. With flat structure, you'd hunt across 5 directories.
- **Scale**: 15 screens in one `screens/` folder gets overwhelming. Feature folders keep each concern small (2-5 files per feature).
- **Import clarity**: `import { useTasks } from '../tasks/useTasks'` immediately tells you the dependency relationship.

## Trade-offs

- **Shared code placement**: Some hooks/components are used by multiple features. These go in `shared/`. The boundary between "feature-specific" and "shared" requires judgment.
- **Circular deps**: Feature A importing from Feature B can create circular dependencies. Shared utilities should be lifted to `shared/`.

## Consequences

- 760 files were reorganized in a single migration
- Every app follows the same `features/` + `shared/` pattern
- New features start with `features/{name}/Screen.tsx` + `features/{name}/useX.ts`
- Cross-feature imports go through `shared/`
