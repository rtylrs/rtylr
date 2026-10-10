import React, { useState } from "react";
import { caretTokens } from "../ui/tokens";
import type { SearchSource } from "../lib/searchSources";

export interface SourceChipsProps {
  sources: SearchSource[];
  /** How many chips show before the "+N" toggle. */
  visible?: number;
}

/** Compact links to the pages a reply's web search drew on. */
export const SourceChips: React.FC<SourceChipsProps> = ({
  sources,
  visible = 4,
}) => {
  const [expanded, setExpanded] = useState(false);
  if (sources.length === 0) return null;

  const shown = expanded ? sources : sources.slice(0, visible);
  const hidden = sources.length - shown.length;

  return (
    <ul
      className="caret-source-chips"
      aria-label="Sources"
      style={{
        display: "flex",
        flexWrap: "wrap",
        gap: 6,
        listStyle: "none",
        margin: "12px 0 0",
        padding: 0,
      }}
    >
      {shown.map((source) => (
        <li key={source.url} style={{ minWidth: 0 }}>
          <a
            className="caret-source-chip"
            href={source.url}
            target="_blank"
            rel="noopener noreferrer"
            title={source.title}
            style={{
              display: "inline-flex",
              alignItems: "center",
              maxWidth: 200,
              padding: "4px 10px",
              borderRadius: 999,
              border: `1px solid ${caretTokens.border}`,
              background: caretTokens.surface,
              color: caretTokens.text,
              fontSize: 12,
              lineHeight: 1.4,
              textDecoration: "none",
            }}
          >
            <span
              style={{
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
            >
              {source.host}
            </span>
            <span className="caret-visually-hidden">: {source.title}</span>
          </a>
        </li>
      ))}
      {hidden > 0 && (
        <li>
          <button
            type="button"
            className="caret-source-chip"
            onClick={() => setExpanded(true)}
            style={{
              padding: "4px 10px",
              borderRadius: 999,
              border: `1px solid ${caretTokens.border}`,
              background: "transparent",
              color: caretTokens.textMuted,
              fontSize: 12,
              lineHeight: 1.4,
              cursor: "pointer",
            }}
          >
            +{hidden} more
          </button>
        </li>
      )}
    </ul>
  );
};

export default SourceChips;
