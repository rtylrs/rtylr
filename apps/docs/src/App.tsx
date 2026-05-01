import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { ErrorBoundary } from '@voxire/ui';
import DocsLayout from './layouts/DocsLayout';
import Home from './features/home/Home';
import Article from './features/article/Article';
import ApiReference from './features/api-reference/ApiReference';
import Changelog from './features/changelog/Changelog';
import SearchResults from './features/search/SearchResults';

const App: React.FC = () => {
  return (
    <ErrorBoundary>
      <Routes>
        <Route element={<DocsLayout />}>
          <Route path="/" element={<Home />} />
          <Route path="/docs/:slug" element={<Article />} />
          <Route path="/api" element={<ApiReference />} />
          <Route path="/changelog" element={<Changelog />} />
          <Route path="/search" element={<SearchResults />} />
        </Route>
      </Routes>
    </ErrorBoundary>
  );
};

export default App;
