import React, { useMemo } from 'react';
import { useRouterState, Link } from '@tanstack/react-router';
import { FileText } from 'lucide-react';
import { content } from '../../shared/content';
import { HelmetContainer } from '@voxire/components';

interface SearchResult {
  slug: string;
  title: string;
  snippet: string;
}

const SearchResults: React.FC = () => {
  const searchStr = useRouterState({ select: (s) => s.location.searchStr });
  const query = new URLSearchParams(searchStr).get('q') || '';

  const results = useMemo<SearchResult[]>(() => {
    if (!query.trim()) return [];

    const q = query.toLowerCase();
    const matches: SearchResult[] = [];

    for (const [slug, doc] of Object.entries(content)) {
      const titleMatch = doc.title.toLowerCase().includes(q);
      const bodyText = doc.body.replace(/<[^>]+>/g, '');
      const bodyMatch = bodyText.toLowerCase().includes(q);

      if (titleMatch || bodyMatch) {
        let snippet = '';
        if (bodyMatch) {
          const idx = bodyText.toLowerCase().indexOf(q);
          const start = Math.max(0, idx - 60);
          const end = Math.min(bodyText.length, idx + query.length + 60);
          snippet =
            (start > 0 ? '...' : '') +
            bodyText.slice(start, end).trim() +
            (end < bodyText.length ? '...' : '');
        } else {
          snippet = bodyText.slice(0, 120).trim() + '...';
        }

        matches.push({ slug, title: doc.title, snippet });
      }
    }

    return matches;
  }, [query]);

  return (
    <>
      <HelmetContainer title="Search results · rtylr" />
      <div>
        <h1 className="text-2xl font-bold mb-2" style={{ color: 'var(--docs-text)' }}>
          Search Results
        </h1>
        {query && (
          <p className="text-sm mb-6" style={{ color: 'var(--docs-text-muted)' }}>
            {results.length} {results.length === 1 ? 'result' : 'results'} for "{query}"
          </p>
        )}

        {!query && (
          <p className="text-sm" style={{ color: 'var(--docs-text-muted)' }}>
            Enter a search term to find documentation.
          </p>
        )}

        {query && results.length === 0 && (
          <div className="py-10 text-center">
            <p className="text-sm" style={{ color: 'var(--docs-text-muted)' }}>
              No results found. Try a different search term.
            </p>
          </div>
        )}

        <div className="flex flex-col gap-2">
          {results.map((result) => (
            <Link
              key={result.slug}
              to="/docs/$slug"
              params={{ slug: result.slug }}
              className="flex items-start gap-3 p-4 rounded-lg border no-underline transition-colors"
              style={{ borderColor: 'var(--docs-border)', background: 'var(--docs-bg)' }}
            >
              <FileText
                size={16}
                className="mt-0.5 flex-shrink-0"
                style={{ color: 'var(--docs-accent)' }}
              />
              <div>
                <span className="text-sm font-medium" style={{ color: 'var(--docs-text)' }}>
                  {result.title}
                </span>
                <p className="text-sm mt-1" style={{ color: 'var(--docs-text-muted)' }}>
                  {result.snippet}
                </p>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </>
  );
};

export default SearchResults;
