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
    icon: '/favicon.ico',
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
