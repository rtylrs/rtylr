import React from 'react';
import { ArrowLeft } from 'lucide-react';
import { useNavigate } from 'react-router-dom';

export interface BackLinkProps {
  to?: string;
  label?: string;
  onClick?: () => void;
}

const BackLink: React.FC<BackLinkProps> = ({
  to = '/',
  label = 'Back to sign in',
  onClick,
}) => {
  const navigate = useNavigate();

  return (
    <button
      type="button"
      onClick={onClick ?? (() => navigate(to))}
      className="inline-flex items-center gap-1.5 text-[#64C5B9] font-medium hover:underline underline-offset-4 text-sm"
    >
      <ArrowLeft className="h-4 w-4" />
      {label}
    </button>
  );
};

export default BackLink;
