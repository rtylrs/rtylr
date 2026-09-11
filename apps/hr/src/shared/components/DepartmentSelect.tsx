import React, { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { ChevronDown, Search } from 'lucide-react';
import { useDepartments } from '../../features/departments/useDepartments';

interface Props {
  value?: string;
  onChange: (name: string) => void;
  placeholder?: string;
}

const DepartmentSelect: React.FC<Props> = ({ value, onChange, placeholder }) => {
  const { t } = useTranslation();
  const { data } = useDepartments();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');

  const options = useMemo(() => {
    const list = data ?? [];
    if (!query) return list;
    return list.filter((d: { name: string }) => d.name.toLowerCase().includes(query.toLowerCase()));
  }, [data, query]);

  return (
    <div style={{ position: 'relative' }}>
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        style={{
          width: '100%',
          padding: '8px 10px',
          borderRadius: 8,
          border: '1px solid var(--hr-border)',
          background: 'var(--hr-input-bg)',
          color: value ? 'var(--hr-text)' : 'var(--hr-text-muted)',
          fontSize: 13,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          cursor: 'pointer',
        }}
      >
        {value || placeholder || t('common.select')}
        <ChevronDown size={14} />
      </button>
      {open && (
        <div
          style={{
            position: 'absolute',
            top: 'calc(100% + 4px)',
            left: 0,
            right: 0,
            background: 'var(--hr-card)',
            border: '1px solid var(--hr-border)',
            borderRadius: 10,
            padding: 8,
            zIndex: 40,
            boxShadow: 'var(--hr-shadow)',
            maxHeight: 240,
            overflowY: 'auto',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '6px 8px',
              border: '1px solid var(--hr-border)',
              borderRadius: 8,
              marginBottom: 6,
            }}
          >
            <Search size={13} style={{ color: 'var(--hr-text-muted)' }} />
            <input
              autoFocus
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder={t('common.search')}
              style={{
                border: 'none',
                outline: 'none',
                background: 'transparent',
                color: 'var(--hr-text)',
                fontSize: 12,
                flex: 1,
              }}
            />
          </div>
          {options.length === 0 && (
            <div
              style={{
                padding: 10,
                textAlign: 'center',
                color: 'var(--hr-text-muted)',
                fontSize: 12,
              }}
            >
              {t('common.noMatches')}
            </div>
          )}
          {options.map((d: { id: number; name: string }) => (
            <button
              type="button"
              key={d.id}
              onClick={() => {
                onChange(d.name);
                setOpen(false);
                setQuery('');
              }}
              style={{
                display: 'block',
                width: '100%',
                textAlign: 'left',
                padding: '7px 8px',
                borderRadius: 6,
                background: value === d.name ? 'var(--hr-active)' : 'transparent',
                color: 'var(--hr-text)',
                border: 'none',
                fontSize: 12,
                cursor: 'pointer',
              }}
              onMouseEnter={(e) =>
                ((e.currentTarget as HTMLButtonElement).style.background = 'var(--hr-hover)')
              }
              onMouseLeave={(e) =>
                ((e.currentTarget as HTMLButtonElement).style.background =
                  value === d.name ? 'var(--hr-active)' : 'transparent')
              }
            >
              {d.name}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

export default DepartmentSelect;
