import React from 'react';
import { Routes, Route } from 'react-router-dom';
import DocsLayout from './layouts/DocsLayout/DocsLayout';
import Home from './screens/Home';
import Article from './screens/Article';
import ApiReference from './screens/ApiReference';
import Changelog from './screens/Changelog';
import SearchResults from './screens/SearchResults';

const App: React.FC = () => {
  return (
    <Routes>
      <Route element={<DocsLayout />}>
        <Route path="/" element={<Home />} />
        <Route path="/docs/:slug" element={<Article />} />
        <Route path="/api" element={<ApiReference />} />
        <Route path="/changelog" element={<Changelog />} />
        <Route path="/search" element={<SearchResults />} />
      </Route>
    </Routes>
  );
};

export default App;
