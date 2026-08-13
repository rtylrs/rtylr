import React from 'react';
import ReactMarkdown, { type Components, type UrlTransform } from 'react-markdown';
import remarkGfm from 'remark-gfm';
import '../../styles/markdown.css';

export interface MarkdownProps {
  source: string;
}

function sanitizeMarkdownUrl(url: string): string {
  const trimmed = url.trim();
  if (/^https?:\/\//i.test(trimmed) || trimmed.startsWith('/')) return trimmed;
  return '';
}

const safeUrlTransform: UrlTransform = (url) => sanitizeMarkdownUrl(url);

function tableHasBodyRows(children: React.ReactNode): boolean {
  return React.Children.toArray(children).some((section) => {
    if (!React.isValidElement<{ children?: React.ReactNode }>(section)) return false;
    if (section.type !== 'tbody') return false;
    return React.Children.toArray(section.props.children).some((row) => React.isValidElement(row));
  });
}

const markdownComponents: Components = {
  h1: ({ node: _node, ...props }) => <h1 className="caret-h" {...props} />,
  h2: ({ node: _node, ...props }) => <h2 className="caret-h" {...props} />,
  h3: ({ node: _node, ...props }) => <h3 className="caret-h" {...props} />,
  h4: ({ node: _node, ...props }) => <h4 className="caret-h" {...props} />,
  h5: ({ node: _node, ...props }) => <h5 className="caret-h" {...props} />,
  h6: ({ node: _node, ...props }) => <h6 className="caret-h" {...props} />,
  p: ({ node: _node, ...props }) => <p className="caret-p" {...props} />,
  ul: ({ node: _node, ...props }) => <ul className="caret-list" {...props} />,
  ol: ({ node: _node, ...props }) => <ol className="caret-list" {...props} />,
  blockquote: ({ node: _node, ...props }) => <blockquote className="caret-quote" {...props} />,
  hr: ({ node: _node, ...props }) => <hr className="caret-hr" {...props} />,
  code: ({ node: _node, className, ...props }) => (
    <code className={['caret-inline-code', className].filter(Boolean).join(' ')} {...props} />
  ),
  pre: ({ node: _node, ...props }) => <pre className="caret-codeblock" {...props} />,
  table: ({ node: _node, children, ...props }) =>
    tableHasBodyRows(children) ? (
      <div className="caret-table-wrap">
        <table className="caret-table" {...props}>
          {children}
        </table>
      </div>
    ) : (
      <p className="caret-table-empty">No matching records.</p>
    ),
  a: ({ node: _node, href, ...props }) => {
    const safeHref = href ? sanitizeMarkdownUrl(href) : '';
    if (!safeHref) return <span>{props.children}</span>;

    const isExternal = /^https?:\/\//i.test(safeHref);
    return (
      <a
        {...props}
        href={safeHref}
        target={isExternal ? '_blank' : undefined}
        rel={isExternal ? 'noopener noreferrer' : undefined}
        className="caret-link"
      />
    );
  },
};

/** Render model Markdown with GitHub-flavored tables, lists, and task items. */
export const Markdown: React.FC<MarkdownProps> = ({ source }) => {
  return (
    <div className="caret-md">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        components={markdownComponents}
        urlTransform={safeUrlTransform}
      >
        {source}
      </ReactMarkdown>
    </div>
  );
};

export default Markdown;
