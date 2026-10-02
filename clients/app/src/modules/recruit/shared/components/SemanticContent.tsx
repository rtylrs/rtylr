import React from 'react';
import { cleanMojibake } from '../lib/sanitize';

interface SemanticContentProps {
  content: string | null | undefined;
  className?: string;
  itemClassName?: string;
}

/**
 * Parses multi-line strings into semantic HTML paragraphs and unordered/ordered lists.
 * This guarantees accessible navigation and landmark hierarchy for screen readers.
 */
export const SemanticContent: React.FC<SemanticContentProps> = ({
  content,
  className,
  itemClassName,
}) => {
  const cleaned = cleanMojibake(content);
  if (!cleaned) return null;

  const lines = cleaned.split('\n');
  const blocks: React.ReactNode[] = [];
  let currentList: string[] = [];
  let isOrdered = false;

  const flushList = () => {
    if (currentList.length === 0) return;
    const ListTag = isOrdered ? 'ol' : 'ul';
    blocks.push(
      <ListTag
        key={`list-${blocks.length}`}
        className={
          isOrdered ? 'my-3 list-decimal space-y-1.5 pl-5' : 'my-3 list-disc space-y-1.5 pl-5'
        }
      >
        {currentList.map((item, idx) => (
          <li key={idx} className={itemClassName}>
            {item}
          </li>
        ))}
      </ListTag>
    );
    currentList = [];
    isOrdered = false;
  };

  for (let i = 0; i < lines.length; i++) {
    const rawLine = lines[i];
    const trimmed = rawLine.trim();

    if (!trimmed) {
      flushList();
      continue;
    }

    // Check for unordered bullet prefixes
    const bulletMatch = trimmed.match(/^[-*•]\s+(.*)$/);
    // Check for numbered bullet prefixes
    const numberMatch = trimmed.match(/^(\d+)[.)]\s+(.*)$/);

    if (bulletMatch) {
      if (isOrdered && currentList.length > 0) {
        flushList();
      }
      isOrdered = false;
      currentList.push(bulletMatch[1]);
    } else if (numberMatch) {
      if (!isOrdered && currentList.length > 0) {
        flushList();
      }
      isOrdered = true;
      currentList.push(numberMatch[2]);
    } else {
      flushList();
      blocks.push(
        <p key={`p-${blocks.length}`} className="my-2 leading-relaxed">
          {trimmed}
        </p>
      );
    }
  }

  flushList();

  return <div className={className}>{blocks}</div>;
};

export default SemanticContent;
