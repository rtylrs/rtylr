import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import AppRoutes from './routes/AppRoutes';

function getCookie(name: string): string | null {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop()?.split(';').shift() ?? null;
  return null;
}

const ProtectedApp: React.FC = () => {
  const navigate = useNavigate();

  useEffect(() => {
    const token = getCookie('token');
    if (!token) {
      window.location.href = 'http://localhost:5176?redirect=erp';
    }
  }, []);

  const token = getCookie('token');
  if (!token) {
    return (
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          height: '100vh',
          background: 'var(--hub-bg)',
          color: 'var(--hub-text-muted)',
          fontSize: '14px',
        }}
      >
        Redirecting to login…
      </div>
    );
  }

  return <AppRoutes />;
};

const App: React.FC = () => {
  return <ProtectedApp />;
};

export default App;
