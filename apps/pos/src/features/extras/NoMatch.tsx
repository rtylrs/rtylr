import React from 'react';
import { useNavigate } from 'react-router-dom';

const NoMatch: React.FC = () => {
  const navigate = useNavigate();
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-pos-bg text-pos-text gap-4">
      <span className="text-8xl font-black text-pos-muted">404</span>
      <p className="text-pos-muted text-sm">Page not found</p>
      <button
        className="px-4 py-2 rounded-lg bg-white/10 text-pos-text text-sm hover:bg-white/15 transition-colors"
        onClick={() => navigate('/')}
      >
        Go home
      </button>
    </div>
  );
};

export default NoMatch;
