import React, { Suspense } from 'react';
import { createRootRoute, createRoute, createRouter, Outlet } from '@tanstack/react-router';
import { ErrorBoundary } from '@voxire/ui';
import DocsLayout from './layouts/DocsLayout';
import Home from './features/home/Home';
import Article from './features/article/Article';
import ApiReference from './features/api-reference/ApiReference';
import Changelog from './features/changelog/Changelog';
import SearchResults from './features/search/SearchResults';

const rootRoute = createRootRoute({
  component: () => (
    <ErrorBoundary>
      <Outlet />
    </ErrorBoundary>
  ),
});

const layoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'docs-layout',
  component: DocsLayout,
});

const indexRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/',
  component: () => <Home />,
});

const articleRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/docs/$slug',
  component: () => <Article />,
});

const apiRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/api',
  component: () => <ApiReference />,
});

const changelogRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/changelog',
  component: () => <Changelog />,
});

const searchRoute = createRoute({
  getParentRoute: () => layoutRoute,
  path: '/search',
  component: () => <SearchResults />,
});

const routeTree = rootRoute.addChildren([
  layoutRoute.addChildren([indexRoute, articleRoute, apiRoute, changelogRoute, searchRoute]),
]);

export const router = createRouter({ routeTree });

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router;
  }
}
