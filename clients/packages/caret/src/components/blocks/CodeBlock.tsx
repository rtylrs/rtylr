import React, { useState } from 'react';
import { Copy, Check } from 'lucide-react';
import { caretTokens, caretRadius } from '../../ui/tokens';
import IconButton from '../../ui/IconButton';

export interface CodeBlockProps {
  code: string;
  language?: string;
}

export const CodeBlock: React.FC<CodeBlockProps> = ({ code, language }) => {
  const [copied, setCopied] = useState(false);

  const copy = async (): Promise<void> => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard API can fail in iframes or when blurred - fail silent.
    }
  };

  return (
    <div
      style={{
        background: caretTokens.surfaceHigh,
        borderRadius: caretRadius.md,
        border: `1px solid ${caretTokens.border}`,
        overflow: 'hidden',
        margin: '4px 0',
      }}
    >
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '4px 8px',
          background: caretTokens.surface,
          borderBottom: `1px solid ${caretTokens.border}`,
        }}
      >
        <span style={{ fontSize: 11, color: caretTokens.textMuted, textTransform: 'lowercase' }}>
          {language || 'code'}
        </span>
        <IconButton size="sm" label={copied ? 'Copied' : 'Copy code'} onClick={copy}>
          {copied ? <Check size={11} /> : <Copy size={11} />}
        </IconButton>
      </div>
      <pre
        style={{
          margin: 0,
          padding: '8px 10px',
          fontFamily: "SFMono-Regular, Menlo, Consolas, 'Liberation Mono', monospace",
          fontSize: 12,
          lineHeight: 1.5,
          color: caretTokens.text,
          overflowX: 'auto',
          whiteSpace: 'pre',
        }}
      >
        <code>{code}</code>
      </pre>
    </div>
  );
};

export default CodeBlock;
