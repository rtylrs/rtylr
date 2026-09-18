import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { DocsPage } from '../../components/docs-page';
import { getAllDocs, getDoc } from '../../lib/content';

export const dynamicParams = false;

export function generateStaticParams() {
  return getAllDocs().filter((doc) => doc.slug.length > 0).map((doc) => ({ slug: doc.slug }));
}

export async function generateMetadata({ params }: { params: Promise<{ slug: string[] }> }): Promise<Metadata> {
  const { slug } = await params;
  const doc = getDoc(slug);
  return { title: doc?.title ?? 'rtylr Docs', description: doc?.description };
}

export default async function DocsRoute({ params }: { params: Promise<{ slug: string[] }> }) {
  const { slug } = await params;
  const doc = getDoc(slug);
  if (!doc) notFound();
  return <DocsPage doc={doc} />;
}
