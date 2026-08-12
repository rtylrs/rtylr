import Link from 'next/link';

const releases = [
  ['AUGUST 12TH, 2026', 'Self-host rtylr on your own infrastructure', '/self-hosting'],
  ['AUGUST 12TH, 2026', 'Docker Compose and Helm install guides', '/self-hosting/compose'],
  ['AUGUST 11TH, 2026', 'A unified docs experience for every rtylr app', '/getting-started/introduction'],
  ['AUGUST 10TH, 2026', 'API and platform reference updates', '/api/overview'],
];

const platformCards = [
  ['Apps', 'POS, ERP, HR, CRM, Finance, Flow, Store and more.', '/apps/hub'],
  ['Self-hosting', 'Deploy rtylr on infrastructure you control.', '/self-hosting'],
  ['API', 'Build connected workflows on the rtylr data model.', '/api/overview'],
  ['Getting started', 'Go from first login to a working organization.', '/getting-started/introduction'],
  ['Platform', 'Extend rtylr with addons and integrations.', '/platform/addons'],
  ['Help', 'Find answers and practical troubleshooting guidance.', '/faq'],
];

export function HomePage() {
  return (
    <>
      <div className="announcement"><span>NEW</span> Self-host rtylr on your own infrastructure <b>↗</b></div>
      <main className="home-page">
        <section className="home-hero">
          <div className="home-hero__copy">
            <p className="eyebrow">RTYLR DOCS</p>
            <h1>Run your business, your way</h1>
            <p className="hero-description">The operating system for modern businesses. Bring POS, ERP, HR, CRM, Finance, Flow, Store and more together in one secure, connected platform.</p>
            <Link className="primary-button" href="/getting-started/introduction">GET STARTED WITH RTYLR</Link>
          </div>
          <div className="hero-visual" aria-label="An abstract rtylr platform illustration">
            <div className="hero-visual__panel hero-visual__panel--center"><img src="/logo.png" alt="" /></div>
            <div className="hero-visual__panel hero-visual__panel--right"><span>r</span></div>
            <div className="hero-visual__orb" />
          </div>
        </section>

        <section className="home-section" id="whats-new">
          <h2>What&apos;s new</h2>
          <p className="section-lede">Updates, changes, and improvements across rtylr.</p>
          <div className="release-list">
            {releases.map(([date, title, href]) => (
              <div className="release-item" key={title}>
                <div className="release-date"><span />{date}</div>
                <Link href={href}>{title}</Link>
              </div>
            ))}
          </div>
        </section>

        <section className="home-section" id="explore-the-platform">
          <h2>Explore the platform</h2>
          <p className="section-lede">Everything you need to build, operate, and extend your business.</p>
          <div className="platform-grid">
            {platformCards.map(([title, detail, href]) => (
              <Link className="platform-card" href={href} key={title}>
                <span className="platform-card__icon" aria-hidden="true">↗</span>
                <strong>{title}</strong>
                <span>{detail}</span>
              </Link>
            ))}
          </div>
        </section>
      </main>
    </>
  );
}
