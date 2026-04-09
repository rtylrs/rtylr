import React, { useState } from 'react';
import { Copy, Check, Download, RefreshCcw } from 'lucide-react';
import notify from '../lib/notify/notify';

export interface RecoveryCodeListProps {
  codes: string[];
  onRegenerate?: () => void;
}

const RecoveryCodeList: React.FC<RecoveryCodeListProps> = ({
  codes,
  onRegenerate,
}) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    try {
      navigator.clipboard.writeText(codes.join('\n'));
      setCopied(true);
      notify.success('Recovery codes copied');
      setTimeout(() => setCopied(false), 1500);
    } catch {
      notify.error('Could not copy to clipboard');
    }
  };

  const handleDownload = () => {
    const blob = new Blob([codes.join('\n')], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'rtylr-recovery-codes.txt';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    notify.success('Recovery codes downloaded');
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="grid grid-cols-2 gap-2 rounded-xl border border-white/10 bg-white/[0.03] p-4 font-mono">
        {codes.map((code, i) => (
          <div
            key={i}
            className="flex items-center gap-2 text-sm text-white/85 tabular-nums"
          >
            <span className="text-white/30 select-none">
              {(i + 1).toString().padStart(2, '0')}
            </span>
            <span>{code}</span>
          </div>
        ))}
      </div>
      <div className="flex flex-wrap gap-2">
        <button
          type="button"
          onClick={handleCopy}
          className="inline-flex items-center gap-1.5 rounded-lg border border-white/10 bg-white/[0.04] hover:bg-white/[0.08] px-3 py-2 text-xs font-semibold text-white/80 transition-colors"
        >
          {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
          {copied ? 'Copied' : 'Copy all'}
        </button>
        <button
          type="button"
          onClick={handleDownload}
          className="inline-flex items-center gap-1.5 rounded-lg border border-white/10 bg-white/[0.04] hover:bg-white/[0.08] px-3 py-2 text-xs font-semibold text-white/80 transition-colors"
        >
          <Download className="h-3.5 w-3.5" />
          Download .txt
        </button>
        {onRegenerate && (
          <button
            type="button"
            onClick={onRegenerate}
            className="inline-flex items-center gap-1.5 rounded-lg border border-white/10 bg-white/[0.04] hover:bg-white/[0.08] px-3 py-2 text-xs font-semibold text-white/80 transition-colors"
          >
            <RefreshCcw className="h-3.5 w-3.5" />
            Regenerate
          </button>
        )}
      </div>
      <p className="text-[11px] text-white/45 leading-relaxed">
        Save these codes in a secure place. Each can be used once if you lose
        access to your authenticator app.
      </p>
    </div>
  );
};

export default RecoveryCodeList;
