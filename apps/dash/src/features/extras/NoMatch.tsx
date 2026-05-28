import React from 'react';
import { useNavigate } from '@tanstack/react-router';
import { Button } from '../../shared/components/ui';

const NoMatch = (): React.JSX.Element => {
  const navigate = useNavigate();
  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-admin-bg gap-4">
      <p className="text-6xl font-bold text-primary/20">404</p>
      <h1 className="text-xl font-semibold text-primary">Page not found</h1>
      <p className="text-sm text-gray-500">The page you're looking for doesn't exist.</p>
      <Button onClick={() => navigate({ to: '/' })}>Go home</Button>
    </div>
  );
};

export default NoMatch;
