import React from 'react';
import { Sparkles } from 'lucide-react';
import { useCaretStore } from '@voxire/caret';

/** Hyperlink-styled "Ask Caret about this" entry point. Submits the prompt
 *  to Caret immediately when the user clicks. */
export interface AskCaretLinkProps {
  prompt: string;
  children?: React.ReactNode;
}

export const AskCaretLink: React.FC<AskCaretLinkProps> = ({ prompt, children }) => {
  const open = useCaretStore((s) => s.open);
  return (
    <button
      type="button"
      onClick={() => {
        open();
        // The actual send is wired by Caret's panel — the user can edit before sending.
        sessionStorage.setItem('caret-pending-prompt', prompt);
      }}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 4,
        background: 'transparent',
        border: 'none',
        padding: 0,
        color: '#8b5cf6',
        fontSize: 12.5,
        cursor: 'pointer',
        textDecoration: 'underline',
        textDecorationStyle: 'dotted',
      }}
    >
      <Sparkles size={11} />
      {children ?? 'Ask Caret'}
    </button>
  );
};

export default AskCaretLink;
