import Link from 'next/link';

const footerGroups = [
  {
    title: 'Platform',
    links: [
      ['Apps', '/apps/hub'],
      ['API', '/api/overview'],
      ['Addon system', '/platform/addons'],
    ],
  },
  {
    title: 'Capabilities',
    links: [
      ['Getting started', '/getting-started/introduction'],
      ['Multi-tenant', '/getting-started/multi-tenant'],
      ['Flow', '/apps/flow'],
      ['FAQ', '/faq'],
    ],
  },
  {
    title: 'Resources',
    links: [
      ['Documentation', '/'],
      ['rtylr.com', 'https://rtylr.com'],
    ],
  },
];

export function SiteFooter() {
  return (
    <footer className="site-footer">
      <div className="site-footer__grid">
        <div className="site-footer__brand">
          <img src="/logo-light.png" alt="rtylr" />
          <p>One platform for the work that keeps your business moving.</p>
        </div>
        {footerGroups.map((group) => (
          <div className="site-footer__group" key={group.title}>
            <h2>{group.title}</h2>
            <ul>
              {group.links.map(([label, href]) => (
                <li key={href}>
                  {href.startsWith('http') ? (
                    <a href={href} target="_blank" rel="noreferrer">{label}</a>
                  ) : (
                    <Link href={href}>{label}</Link>
                  )}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <div className="site-footer__bottom">
        <div>
          <Link href="/faq">Help</Link>
          <a href="https://github.com/tryrtylr" target="_blank" rel="noreferrer">GitHub</a>
        </div>
        <span>© {new Date().getFullYear()} rtylr</span>
      </div>
    </footer>
  );
}
