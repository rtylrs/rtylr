import React, { useEffect, useRef, useState } from 'react';
import { Mic, Search } from 'lucide-react';

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  enableVoice?: boolean;
  autoFocus?: boolean;
  onSubmit?: (value: string) => void;
}

const SearchBar: React.FC<SearchBarProps> = ({
  value,
  onChange,
  placeholder = 'Search...',
  enableVoice = true,
  autoFocus = false,
  onSubmit,
}) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const [listening, setListening] = useState(false);

  useEffect(() => {
    if (autoFocus) inputRef.current?.focus();
  }, [autoFocus]);

  const startVoice = () => {
    // Web Speech API stub: does not throw if unavailable (Tauri webviews may not expose it)
    const SpeechRecognition =
      (typeof window !== 'undefined' &&
        ((window as any).SpeechRecognition || (window as any).webkitSpeechRecognition)) ||
      null;
    if (!SpeechRecognition) {
      setListening(false);
      return;
    }
    const rec = new SpeechRecognition();
    rec.continuous = false;
    rec.interimResults = false;
    rec.lang = 'en-US';
    rec.onresult = (e: { results?: ArrayLike<ArrayLike<{ transcript?: string }>> }) => {
      const transcript = e.results?.[0]?.[0]?.transcript ?? '';
      onChange(transcript);
      setListening(false);
    };
    rec.onerror = () => setListening(false);
    rec.onend = () => setListening(false);
    setListening(true);
    try {
      rec.start();
    } catch {
      setListening(false);
    }
  };

  return (
    <div className="relative flex items-center w-full">
      <Search className="absolute left-3 h-4 w-4 text-pos-muted" aria-hidden="true" />
      <input
        ref={inputRef}
        type="search"
        value={value}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && onSubmit?.(value)}
        className="w-full h-11 pl-9 pr-11 rounded-lg bg-white/6 border border-pos-border text-sm text-pos-text placeholder:text-pos-muted outline-none focus:border-white/25 transition-colors"
      />
      {enableVoice && (
        <button
          type="button"
          onClick={startVoice}
          className={`absolute right-2 w-8 h-8 flex items-center justify-center rounded-lg transition ${
            listening
              ? 'bg-destructive text-white animate-pulse'
              : 'text-pos-muted hover:bg-white/10'
          }`}
          aria-label="Voice search"
          title="Voice search"
        >
          <Mic className="h-4 w-4" aria-hidden="true" />
        </button>
      )}
    </div>
  );
};

export default SearchBar;
