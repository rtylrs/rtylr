import React, { useMemo, useState } from 'react';
import { ChevronDown, Search } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

export interface Country {
  code: string;
  name: string;
  dial: string;
  flag: string;
}

export const COUNTRIES: Country[] = [
  { code: 'AE', name: 'United Arab Emirates', dial: '+971', flag: '🇦🇪' },
  { code: 'SA', name: 'Saudi Arabia', dial: '+966', flag: '🇸🇦' },
  { code: 'US', name: 'United States', dial: '+1', flag: '🇺🇸' },
  { code: 'GB', name: 'United Kingdom', dial: '+44', flag: '🇬🇧' },
  { code: 'EG', name: 'Egypt', dial: '+20', flag: '🇪🇬' },
  { code: 'JO', name: 'Jordan', dial: '+962', flag: '🇯🇴' },
  { code: 'LB', name: 'Lebanon', dial: '+961', flag: '🇱🇧' },
  { code: 'KW', name: 'Kuwait', dial: '+965', flag: '🇰🇼' },
  { code: 'QA', name: 'Qatar', dial: '+974', flag: '🇶🇦' },
  { code: 'BH', name: 'Bahrain', dial: '+973', flag: '🇧🇭' },
  { code: 'OM', name: 'Oman', dial: '+968', flag: '🇴🇲' },
  { code: 'TR', name: 'Turkey', dial: '+90', flag: '🇹🇷' },
  { code: 'FR', name: 'France', dial: '+33', flag: '🇫🇷' },
  { code: 'DE', name: 'Germany', dial: '+49', flag: '🇩🇪' },
];

export interface CountrySelectProps {
  value: string;
  onChange: (country: Country) => void;
  label?: string;
}

const CountrySelect: React.FC<CountrySelectProps> = ({
  value,
  onChange,
  label = 'Country',
}) => {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const selected = COUNTRIES.find((c) => c.code === value) || COUNTRIES[0];

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return COUNTRIES;
    return COUNTRIES.filter(
      (c) =>
        c.name.toLowerCase().includes(q) ||
        c.code.toLowerCase().includes(q) ||
        c.dial.includes(q)
    );
  }, [query]);

  return (
    <div className="flex flex-col gap-2">
      <label className="text-[11px] font-semibold uppercase tracking-[0.08em] text-white/50">
        {label}
      </label>
      <div className="relative">
        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          className="w-full h-[52px] rounded-xl bg-white/[0.04] border border-white/10 hover:bg-white/[0.06] hover:border-white/15 focus:border-[#4583C4] focus:ring-2 focus:ring-[#4583C4]/25 outline-none px-4 flex items-center justify-between text-[15px] text-white transition-all"
        >
          <span className="inline-flex items-center gap-2.5">
            <span className="text-lg">{selected.flag}</span>
            <span className="text-white/80">{selected.name}</span>
            <span className="text-white/45">{selected.dial}</span>
          </span>
          <ChevronDown
            className={[
              'h-4 w-4 text-white/50 transition-transform',
              open ? 'rotate-180' : '',
            ].join(' ')}
          />
        </button>
        <AnimatePresence>
          {open && (
            <motion.div
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: 6 }}
              transition={{ duration: 0.15 }}
              className="absolute left-0 right-0 z-50 mt-2 rounded-xl border border-white/10 bg-[#0a0b2e]/95 backdrop-blur-md shadow-xl overflow-hidden"
            >
              <div className="p-2 border-b border-white/10">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-white/40" />
                  <input
                    autoFocus
                    type="text"
                    placeholder="Search country"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    className="w-full h-10 rounded-lg bg-white/[0.04] border border-white/10 pl-9 pr-3 text-sm text-white placeholder:text-white/30 outline-none focus:border-[#4583C4]"
                  />
                </div>
              </div>
              <div className="max-h-60 overflow-y-auto py-1">
                {filtered.map((c) => (
                  <button
                    key={c.code}
                    type="button"
                    onClick={() => {
                      onChange(c);
                      setOpen(false);
                      setQuery('');
                    }}
                    className="w-full flex items-center gap-2.5 px-3 py-2 text-sm text-white/85 hover:bg-white/[0.06] transition-colors"
                  >
                    <span className="text-lg">{c.flag}</span>
                    <span className="flex-1 text-left">{c.name}</span>
                    <span className="text-white/50">{c.dial}</span>
                  </button>
                ))}
                {filtered.length === 0 && (
                  <p className="px-3 py-4 text-center text-xs text-white/40">
                    No countries found
                  </p>
                )}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
};

export default CountrySelect;
