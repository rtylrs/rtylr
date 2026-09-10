import React from 'react';
import { useTranslation } from 'react-i18next';
import { cdnImageUrl } from '@voxire/config';
import { ChevronDown } from 'lucide-react';

import type { Section, PerkItem, FaqItem, ValueItem } from '../../shared/lib/types';

/**
 * Sections renders the tenant's content blocks between the hero and
 * the jobs list. A single shared component so every template gets
 * the same structure; the visual layer differs via CSS vars picked
 * up from applyBrand().
 *
 * Rendering is defensive: unknown kinds fall back to a plain
 * title+body card. Broken items_json never crashes the page.
 */
interface SectionsProps {
  sections: Section[];
}

const Sections: React.FC<SectionsProps> = ({ sections }) => {
  if (sections.length === 0) return null;
  return (
    <div className="mx-auto max-w-5xl px-6 py-16 space-y-20">
      {sections.map((s) => (
        <SectionBlock key={s.id} section={s} />
      ))}
    </div>
  );
};

const SectionBlock: React.FC<{ section: Section }> = ({ section: s }) => {
  switch (s.kind) {
    case 'perks':
      return <PerksBlock section={s} />;
    case 'culture':
      return <CultureBlock section={s} />;
    case 'testimonial':
      return <TestimonialBlock section={s} />;
    case 'video':
      return <VideoBlock section={s} />;
    case 'faq':
      return <FaqBlock section={s} />;
    case 'values':
      return <ValuesBlock section={s} />;
    case 'cta':
      return <CtaBlock section={s} />;
    default:
      return <FallbackBlock section={s} />;
  }
};

function parseItems<T>(json: string): T[] {
  if (!json) return [];
  try {
    const parsed = JSON.parse(json);
    return Array.isArray(parsed) ? (parsed as T[]) : [];
  } catch {
    return [];
  }
}

// ── per-kind renderers ───────────────────────────────────────────────

const PerksBlock: React.FC<{ section: Section }> = ({ section }) => {
  const items = parseItems<PerkItem>(section.itemsJson);
  if (items.length === 0 && !section.title) return null;
  return (
    <section>
      {section.title && (
        <h2 className="font-display text-3xl font-semibold tracking-tight sm:text-4xl">
          {section.title}
        </h2>
      )}
      <div className="mt-8 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
        {items.map((p, i) => (
          <div
            key={i}
            className="rounded-lg border p-5"
            style={{ borderColor: 'var(--recruit-border)' }}
          >
            <div
              className="inline-flex h-9 w-9 items-center justify-center rounded-md"
              style={{ background: 'color-mix(in srgb, var(--recruit-accent) 15%, transparent)' }}
            >
              <span className="text-sm font-bold" style={{ color: 'var(--recruit-accent)' }}>
                {p.icon?.slice(0, 2).toUpperCase() ?? '·'}
              </span>
            </div>
            <div className="mt-3 font-semibold">{p.label}</div>
            {p.blurb && (
              <div className="mt-1 text-sm" style={{ color: 'var(--recruit-muted)' }}>
                {p.blurb}
              </div>
            )}
          </div>
        ))}
      </div>
    </section>
  );
};

const CultureBlock: React.FC<{ section: Section }> = ({ section }) => {
  const mediaUrl = cdnImageUrl(section.mediaUrl);
  return (
    <section className="grid gap-10 sm:grid-cols-2 sm:items-center">
      {mediaUrl && <img src={mediaUrl} alt="" className="h-80 w-full rounded-lg object-cover" />}
      <div className={mediaUrl ? '' : 'sm:col-span-2'}>
        {section.title && (
          <h2 className="font-display text-3xl font-semibold tracking-tight sm:text-4xl">
            {section.title}
          </h2>
        )}
        {section.body && (
          <p
            className="mt-4 whitespace-pre-wrap text-base leading-relaxed"
            style={{ color: 'var(--recruit-fg)' }}
          >
            {section.body}
          </p>
        )}
      </div>
    </section>
  );
};

const TestimonialBlock: React.FC<{ section: Section }> = ({ section }) => (
  <section className="mx-auto max-w-2xl text-center">
    <blockquote
      className="font-display text-2xl leading-relaxed sm:text-3xl"
      style={{ color: 'var(--recruit-fg)' }}
    >
      &ldquo;{section.body}&rdquo;
    </blockquote>
    {section.mediaUrl && (
      <div className="mt-6 text-sm" style={{ color: 'var(--recruit-muted)' }}>
        - {section.mediaUrl}
      </div>
    )}
  </section>
);

const VideoBlock: React.FC<{ section: Section }> = ({ section }) => {
  const { t } = useTranslation();
  if (!section.mediaUrl) return null;
  // Accept YouTube / Vimeo / Loom share URLs. Normalise to an embed
  // URL when possible; fall back to a plain link if the host isn't
  // recognised (rather than gambling with an arbitrary iframe src).
  const embed = toEmbedUrl(section.mediaUrl);
  return (
    <section>
      {section.title && (
        <h2 className="mb-6 font-display text-3xl font-semibold tracking-tight sm:text-4xl">
          {section.title}
        </h2>
      )}
      {embed ? (
        <div
          className="aspect-video overflow-hidden rounded-lg border"
          style={{ borderColor: 'var(--recruit-border)' }}
        >
          <iframe
            src={embed}
            title={section.title || 'Video'}
            className="h-full w-full"
            frameBorder="0"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowFullScreen
          />
        </div>
      ) : (
        <a href={section.mediaUrl} className="underline" style={{ color: 'var(--recruit-accent)' }}>
          {t('public.sections.watchVideo')}
        </a>
      )}
    </section>
  );
};

const FaqBlock: React.FC<{ section: Section }> = ({ section }) => {
  const items = parseItems<FaqItem>(section.itemsJson);
  return (
    <section>
      {section.title && (
        <h2 className="font-display text-3xl font-semibold tracking-tight sm:text-4xl">
          {section.title}
        </h2>
      )}
      <div className="mt-6 space-y-2">
        {items.map((f, i) => (
          <details
            key={i}
            className="group rounded-lg border"
            style={{ borderColor: 'var(--recruit-border)' }}
          >
            <summary className="flex cursor-pointer items-center justify-between px-4 py-3 text-sm font-medium">
              {f.question}
              <ChevronDown className="h-4 w-4 transition group-open:rotate-180" />
            </summary>
            <div
              className="border-t px-4 py-3 text-sm leading-relaxed"
              style={{ borderColor: 'var(--recruit-border)', color: 'var(--recruit-muted)' }}
            >
              {f.answer}
            </div>
          </details>
        ))}
      </div>
    </section>
  );
};

const ValuesBlock: React.FC<{ section: Section }> = ({ section }) => {
  const items = parseItems<ValueItem>(section.itemsJson);
  return (
    <section>
      {section.title && (
        <h2 className="font-display text-3xl font-semibold tracking-tight sm:text-4xl">
          {section.title}
        </h2>
      )}
      <div className="mt-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {items.map((v, i) => (
          <div key={i} className="border-l-2 pl-4" style={{ borderColor: 'var(--recruit-accent)' }}>
            <div className="font-semibold">{v.label}</div>
            {v.blurb && (
              <div className="mt-1 text-sm" style={{ color: 'var(--recruit-muted)' }}>
                {v.blurb}
              </div>
            )}
          </div>
        ))}
      </div>
    </section>
  );
};

const CtaBlock: React.FC<{ section: Section }> = ({ section }) => {
  const { t } = useTranslation();
  return (
    <section
      className="rounded-xl p-10 text-center"
      style={{ background: 'var(--recruit-accent)' }}
    >
      {section.title && (
        <h2 className="font-display text-3xl font-semibold tracking-tight text-white sm:text-4xl">
          {section.title}
        </h2>
      )}
      {section.body && <p className="mt-3 text-white/80">{section.body}</p>}
      {section.mediaUrl && (
        <a
          href={section.mediaUrl}
          className="mt-6 inline-flex items-center gap-2 rounded-md bg-white px-5 py-2.5 text-sm font-semibold"
          style={{ color: 'var(--recruit-accent)' }}
        >
          {t('public.sections.getInTouch')}
        </a>
      )}
    </section>
  );
};

const FallbackBlock: React.FC<{ section: Section }> = ({ section }) => (
  <section>
    {section.title && <h2 className="text-2xl font-semibold">{section.title}</h2>}
    {section.body && (
      <p className="mt-3 whitespace-pre-wrap text-base" style={{ color: 'var(--recruit-fg)' }}>
        {section.body}
      </p>
    )}
  </section>
);

// toEmbedUrl normalises the most common video share URLs to their
// iframe-embed equivalents. Unknown hosts return '' so the caller
// renders a plain link instead of a blocked iframe.
function toEmbedUrl(raw: string): string {
  try {
    const url = new URL(raw);
    const host = url.hostname.replace(/^www\./, '');

    if (host === 'youtu.be') {
      return `https://www.youtube.com/embed/${url.pathname.replace(/^\//, '')}`;
    }
    if (host === 'youtube.com' || host === 'm.youtube.com') {
      const v = url.searchParams.get('v');
      if (v) return `https://www.youtube.com/embed/${v}`;
      // /embed/ URLs pass through.
      if (url.pathname.startsWith('/embed/')) return raw;
    }
    if (host === 'vimeo.com') {
      return `https://player.vimeo.com/video${url.pathname}`;
    }
    if (host.endsWith('loom.com') && url.pathname.startsWith('/share/')) {
      return raw.replace('/share/', '/embed/');
    }
  } catch {
    // fall through
  }
  return '';
}

export default Sections;
