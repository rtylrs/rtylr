# ADR-004: Vite for All Apps (not Next.js)

**Status:** Accepted  
**Date:** 2026-01-15  
**Decision:** All 14 apps use Vite + React Router. No Next.js.

## Context

rtylr is a multi-app platform with 14 frontend apps. Each app is a separate SPA deployed independently. Options considered:

1. **Next.js** - SSR/SSG, file-based routing, API routes, image optimization
2. **Vite** - fast HMR, minimal config, pure SPA output
3. **Remix** - nested routing, server-first, progressive enhancement

## Decision

Vite 7 + React Router for all apps. No SSR. Pure client-side SPAs.

## Rationale

- **No SEO requirement**: These are authenticated business apps (POS, CRM, HR, Flow). Nobody googles a POS checkout screen. SSR provides zero value.
- **Deployment simplicity**: Each app builds to static files (`dist/`). Deploy to any CDN or serve from a single NGINX container. No Node.js server needed in production.
- **Build speed**: Vite's esbuild-based dev server starts in <500ms. HMR is instant. Next.js dev server is significantly slower with 14 apps.
- **No framework lock-in**: Vite is a build tool, not a framework. We can use any React library without worrying about Next.js compatibility (e.g., framer-motion works everywhere without `'use client'` directives).
- **POS requirement**: The POS app runs in Tauri (desktop) and needs to be a pure SPA. Next.js SSR is incompatible with Tauri's webview.

## Trade-offs

- **No API routes**: Backend is entirely Go. No BFF (Backend For Frontend) layer. Frontend calls Go API directly.
- **No SSR/SSG**: If we ever need a marketing site with SEO, it would need a separate Next.js app.
- **No ISR**: Dynamic content requires client-side fetching. No incremental static regeneration.

## Consequences

- `vite.config.ts` in every app. Shared config via `packages/vite-config`.
- React Router v7 for client-side routing in each app.
- All apps output to `dist/` as static files.
- Docker deployment: single NGINX container serving all SPAs with path-based routing.
