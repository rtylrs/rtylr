import Link from 'next/link';
import { marked } from 'marked';
import { getTableOfContents, slugifyHeading, type Doc } from '../lib/content';
import { sidebarGroups } from './navigation-data';

type DocsPageProps = {
  doc: Doc;
};

function isActive(route: string, currentRoute: string) {
  return route === currentRoute;
}

function contentHtml(body: string) {
  const renderer = new marked.Renderer();
  renderer.heading = ({ tokens, depth }) => {
    const text = renderer.parser.parseInline(tokens);
    return `<h${depth} id="${slugifyHeading(text)}">${text}</h${depth}>\n`;
  };
  return marked.parse(body, { gfm: true, breaks: false, renderer }) as string;
}

export function DocsPage({ doc }: DocsPageProps) {
  const toc = getTableOfContents(doc.body);

  return (
    <>
      <div className="docs-frame">
        <aside className="docs-sidebar" aria-label="Docs sidebar">
          {sidebarGroups.map((group) => (
            <section key={group.label}>
              <h2>{group.label}</h2>
              <ul>
                {group.items.map(([label, href]) => (
                  <li key={href}>
                    <Link className={isActive(href, doc.route) ? 'is-active' : ''} href={href}>{label}</Link>
                  </li>
                ))}
              </ul>
            </section>
          ))}
        </aside>

        <main className="doc-main">
          <nav className="breadcrumbs" aria-label="Breadcrumbs">
            <Link href="/">HOME</Link><span>›</span><span>{doc.slug[0]?.replace(/-/g, ' ').toUpperCase() ?? 'DOCS'}</span><span>›</span><strong>{doc.title.toUpperCase()}</strong>
          </nav>
          <article className="prose" dangerouslySetInnerHTML={{ __html: contentHtml(doc.body) }} />
          <div className="need-help">
            <span>Need more help?</span>
            <Link href="/faq">Read the rtylr FAQ →</Link>
          </div>
        </main>

        <aside className="doc-toc" aria-label="On this page">
          {toc.length > 0 && <p>On this page</p>}
          <ul>
            {toc.map((item) => (
              <li className={item.level === 3 ? 'toc-subitem' : ''} key={item.id}>
                <a href={`#${item.id}`}>{item.title}</a>
              </li>
            ))}
          </ul>
        </aside>
      </div>
    </>
  );
}
