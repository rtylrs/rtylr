import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { ChevronLeft } from 'lucide-react';
import { content } from '../content';

const Article: React.FC = () => {
  const { slug } = useParams<{ slug: string }>();
  const article = slug ? content[slug] : undefined;

  if (!article) {
    return (
      <div>
        <Link to="/" className="inline-flex items-center gap-1 text-sm mb-6 no-underline"
          style={{ color: 'var(--docs-accent)' }}>
          <ChevronLeft size={14} />
          Back to docs
        </Link>
        <h1 className="text-2xl font-bold mb-3" style={{ color: 'var(--docs-text)' }}>
          Page not found
        </h1>
        <p className="text-sm" style={{ color: 'var(--docs-text-muted)' }}>
          This documentation page does not exist yet. Check the sidebar for available pages.
        </p>
      </div>
    );
  }

  return (
    <div>
      <Link to="/" className="inline-flex items-center gap-1 text-sm mb-6 no-underline"
        style={{ color: 'var(--docs-accent)' }}>
        <ChevronLeft size={14} />
        Back to docs
      </Link>
      <article className="docs-prose">
        <h1>{article.title}</h1>
        <div dangerouslySetInnerHTML={{ __html: article.body }} />
      </article>
    </div>
  );
};

export default Article;
