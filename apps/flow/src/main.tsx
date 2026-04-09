import { Suspense } from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { Toaster } from 'sonner';
import App from './App';
import { initTheme } from './stores/themeStore';
import './globals.css';

// Apply persisted theme before first render to prevent flash
initTheme();

const rootElement = document.getElementById('root');

if (rootElement) {
  ReactDOM.createRoot(rootElement).render(
    <Suspense
      fallback={
        <div
          style={{
            display: 'flex',
            height: '100vh',
            alignItems: 'center',
            justifyContent: 'center',
            background: '#0a0a0a',
            color: 'rgba(240,240,238,0.5)',
            fontFamily: 'Plus Jakarta Sans, system-ui, sans-serif',
            fontSize: '14px',
          }}
        >
          Loading Flow…
        </div>
      }
    >
      <BrowserRouter>
        <App />
        <Toaster expand={false} richColors />
      </BrowserRouter>
    </Suspense>
  );
}
