import React from 'react';
import { HelpCircle } from 'lucide-react';

export interface HelpLinkProps {
  href?: string;
  label?: string;
}

const HelpLink: React.FC<HelpLinkProps> = ({
  href = 'mailto:support@rtylr.com',
  label = 'Trouble signing in?',
}) => {
  return (
    <a
      href={href}
      target={href.startsWith('http') ? '_blank' : undefined}
      rel={href.startsWith('http') ? 'noopener noreferrer' : undefined}
      className="inline-flex items-center gap-1.5 text-xs text-white/50 hover:text-white transition-colors"
    >
      <HelpCircle className="h-3.5 w-3.5" />
      {label}
    </a>
  );
};

export default HelpLink;
