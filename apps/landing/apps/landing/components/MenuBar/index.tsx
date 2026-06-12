import React from 'react';
import { useRouter } from 'next/router';
import { APP_URLS } from '@voxire/config';
import { AppMenu } from './AppMenu';
import { MenuBarMenu } from './MenuBarMenu';
import { useDesktopStore, WindowType } from '../../stores/useDesktopStore';

export const MenuBar: React.FC = () => {
  const router = useRouter();
  const addWindow = useDesktopStore((s) => s.addWindow);

  const productsItems = [
    { label: 'POS', onClick: () => router.push('/pos') },
    { label: 'ERP', onClick: () => router.push('/erp') },
    { label: 'CRM', onClick: () => router.push('/crm') },
    { label: 'HR', onClick: () => router.push('/hr') },
    { label: 'Finance', onClick: () => router.push('/finance') },
    { label: 'Flow', onClick: () => router.push('/flow') },
  ];

  const docsItems = [
    {
      label: 'Documentation',
      onClick: () => addWindow('docs' as WindowType, 'https://docs.rtylr.com'),
    },
    { separator: true as const },
    {
      label: 'API Reference',
      onClick: () => addWindow('docs' as WindowType, 'https://docs.rtylr.com/api'),
    },
    {
      label: 'Changelog',
      onClick: () => addWindow('docs' as WindowType, 'https://docs.rtylr.com/changelog'),
    },
  ];

  const moreItems = [
    { label: 'About', onClick: () => router.push('/about') },
    { label: 'Partners', onClick: () => router.push('/partners') },
    { label: 'Contact', onClick: () => router.push('/contact') },
    { separator: true as const },
    { label: 'Hardware', onClick: () => router.push('/hardware') },
    { label: 'Integrations', onClick: () => router.push('/integrations') },
    { label: 'Help Center', onClick: () => router.push('/support') },
    { label: 'Trust Center', onClick: () => router.push('/trust') },
    { label: 'Security', onClick: () => router.push('/legal/security') },
    { separator: true as const },
    { label: 'Privacy Policy', onClick: () => router.push('/legal/privacy-policy') },
    { label: 'Terms of Use', onClick: () => router.push('/legal/terms') },
    { separator: true as const },
    { label: 'Keyboard Shortcuts', onClick: () => addWindow('shortcuts' as WindowType) },
    {
      label: 'System Status',
      onClick: () => addWindow('docs' as WindowType, 'https://status.rtylr.com'),
    },
    { separator: true as const },
    { label: 'Pricing', onClick: () => addWindow('pricing' as WindowType, '/pricing?embed=1') },
  ];

  return (
    <div
      style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        height: 36,
        display: 'flex',
        alignItems: 'center',
        padding: '0 16px',
        background: 'rgba(12,12,14,0.92)',
        backdropFilter: 'blur(20px)',
        borderBottom: '1px solid rgba(255,255,255,0.07)',
        zIndex: 9000,
        gap: 2,
      }}
    >
      <AppMenu />

      <div style={{ display: 'flex', gap: 2, marginLeft: 6 }}>
        <MenuBarMenu label="Products" items={productsItems} />
        <MenuBarMenu
          label="Pricing"
          items={[
            {
              label: 'Pricing Plans',
              onClick: () => addWindow('pricing' as WindowType, '/pricing?embed=1'),
            },
          ]}
        />
        <MenuBarMenu label="Docs" items={docsItems} />
        <MenuBarMenu label="Blog" items={[{ label: 'Blog', onClick: () => router.push('/blog') }]} />
        <MenuBarMenu label="More" items={moreItems} />
      </div>

      <div style={{ marginLeft: 'auto', display: 'flex', alignItems: 'center', gap: 8 }}>
        <a
          href={APP_URLS.auth}
          style={{
            fontSize: 12,
            fontWeight: 500,
            color: 'rgba(255,255,255,0.7)',
            textDecoration: 'none',
            padding: '5px 12px',
            borderRadius: 5,
            transition: 'color 0.15s',
          }}
          onMouseEnter={(e) => ((e.currentTarget as HTMLAnchorElement).style.color = '#e8e6e1')}
          onMouseLeave={(e) =>
            ((e.currentTarget as HTMLAnchorElement).style.color = 'rgba(255,255,255,0.7)')
          }
        >
          Sign in
        </a>
        <a
          href={`${APP_URLS.auth}/register`}
          style={{
            fontSize: 12,
            fontWeight: 600,
            color: '#0a0a0a',
            textDecoration: 'none',
            padding: '5px 14px',
            background: '#e8e6e1',
            borderRadius: 5,
            transition: 'background 0.15s',
            whiteSpace: 'nowrap',
          }}
          onMouseEnter={(e) =>
            ((e.currentTarget as HTMLAnchorElement).style.background = '#ffffff')
          }
          onMouseLeave={(e) =>
            ((e.currentTarget as HTMLAnchorElement).style.background = '#e8e6e1')
          }
        >
          Sign up
        </a>
      </div>
    </div>
  );
};
