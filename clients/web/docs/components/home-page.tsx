import Link from 'next/link';

const quickStartCards = [
  ['01', 'Set up rtylr', 'Understand the platform, create your organization, and get your first team moving.', '/getting-started/introduction'],
  ['02', 'Build with the API', 'Connect your own tools and workflows to the rtylr data model.', '/api/overview'],
];

const appCards = [
  ['HU', 'Hub', 'Your command center for people, work, and connected business data.', '/apps/hub'],
  ['PO', 'POS', 'Sell in person and keep every transaction connected to inventory.', '/apps/pos'],
  ['ER', 'ERP', 'Manage products, stock, purchasing, and supply chain operations.', '/apps/erp'],
  ['HR', 'HR', 'Organize your people, leave, payroll, and day-to-day operations.', '/apps/hr'],
  ['CR', 'CRM', 'Keep customer relationships, conversations, and follow-ups together.', '/apps/crm'],
  ['FI', 'Finance', 'See the financial picture behind every sale and decision.', '/apps/finance'],
  ['FL', 'Flow', 'Turn repeatable work into clear workflows your team can run.', '/apps/flow'],
  ['ST', 'Store', 'Bring your catalog and commerce experience online.', '/apps/store'],
];

export function HomePage() {
  return (
    <main className="home-page">
      <section className="home-hero">
        <div className="home-hero__copy">
          <p className="eyebrow"><span className="eyebrow__mark" /> RTYLR DOCUMENTATION</p>
          <h1>Everything you need to run rtylr.</h1>
          <p className="hero-description">One place to learn the platform, launch every app, and connect your data.</p>
          <div className="hero-actions">
            <Link className="primary-button" href="/getting-started/introduction">Start with rtylr <span aria-hidden="true">↗</span></Link>
          </div>
        </div>
        <div className="hero-visual" aria-label="The rtylr platform and its apps">
          <div className="hero-visual__header"><span>RTYLR / PLATFORM</span><span>08 APPS</span></div>
          <div className="hero-visual__logo"><img src="/logo-light.png" alt="rtylr" /></div>
          <div className="hero-visual__apps">
            {appCards.slice(0, 6).map(([code, title]) => (
              <span key={code}><b>{code}</b>{title}</span>
            ))}
          </div>
          <div className="hero-visual__footer"><span>ONE DATA MODEL</span><span>ONE PLATFORM</span></div>
        </div>
      </section>

      <section className="home-section home-section--start" id="start-here">
        <div className="section-heading">
          <div>
            <p className="section-kicker">START HERE</p>
            <h2>Choose your next move.</h2>
          </div>
          <p className="section-lede">A short path to the part of rtylr you need right now.</p>
        </div>
        <div className="quick-start-grid">
          {quickStartCards.map(([number, title, detail, href]) => (
            <Link className="quick-start-card" href={href} key={title}>
              <span className="quick-start-card__number">{number}</span>
              <span className="quick-start-card__arrow" aria-hidden="true">↗</span>
              <strong>{title}</strong>
              <span>{detail}</span>
            </Link>
          ))}
        </div>
      </section>

      <section className="home-section home-section--apps" id="apps">
        <div className="section-heading">
          <div>
            <p className="section-kicker">THE PLATFORM</p>
            <h2>One system. Every team.</h2>
          </div>
          <p className="section-lede">Explore the apps that make up the rtylr operating system.</p>
        </div>
        <div className="app-index">
          {appCards.map(([code, title, detail, href]) => (
            <Link className="app-index__item" href={href} key={code}>
              <span className="app-index__code">{code}</span>
              <span className="app-index__body"><strong>{title}</strong><span>{detail}</span></span>
              <span className="app-index__arrow" aria-hidden="true">↗</span>
            </Link>
          ))}
        </div>
      </section>
    </main>
  );
}
