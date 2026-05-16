import React from 'react';
import { Search, X } from 'lucide-react';

export interface SearchBarProps {
  value: string;
  onChange: (v: string) => void;
}

/** Search input used by the Caret history page. */
export const SearchBar: React.FC<SearchBarProps> = ({ value, onChange }) => (
  <div
    style={{
      display: 'flex',
      alignItems: 'center',
      gap: 8,
      padding: '8px 10px',
      background: 'var(--dash-surface)',
      borderRadius: 8,
      border: '1px solid var(--dash-border)',
    }}
  >
    <Search size={13} style={{ color: 'var(--dash-text-dim)' }} />
    <input
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder="Search chats…"
      aria-label="Search chats"
      style={{
        flex: 1,
        background: 'transparent',
        border: 'none',
        outline: 'none',
        color: 'var(--dash-text)',
        fontSize: 13,
      }}
    />
    {value && (
      <button
        type="button"
        aria-label="Clear search"
        onClick={() => onChange('')}
        style={{
          background: 'transparent',
          border: 'none',
          cursor: 'pointer',
          padding: 0,
          color: 'var(--dash-text-dim)',
          display: 'flex',
        }}
      >
        <X size={12} />
      </button>
    )}
  </div>
);

export default SearchBar;
