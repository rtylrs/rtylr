import React, { useMemo } from 'react';
import { renderMarkdown } from '../../markdown';
import '../../styles/markdown.css';

export interface MarkdownProps {
  source: string;
}

/** Render a markdown string to safe HTML. Memoized on the source text to
 *  avoid reparsing every streaming tick when content hasn't changed. */
export const Markdown: React.FC<MarkdownProps> = ({ source }) => {
  const html = useMemo(() => renderMarkdown(source), [source]);
  return <div className="caret-md" dangerouslySetInnerHTML={{ __html: html }} />;
};

export default Markdown;
