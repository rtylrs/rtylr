import type { Metadata } from 'next';
import { HomePage } from '../components/home-page';

export const metadata: Metadata = {
  title: 'rtylr Docs',
  description: 'Documentation for the rtylr business operating system.',
};

export default function Page() {
  return <HomePage />;
}
