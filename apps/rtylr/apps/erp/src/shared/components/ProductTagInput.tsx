import React, { useState } from 'react';
import { X, Tag } from 'lucide-react';

interface ProductTagInputProps {
  value: string[];
  onChange: (tags: string[]) => void;
  placeholder?: string;
}

const ProductTagInput: React.FC<ProductTagInputProps> = ({
  value,
  onChange,
  placeholder = 'Add tag…',
}) => {
  const [input, setInput] = useState('');

  const add = () => {
    const t = input.trim();
    if (!t || value.includes(t)) return;
    onChange([...value, t]);
    setInput('');
  };

  const remove = (t: string) => {
    onChange(value.filter((x) => x !== t));
  };

  return (
    <div
      className="flex items-center gap-1.5 flex-wrap rounded-lg px-2 py-1.5"
      style={{ background: 'var(--hub-input-bg)', border: '1px solid var(--hub-border)' }}
    >
      <Tag size={12} style={{ color: 'var(--hub-text-dim)' }} />
      {value.map((t) => (
        <span
          key={t}
          className="flex items-center gap-1 text-[11px] px-2 py-0.5 rounded-full"
          style={{
            background: 'color-mix(in srgb, var(--erp-accent) 12%, transparent)',
            color: 'var(--erp-accent)',
          }}
        >
          {t}
          <button type="button" onClick={() => remove(t)}>
            <X size={9} />
          </button>
        </span>
      ))}
      <input
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ',') {
            e.preventDefault();
            add();
          } else if (e.key === 'Backspace' && !input && value.length > 0) {
            onChange(value.slice(0, -1));
          }
        }}
        placeholder={value.length === 0 ? placeholder : ''}
        className="flex-1 min-w-[60px] bg-transparent outline-none text-xs"
        style={{ color: 'var(--hub-text)' }}
      />
    </div>
  );
};

export default ProductTagInput;
