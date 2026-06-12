import React from 'react';
import { APP_URLS } from '@voxire/config';

export default function AboutWindow(): React.JSX.Element {
  return (
    <div
      style={{
        height: '100%',
        overflowY: 'auto',
        background: '#151515',
        color: '#e8e6e1',
        fontFamily: "'Inter', system-ui, sans-serif",
      }}
    >
      {/* Hero */}
      <div
        style={{
          padding: '72px 80px 64px',
          textAlign: 'center',
          borderBottom: '1px solid rgba(255,255,255,0.06)',
        }}
      >
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: 6,
            padding: '4px 12px',
            background: 'rgba(255,255,255,0.06)',
            border: '1px solid rgba(255,255,255,0.1)',
            borderRadius: 20,
            fontSize: 11,
            fontWeight: 600,
            letterSpacing: '0.08em',
            textTransform: 'uppercase',
            color: 'rgba(255,255,255,0.5)',
            marginBottom: 32,
          }}
        >
          <span
            style={{
              width: 6,
              height: 6,
              borderRadius: '50%',
              background: '#28c840',
              display: 'inline-block',
            }}
          />
          About Voxire & rtylr
        </div>
        <h1
          style={{
            fontSize: 52,
            fontWeight: 800,
            lineHeight: 1.08,
            letterSpacing: '-0.03em',
            margin: '0 0 24px',
            color: '#e8e6e1',
          }}
        >
          An operating system
          <br />
          for real businesses.
        </h1>
        <p
          style={{
            fontSize: 16,
            color: 'rgba(255,255,255,0.5)',
            maxWidth: 600,
            margin: '0 auto',
            lineHeight: 1.7,
          }}
        >
          Most small businesses run on six to eight disconnected apps. They export CSVs all Friday.
          They reconcile on Sundays. They pay for integrations they hate. We think that's a bug in
          the industry. rtylr is our fix.
        </p>
      </div>

      {/* Mission */}
      <div
        style={{
          padding: '64px 80px',
          borderBottom: '1px solid rgba(255,255,255,0.06)',
          display: 'grid',
          gridTemplateColumns: '1fr 1fr',
          gap: 64,
        }}
      >
        <div>
          <div
            style={{
              fontSize: 11,
              fontWeight: 600,
              letterSpacing: '0.1em',
              textTransform: 'uppercase',
              color: 'rgba(255,255,255,0.35)',
              marginBottom: 16,
            }}
          >
            Our Mission
          </div>
          <h2
            style={{
              fontSize: 32,
              fontWeight: 700,
              letterSpacing: '-0.02em',
              margin: '0 0 20px',
              color: '#e8e6e1',
            }}
          >
            Stop the CSV life.
          </h2>
          <p style={{ fontSize: 14, color: 'rgba(255,255,255,0.5)', lineHeight: 1.75, margin: 0 }}>
            Commerce software should work the way commerce does - one sale, one inventory drop, one
            payroll line, one customer record. Not seven apps pretending they talk to each other.
            rtylr replaces the whole stack with a single platform where everything updates at once,
            because it's all the same database underneath.
          </p>
        </div>
        <div>
          <div
            style={{
              fontSize: 11,
              fontWeight: 600,
              letterSpacing: '0.1em',
              textTransform: 'uppercase',
              color: 'rgba(255,255,255,0.35)',
              marginBottom: 16,
            }}
          >
            What We Build
          </div>
          <h2
            style={{
              fontSize: 32,
              fontWeight: 700,
              letterSpacing: '-0.02em',
              margin: '0 0 20px',
              color: '#e8e6e1',
            }}
          >
            One platform. Six modules.
          </h2>
          <p style={{ fontSize: 14, color: 'rgba(255,255,255,0.5)', lineHeight: 1.75, margin: 0 }}>
            POS, ERP, CRM, HR, Finance & Flow - built as a unified system from the ground up, not
            stitched together from acquisitions. Every module shares the same data model, the same
            permissions layer, and the same AI core.
          </p>
        </div>
      </div>

      {/* Stats */}
      <div
        style={{
          padding: '56px 80px',
          display: 'grid',
          gridTemplateColumns: 'repeat(3, 1fr)',
          gap: 0,
          borderBottom: '1px solid rgba(255,255,255,0.06)',
        }}
      >
        {[
          { num: '10k+', label: 'Businesses served' },
          { num: '6', label: 'Integrated modules' },
          { num: '2024', label: 'Founded' },
        ].map((s, i) => (
          <div
            key={s.label}
            style={{
              textAlign: 'center',
              padding: '0 40px',
              borderRight: i < 2 ? '1px solid rgba(255,255,255,0.06)' : 'none',
            }}
          >
            <div
              style={{ fontSize: 40, fontWeight: 800, letterSpacing: '-0.03em', color: '#e8e6e1' }}
            >
              {s.num}
            </div>
            <div style={{ fontSize: 13, color: 'rgba(255,255,255,0.4)', marginTop: 6 }}>
              {s.label}
            </div>
          </div>
        ))}
      </div>

      {/* CTA */}
      <div style={{ padding: '48px 80px', textAlign: 'center' }}>
        <a
          href={`${APP_URLS.auth}/register`}
          style={{
            display: 'inline-block',
            padding: '12px 28px',
            background: '#e8e6e1',
            color: '#0a0a0a',
            fontWeight: 700,
            fontSize: 14,
            borderRadius: 7,
            textDecoration: 'none',
          }}
        >
          Get started free
        </a>
      </div>
    </div>
  );
}
