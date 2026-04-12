import React, { useState } from 'react';
import { motion } from 'motion/react';
import { Link } from 'react-router-dom';
import { BookOpen, Search, Plus, Folder } from 'lucide-react';
import { useArticles, useArticleCategories } from './useKnowledgeBase';

const KnowledgeBase: React.FC = () => {
  const [search, setSearch] = useState('');
  const [categoryFilter, setCategoryFilter] = useState<string>('');
  const { data, isLoading, isError } = useArticles({
    search: search || undefined,
    category: categoryFilter || undefined,
  });
  const { data: categories } = useArticleCategories();

  const articles = data ?? [];
  const categoryList = categories ?? [];

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <div>
          <h2 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>
            Knowledge Base
          </h2>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-muted)' }}>
            Help articles organized by category
          </p>
        </div>
        <Link
          to="/knowledge/new"
          className="flex items-center gap-2 h-8 px-4 rounded-lg text-xs font-medium text-white no-underline transition-colors"
          style={{ background: '#4583C4' }}
        >
          <Plus size={14} />
          New article
        </Link>
      </div>

      <div className="flex items-center gap-3 mb-4">
        <div
          className="flex items-center gap-2 flex-1 h-9 px-3 rounded-lg text-sm"
          style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)' }}
        >
          <Search size={14} style={{ color: 'var(--hub-text-dim)' }} />
          <input
            type="text"
            placeholder="Search articles..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="flex-1 bg-transparent outline-none text-sm"
            style={{ color: 'var(--hub-text)' }}
          />
        </div>
        <select
          value={categoryFilter}
          onChange={(e) => setCategoryFilter(e.target.value)}
          className="h-9 px-3 rounded-lg text-sm outline-none"
          style={{
            background: 'var(--hub-input-bg)',
            border: '1px solid var(--hub-border)',
            color: 'var(--hub-text)',
          }}
        >
          <option value="">All categories</option>
          {categoryList.map((cat) => (
            <option key={cat.id || cat.name} value={cat.name}>
              {cat.name}
            </option>
          ))}
        </select>
      </div>

      {isLoading && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {[...Array(6)].map((_, i) => (
            <div
              key={i}
              className="h-32 rounded-xl animate-pulse"
              style={{ background: 'var(--hub-hover)' }}
            />
          ))}
        </div>
      )}

      {isError && (
        <div
          className="flex items-center justify-center py-16 rounded-xl text-sm"
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            color: 'var(--hub-text-muted)',
          }}
        >
          Failed to load articles. Please try again.
        </div>
      )}

      {!isLoading && !isError && articles.length === 0 && (
        <div
          className="flex flex-col items-center justify-center py-16 rounded-xl text-sm"
          style={{
            background: 'var(--hub-card)',
            border: '1px solid var(--hub-border)',
            color: 'var(--hub-text-muted)',
          }}
        >
          <BookOpen size={32} className="mb-3" style={{ color: 'var(--hub-text-dim)' }} />
          <p>No articles found</p>
          <Link to="/knowledge/new" className="mt-2 text-xs" style={{ color: '#4583C4' }}>
            Create your first article
          </Link>
        </div>
      )}

      {!isLoading && !isError && articles.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
          {articles.map((article) => (
            <Link key={article.id} to={`/knowledge/${article.id}/edit`} className="no-underline">
              <motion.div
                initial={{ opacity: 0, y: 4 }}
                animate={{ opacity: 1, y: 0 }}
                className="rounded-xl p-4 transition-colors cursor-pointer"
                style={{ background: 'var(--hub-card)', border: '1px solid var(--hub-border)' }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'var(--hub-hover)';
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.background = 'var(--hub-card)';
                }}
              >
                <div className="flex items-center gap-2 mb-2">
                  <Folder size={12} style={{ color: '#4583C4' }} />
                  <span className="text-[10px] font-medium" style={{ color: '#4583C4' }}>
                    {article.category}
                  </span>
                </div>
                <h3 className="text-sm font-medium mb-1" style={{ color: 'var(--hub-text)' }}>
                  {article.title}
                </h3>
                <p className="text-xs line-clamp-2" style={{ color: 'var(--hub-text-muted)' }}>
                  {article.body?.slice(0, 120)}
                </p>
                <p className="text-[10px] mt-2" style={{ color: 'var(--hub-text-dim)' }}>
                  {article.updatedAt}
                </p>
              </motion.div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
};

export default KnowledgeBase;
