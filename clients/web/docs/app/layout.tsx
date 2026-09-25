import type { Metadata } from 'next';
import { getSearchEntries } from '../lib/content';
import { SiteFooter } from '../components/site-footer';
import { SiteHeader } from '../components/site-header';
import './globals.css';

export const metadata: Metadata = {
  title: {
    default: 'rtylr Docs',
    template: '%s · rtylr Docs',
  },
  description: 'Documentation for the rtylr business operating system.',
  icons: {
    icon: [
      { url: '/favicon.ico' },
      { url: '/favicon/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
      { url: '/favicon/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
    ],
    apple: [{ url: '/favicon/apple-touch-icon.png', sizes: '180x180', type: 'image/png' }],
  },
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en">
      <body>
        <SiteHeader searchEntries={getSearchEntries()} />
        {children}
        <SiteFooter />
      </body>
    </html>
  );
}
