import React, { useEffect, useState } from 'react';
import { toast } from 'sonner';
import { APP_URLS } from '@voxire/config';
import { ExternalLink } from 'lucide-react';

import { useAdminCustomization, useUpsertCustomization } from '../../shared/hooks/queries/useAdmin';
import type { TemplateId } from '../../shared/lib/types';

/**
 * CustomizationAdmin — edits the business's careers-page branding.
 * Picking a template flips the whole public page layout; the preview
 * link in the header lets the admin open their live page in a new tab
 * as soon as they save.
 */
const TEMPLATES: { id: TemplateId; name: string; blurb: string }[] = [
  {
    id: 'minimal',
    name: 'Minimal',
    blurb: 'Calm, single-column list. Works for every business without tuning.',
  },
  {
    id: 'editorial',
    name: 'Editorial',
    blurb: 'Magazine-style hero with serif display type and generous whitespace.',
  },
  {
    id: 'bold',
    name: 'Bold',
    blurb: 'Saturated accent block, oversized type. Pair with a strong brand color.',
  },
];

const CustomizationAdmin: React.FC = (): React.JSX.Element => {
  const { data } = useAdminCustomization();
  const save = useUpsertCustomization();

  const existing = data?.customization;
  const [slug, setSlug] = useState(existing?.slug ?? '');
  const [brandName, setBrandName] = useState(existing?.brandName ?? '');
  const [tagline, setTagline] = useState(existing?.tagline ?? '');
  const [aboutBlurb, setAboutBlurb] = useState(existing?.aboutBlurb ?? '');
  const [logoUrl, setLogoUrl] = useState(existing?.logoUrl ?? '');
  const [primaryColor, setPrimaryColor] = useState(existing?.primaryColor ?? '#111111');
  const [template, setTemplate] = useState<TemplateId>(
    (existing?.template as TemplateId) ?? 'minimal'
  );
  const [headerHtml, setHeaderHtml] = useState(existing?.headerHtml ?? '');

  // Pull in server state once it arrives (hook is async).
  useEffect(() => {
    if (!existing) return;
    setSlug(existing.slug);
    setBrandName(existing.brandName);
    setTagline(existing.tagline);
    setAboutBlurb(existing.aboutBlurb);
    setLogoUrl(existing.logoUrl);
    setPrimaryColor(existing.primaryColor || '#111111');
    setTemplate((existing.template as TemplateId) ?? 'minimal');
    setHeaderHtml(existing.headerHtml);
  }, [existing]);

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    save.mutate(
      {
        slug,
        brandName,
        tagline,
        aboutBlurb,
        logoUrl,
        primaryColor,
        template,
        headerHtml,
      },
      {
        onSuccess: () => toast.success('Careers page saved'),
        onError: () => toast.error("Couldn't save — check your slug is unique"),
      }
    );
  };

  return (
    <div className="px-10 py-8">
      <div className="flex items-baseline justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Page branding</h1>
          <p className="mt-1 text-sm text-white/60">
            What candidates see at {APP_URLS.recruit}/<span className="text-white">your-slug</span>
          </p>
        </div>
        {slug && (
          <a
            href={`${APP_URLS.recruit}/${slug}`}
            target="_blank"
            rel="noreferrer"
            className="inline-flex items-center gap-2 rounded-md border border-white/10 px-3 py-2 text-sm hover:bg-white/5"
          >
            <ExternalLink className="h-4 w-4" />
            Preview
          </a>
        )}
      </div>

      <form onSubmit={onSubmit} className="mt-8 grid grid-cols-1 gap-8 lg:grid-cols-[1fr_340px]">
        <section className="space-y-5">
          <Row>
            <Field label="URL slug" required hint="a–z, 0–9, dashes">
              <Input value={slug} onChange={setSlug} required />
            </Field>
            <Field label="Brand name" required>
              <Input value={brandName} onChange={setBrandName} required />
            </Field>
          </Row>

          <Field label="Tagline">
            <Input value={tagline} onChange={setTagline} />
          </Field>

          <Field label="About" hint="Short paragraph shown under the hero">
            <Textarea value={aboutBlurb} onChange={setAboutBlurb} rows={4} />
          </Field>

          <Row>
            <Field label="Logo URL">
              <Input value={logoUrl} onChange={setLogoUrl} type="url" />
            </Field>
            <Field label="Primary color" hint="Hex">
              <div className="flex gap-2">
                <input
                  type="color"
                  value={primaryColor}
                  onChange={(e) => setPrimaryColor(e.target.value)}
                  className="h-10 w-16 rounded-md border"
                  style={{ borderColor: '#2a2a2a' }}
                />
                <Input value={primaryColor} onChange={setPrimaryColor} />
              </div>
            </Field>
          </Row>

          <Field label="Custom header HTML" hint="Advanced — injected raw above the brand header">
            <Textarea value={headerHtml} onChange={setHeaderHtml} rows={3} />
          </Field>

          <div className="flex items-center justify-end gap-3 pt-2">
            <button
              type="submit"
              disabled={save.isPending}
              className="rounded-md bg-white px-5 py-2.5 text-sm font-semibold text-black hover:bg-white/90 disabled:opacity-50"
            >
              {save.isPending ? 'Saving…' : 'Save changes'}
            </button>
          </div>
        </section>

        <aside>
          <div className="mb-3 text-xs font-medium uppercase tracking-wider text-white/70">
            Template
          </div>
          <div className="space-y-2">
            {TEMPLATES.map((t) => (
              <button
                key={t.id}
                type="button"
                onClick={() => setTemplate(t.id)}
                className="block w-full rounded-lg border p-4 text-left transition"
                style={{
                  borderColor: template === t.id ? '#fff' : '#1f1f1f',
                  background: template === t.id ? 'rgba(255,255,255,0.05)' : '#0f0f0f',
                }}
              >
                <div className="flex items-center justify-between">
                  <div className="font-semibold">{t.name}</div>
                  {template === t.id && (
                    <div className="rounded-full bg-white px-2 py-0.5 text-[10px] font-semibold uppercase text-black">
                      Selected
                    </div>
                  )}
                </div>
                <div className="mt-1 text-xs text-white/60">{t.blurb}</div>
              </button>
            ))}
          </div>
        </aside>
      </form>
    </div>
  );
};

const Row: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">{children}</div>
);

const Field: React.FC<{
  label: string;
  required?: boolean;
  hint?: string;
  children: React.ReactNode;
}> = ({ label, required, hint, children }) => (
  <label className="block">
    <div className="mb-1.5 flex items-baseline justify-between">
      <span className="text-xs font-medium uppercase tracking-wider text-white/70">
        {label}
        {required && <span className="text-red-400"> *</span>}
      </span>
      {hint && <span className="text-[10px] text-white/40">{hint}</span>}
    </div>
    {children}
  </label>
);

const Input: React.FC<{
  value: string;
  onChange: (v: string) => void;
  required?: boolean;
  type?: string;
}> = ({ value, onChange, required, type = 'text' }) => (
  <input
    type={type}
    value={value}
    onChange={(e) => onChange(e.target.value)}
    required={required}
    className="w-full rounded-md border bg-black/40 px-3 py-2.5 text-sm outline-none focus:border-white/40"
    style={{ borderColor: '#2a2a2a' }}
  />
);

const Textarea: React.FC<{
  value: string;
  onChange: (v: string) => void;
  rows?: number;
}> = ({ value, onChange, rows = 4 }) => (
  <textarea
    value={value}
    onChange={(e) => onChange(e.target.value)}
    rows={rows}
    className="w-full rounded-md border bg-black/40 px-3 py-2.5 text-sm leading-relaxed outline-none focus:border-white/40"
    style={{ borderColor: '#2a2a2a' }}
  />
);

export default CustomizationAdmin;
