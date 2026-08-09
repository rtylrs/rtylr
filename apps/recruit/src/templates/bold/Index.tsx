import React from 'react';
import { Link } from '@tanstack/react-router';
import { ArrowUpRight } from 'lucide-react';

import Header from '../shared/Header';
import Sections from '../shared/Sections';
import type { IndexProps } from '../resolver';

// Bold - saturated accent block, oversized type, aggressive hover.
// For businesses that want the careers page to feel like a campaign
// landing page. Best paired with a strong primary_color.
const BoldIndex: React.FC<IndexProps> = ({ careers, customization, sections }) => {
  return (
    <div className="min-h-screen bg-white">
      <Header customization={customization} />

      <section style={{ background: 'var(--recruit-accent)' }}>
        <div className="mx-auto max-w-6xl px-6 py-28 text-white sm:py-40">
          <div className="mb-6 inline-block border border-white/40 px-3 py-1 text-[10px] uppercase tracking-[0.3em]">
            {customization.brandName} · Careers
          </div>
          <h1 className="text-6xl font-black leading-[0.95] tracking-tight sm:text-8xl">
            Work
            <br />
            that moves
            <br />
            the needle.
          </h1>
          {customization.aboutBlurb && (
            <p className="mt-10 max-w-2xl text-xl leading-relaxed text-white/80">
              {customization.aboutBlurb}
            </p>
          )}
        </div>
      </section>

      <section className="mx-auto max-w-6xl px-6 py-24">
        <div className="mb-12 flex items-baseline justify-between">
          <h2 className="text-3xl font-bold tracking-tight sm:text-4xl">Open positions</h2>
          <div className="text-sm font-medium" style={{ color: 'var(--recruit-muted)' }}>
            {careers.length} {careers.length === 1 ? 'role' : 'roles'}
          </div>
        </div>

        <div className="grid gap-0 border-t" style={{ borderColor: 'var(--recruit-border)' }}>
          {careers.length === 0 ? (
            <div className="py-16 text-center text-sm" style={{ color: 'var(--recruit-muted)' }}>
              No open roles right now. Check back soon.
            </div>
          ) : (
            careers.map((c) => (
              <Link
                key={c.id}
                to={`/${customization.slug}/${c.slug}` as any}
                className="group flex items-center justify-between border-b py-8 transition hover:pl-4"
                style={{ borderColor: 'var(--recruit-border)' }}
              >
                <div>
                  <div className="text-2xl font-bold leading-tight sm:text-3xl">{c.title}</div>
                  <div
                    className="mt-2 flex items-center gap-3 text-sm font-medium"
                    style={{ color: 'var(--recruit-muted)' }}
                  >
                    {c.department && <span>{c.department}</span>}
                    {c.department && c.location && <span>·</span>}
                    {c.location && <span>{c.location}</span>}
                    {c.workMode && <span>· {c.workMode}</span>}
                  </div>
                </div>
                <ArrowUpRight
                  className="h-6 w-6 transition group-hover:rotate-12"
                  style={{ color: 'var(--recruit-accent)' }}
                />
              </Link>
            ))
          )}
        </div>
      </section>
      <Sections sections={sections} />
    </div>
  );
};

export default BoldIndex;
