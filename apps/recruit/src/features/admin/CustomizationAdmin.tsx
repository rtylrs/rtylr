import React, { useEffect, useRef, useState } from 'react';
import { cn } from '@voxire/ui';
import { toast } from 'sonner';
import { uploadAsset } from '@voxire/api';
import { APP_URLS, cdnImageUrl } from '@voxire/config';
import {
  ExternalLink,
  RotateCcw,
  Palette,
  Type,
  ImageIcon,
  Layout,
  Search,
  Upload,
} from 'lucide-react';
import { HelmetContainer } from '@voxire/components';

import { useAdminCustomization, useUpsertCustomization } from '../../shared/hooks/queries/useAdmin';
import type { TemplateId, FontPairingId } from '../../shared/lib/types';
import { FONT_PAIRINGS } from '../../templates/fonts';

/**
 * CustomizationAdmin - edits every knob a business can turn on their
 * careers page. Laid out as four tabbed sections so the form isn't a
 * 20-field wall:
 *
 *   Brand      - slug, name, tagline, about, logo, favicon, header HTML
 *   Palette    - template pick + five color fields with a live preview
 *   Typography - font-pairing picker with sample rendering
 *   Media      - hero image for bold/editorial templates
 *
 * The Save button lives in a sticky footer so it stays reachable
 * regardless of which tab is open.
 */

const TEMPLATES: { id: TemplateId; name: string; blurb: string }[] = [
  {
    id: 'minimal',
    name: 'Minimal',
    blurb: 'Calm single-column list. Default for every new page.',
  },
  {
    id: 'editorial',
    name: 'Editorial',
    blurb: 'Magazine-style hero, serif display, generous whitespace.',
  },
  {
    id: 'bold',
    name: 'Bold',
    blurb: 'Saturated hero block, oversized type. Needs a strong color.',
  },
];

// TEMPLATE_DEFAULTS mirrors the defaults in templates/resolver.tsx so
// "Reset palette" can give the admin a sensible starting point that
// actually matches what the template expects.
const TEMPLATE_DEFAULTS: Record<TemplateId, Record<string, string>> = {
  minimal: {
    primary: '#111111',
    accent: '#111111',
    bg: '#fafaf9',
    fg: '#18181b',
    muted: '#6b7280',
  },
  editorial: {
    primary: '#d94f2f',
    accent: '#111827',
    bg: '#ffffff',
    fg: '#111827',
    muted: '#6b7280',
  },
  bold: {
    primary: '#1d4ed8',
    accent: '#0f172a',
    bg: '#ffffff',
    fg: '#0f172a',
    muted: '#64748b',
  },
};

type Tab = 'brand' | 'palette' | 'typography' | 'media' | 'seo';

const CustomizationAdmin: React.FC = (): React.JSX.Element => {
  const { data } = useAdminCustomization();
  const save = useUpsertCustomization();

  const existing = data?.customization;
  const [tab, setTab] = useState<Tab>('brand');

  const [slug, setSlug] = useState(existing?.slug ?? '');
  const [brandName, setBrandName] = useState(existing?.brandName ?? '');
  const [tagline, setTagline] = useState(existing?.tagline ?? '');
  const [aboutBlurb, setAboutBlurb] = useState(existing?.aboutBlurb ?? '');
  const [logoUrl, setLogoUrl] = useState(existing?.logoUrl ?? '');
  const [primaryColor, setPrimaryColor] = useState(existing?.primaryColor ?? '#111111');
  const [accentColor, setAccentColor] = useState(existing?.accentColor ?? '');
  const [bgColor, setBgColor] = useState(existing?.bgColor ?? '');
  const [textColor, setTextColor] = useState(existing?.textColor ?? '');
  const [mutedTextColor, setMutedTextColor] = useState(existing?.mutedTextColor ?? '');
  const [fontPairing, setFontPairing] = useState<FontPairingId>(
    (existing?.fontPairing as FontPairingId) ?? 'default'
  );
  const [heroImageUrl, setHeroImageUrl] = useState(existing?.heroImageUrl ?? '');
  const [faviconUrl, setFaviconUrl] = useState(existing?.faviconUrl ?? '');
  const [metaTitle, setMetaTitle] = useState(existing?.metaTitle ?? '');
  const [metaDescription, setMetaDescription] = useState(existing?.metaDescription ?? '');
  const [privacyUrl, setPrivacyUrl] = useState(existing?.privacyUrl ?? '');
  const [termsUrl, setTermsUrl] = useState(existing?.termsUrl ?? '');
  const [gaMeasurementId, setGaMeasurementId] = useState(existing?.gaMeasurementId ?? '');
  const [metaPixelId, setMetaPixelId] = useState(existing?.metaPixelId ?? '');
  const [requireConsent, setRequireConsent] = useState(existing?.requireConsent ?? false);
  const [notificationEmail, setNotificationEmail] = useState(existing?.notificationEmail ?? '');
  const [emailNotifications, setEmailNotifications] = useState(
    existing?.emailNotifications ?? true
  );
  const [template, setTemplate] = useState<TemplateId>(
    (existing?.template as TemplateId) ?? 'minimal'
  );
  const [headerHtml, setHeaderHtml] = useState(existing?.headerHtml ?? '');
  const [uploadingField, setUploadingField] = useState<'logo' | 'hero' | null>(null);
  const logoInputRef = useRef<HTMLInputElement>(null);
  const heroInputRef = useRef<HTMLInputElement>(null);

  const uploadImage = async (
    field: 'logo' | 'hero',
    file: File | undefined | null
  ): Promise<void> => {
    if (!file) return;
    setUploadingField(field);
    try {
      const uploaded = await uploadAsset(file);
      if (field === 'logo') {
        setLogoUrl(uploaded.url);
      } else {
        setHeroImageUrl(uploaded.url);
      }
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Upload failed');
    } finally {
      setUploadingField(null);
      if (field === 'logo' && logoInputRef.current) logoInputRef.current.value = '';
      if (field === 'hero' && heroInputRef.current) heroInputRef.current.value = '';
    }
  };

  useEffect(() => {
    if (!existing) return;
    setSlug(existing.slug);
    setBrandName(existing.brandName);
    setTagline(existing.tagline);
    setAboutBlurb(existing.aboutBlurb);
    setLogoUrl(existing.logoUrl);
    setPrimaryColor(existing.primaryColor || '#111111');
    setAccentColor(existing.accentColor);
    setBgColor(existing.bgColor);
    setTextColor(existing.textColor);
    setMutedTextColor(existing.mutedTextColor);
    setFontPairing((existing.fontPairing as FontPairingId) ?? 'default');
    setHeroImageUrl(existing.heroImageUrl);
    setFaviconUrl(existing.faviconUrl);
    setTemplate((existing.template as TemplateId) ?? 'minimal');
    setHeaderHtml(existing.headerHtml);
    setMetaTitle(existing.metaTitle);
    setMetaDescription(existing.metaDescription);
    setPrivacyUrl(existing.privacyUrl);
    setTermsUrl(existing.termsUrl);
    setGaMeasurementId(existing.gaMeasurementId);
    setMetaPixelId(existing.metaPixelId);
    setRequireConsent(existing.requireConsent);
    setNotificationEmail(existing.notificationEmail);
    setEmailNotifications(existing.emailNotifications);
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
        accentColor,
        bgColor,
        textColor,
        mutedTextColor,
        fontPairing,
        heroImageUrl,
        faviconUrl,
        template,
        headerHtml,
        metaTitle,
        metaDescription,
        privacyUrl,
        termsUrl,
        gaMeasurementId,
        metaPixelId,
        requireConsent,
        notificationEmail,
        emailNotifications,
      },
      {
        onSuccess: () => toast.success('Careers page saved'),
        onError: () => toast.error("Couldn't save - check your slug is unique"),
      }
    );
  };

  function resetPalette() {
    const d = TEMPLATE_DEFAULTS[template];
    setPrimaryColor(d.primary);
    setAccentColor(d.accent);
    setBgColor(d.bg);
    setTextColor(d.fg);
    setMutedTextColor(d.muted);
    toast.success(`Palette reset to ${template} defaults`);
  }

  return (
    <>
      <HelmetContainer title="Page branding · rtylr" />
      <div className="flex min-h-screen flex-col">
        {/* Header - title + preview link. Stays pinned above the tabs. */}
        <div
          className="sticky top-0 z-20 border-b px-10 py-6"
          style={{ borderColor: '#1f1f1f', background: '#0a0a0a' }}
        >
          <div className="flex items-baseline justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Page branding</h1>
              <p className="mt-1 text-sm text-white/60">
                What candidates see at {APP_URLS.recruit}/
                <span className="text-white">{slug || 'your-slug'}</span>
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

          {/* Tabs. Icon + label so the admin doesn't have to guess what's
            behind each. */}
          <div className="mt-6 flex gap-1 text-sm">
            <TabButton active={tab === 'brand'} onClick={() => setTab('brand')} icon={<Layout />}>
              Brand
            </TabButton>
            <TabButton
              active={tab === 'palette'}
              onClick={() => setTab('palette')}
              icon={<Palette />}
            >
              Palette
            </TabButton>
            <TabButton
              active={tab === 'typography'}
              onClick={() => setTab('typography')}
              icon={<Type />}
            >
              Typography
            </TabButton>
            <TabButton
              active={tab === 'media'}
              onClick={() => setTab('media')}
              icon={<ImageIcon />}
            >
              Media
            </TabButton>
            <TabButton active={tab === 'seo'} onClick={() => setTab('seo')} icon={<Search />}>
              SEO &amp; legal
            </TabButton>
          </div>
        </div>

        <form onSubmit={onSubmit} className="flex-1">
          <div className="px-10 py-8">
            {tab === 'brand' && (
              <Section maxWidth={640}>
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
                <Field label="Logo URL">
                  <div className="flex gap-2">
                    <Input value={logoUrl} onChange={setLogoUrl} type="url" />
                    <button
                      type="button"
                      onClick={() => logoInputRef.current?.click()}
                      disabled={uploadingField === 'logo'}
                      className="inline-flex items-center gap-2 rounded-md border border-white/10 px-3 text-sm font-semibold text-white/80 hover:bg-white/5 disabled:opacity-60"
                    >
                      <Upload size={14} />
                      {uploadingField === 'logo' ? 'Uploading...' : 'Upload'}
                    </button>
                    <input
                      ref={logoInputRef}
                      type="file"
                      accept="image/*"
                      className="hidden"
                      onChange={(e) => {
                        void uploadImage('logo', e.target.files?.[0]);
                      }}
                    />
                  </div>
                </Field>
                <Field
                  label="Custom header HTML"
                  hint="Advanced - raw HTML strip above the brand header"
                >
                  <Textarea value={headerHtml} onChange={setHeaderHtml} rows={3} />
                </Field>
              </Section>
            )}

            {tab === 'palette' && (
              <Section maxWidth={1000}>
                <div className="flex items-baseline justify-between">
                  <p className="text-sm text-white/60">
                    Pick a template first - it sets the structural layout. Then tune the palette.
                    Empty fields fall back to the template's default.
                  </p>
                  <button
                    type="button"
                    onClick={resetPalette}
                    className="inline-flex items-center gap-1.5 rounded-md border border-white/10 px-3 py-1.5 text-xs hover:bg-white/5"
                  >
                    <RotateCcw className="h-3 w-3" />
                    Reset to template defaults
                  </button>
                </div>

                <div className="mt-6 grid grid-cols-3 gap-3">
                  {TEMPLATES.map((t) => (
                    <button
                      key={t.id}
                      type="button"
                      onClick={() => setTemplate(t.id)}
                      className="block rounded-lg border p-4 text-left transition"
                      style={{
                        borderColor: template === t.id ? '#fff' : '#1f1f1f',
                        background: template === t.id ? 'rgba(255,255,255,0.06)' : '#0f0f0f',
                      }}
                    >
                      <div className="font-semibold">{t.name}</div>
                      <div className="mt-1 text-xs text-white/60">{t.blurb}</div>
                    </button>
                  ))}
                </div>

                <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-[1fr_360px]">
                  <div className="space-y-4">
                    <ColorField label="Primary" value={primaryColor} onChange={setPrimaryColor} />
                    <ColorField
                      label="Accent"
                      value={accentColor}
                      onChange={setAccentColor}
                      hint="Secondary highlight. Empty → falls back to Primary"
                    />
                    <ColorField
                      label="Background"
                      value={bgColor}
                      onChange={setBgColor}
                      hint="Page canvas"
                    />
                    <ColorField label="Text" value={textColor} onChange={setTextColor} />
                    <ColorField
                      label="Muted text"
                      value={mutedTextColor}
                      onChange={setMutedTextColor}
                      hint="Captions, metadata"
                    />
                  </div>

                  {/* Live preview - shows a miniature hero so the admin
                    can see the palette land without navigating away. */}
                  <PalettePreview
                    primary={primaryColor || '#111111'}
                    accent={accentColor || primaryColor || '#111111'}
                    bg={bgColor || TEMPLATE_DEFAULTS[template].bg}
                    fg={textColor || TEMPLATE_DEFAULTS[template].fg}
                    muted={mutedTextColor || TEMPLATE_DEFAULTS[template].muted}
                    brandName={brandName || 'Your brand'}
                  />
                </div>
              </Section>
            )}

            {tab === 'typography' && (
              <Section maxWidth={720}>
                <p className="text-sm text-white/60">
                  Five curated pairings. Fonts load on demand from Google Fonts - picking one
                  doesn't block the save.
                </p>
                <div className="mt-6 space-y-3">
                  {Object.values(FONT_PAIRINGS).map((p) => (
                    <button
                      key={p.id}
                      type="button"
                      onClick={() => setFontPairing(p.id)}
                      className="block w-full rounded-lg border p-5 text-left transition"
                      style={{
                        borderColor: fontPairing === p.id ? '#fff' : '#1f1f1f',
                        background: fontPairing === p.id ? 'rgba(255,255,255,0.06)' : '#0f0f0f',
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <div>
                          <div className="text-xs uppercase tracking-wider text-white/50">
                            {p.label}
                          </div>
                          <div className="mt-2 text-3xl" style={{ fontFamily: p.heading }}>
                            The quick brown fox
                          </div>
                          <div
                            className="mt-1 text-sm text-white/70"
                            style={{ fontFamily: p.body }}
                          >
                            jumps over the lazy dog.
                          </div>
                        </div>
                        {fontPairing === p.id && (
                          <div className="rounded-full bg-white px-2 py-0.5 text-[10px] font-semibold uppercase text-black">
                            Selected
                          </div>
                        )}
                      </div>
                    </button>
                  ))}
                </div>
              </Section>
            )}

            {tab === 'seo' && (
              <Section maxWidth={640}>
                <div>
                  <div className="text-sm font-semibold">Search &amp; social</div>
                  <div className="text-xs text-white/50">
                    Override the defaults that LinkedIn, Google, and Twitter show. Blank fields fall
                    back to your brand name + tagline.
                  </div>
                </div>
                <Field label="Meta title" hint="Overrides browser tab + search result title">
                  <Input value={metaTitle} onChange={setMetaTitle} />
                </Field>
                <Field label="Meta description" hint="1–2 sentences, under 160 chars">
                  <Textarea value={metaDescription} onChange={setMetaDescription} rows={3} />
                </Field>

                <div className="pt-4">
                  <div className="text-sm font-semibold">Legal</div>
                  <div className="text-xs text-white/50">
                    Links shown on the apply form. Useful for GDPR / EEO disclosures.
                  </div>
                </div>
                <Row>
                  <Field label="Privacy policy URL">
                    <Input value={privacyUrl} onChange={setPrivacyUrl} type="url" />
                  </Field>
                  <Field label="Terms URL">
                    <Input value={termsUrl} onChange={setTermsUrl} type="url" />
                  </Field>
                </Row>
                <label className="flex items-center gap-2 pt-1 text-sm">
                  <input
                    type="checkbox"
                    checked={requireConsent}
                    onChange={(e) => setRequireConsent(e.target.checked)}
                    className="h-4 w-4"
                  />
                  Require explicit consent checkbox on the apply form (GDPR)
                </label>

                <div className="pt-4">
                  <div className="text-sm font-semibold">Analytics</div>
                  <div className="text-xs text-white/50">
                    Only loaded when an ID is set. Leave blank to ship zero tracking.
                  </div>
                </div>
                <Row>
                  <Field label="GA4 measurement ID" hint="Starts with G-">
                    <Input value={gaMeasurementId} onChange={setGaMeasurementId} />
                  </Field>
                  <Field label="Meta Pixel ID" hint="Numeric">
                    <Input value={metaPixelId} onChange={setMetaPixelId} />
                  </Field>
                </Row>

                <div className="pt-4">
                  <div className="text-sm font-semibold">Email notifications</div>
                  <div className="text-xs text-white/50">
                    Candidates always get a confirmation when they apply. You get a "new
                    application" ping at the address below, if set.
                  </div>
                </div>
                <Field
                  label="HR notification inbox"
                  hint="Team alias works best (e.g. jobs@acme.com)"
                >
                  <Input value={notificationEmail} onChange={setNotificationEmail} type="email" />
                </Field>
                <label className="flex items-center gap-2 pt-1 text-sm">
                  <input
                    type="checkbox"
                    checked={emailNotifications}
                    onChange={(e) => setEmailNotifications(e.target.checked)}
                    className="h-4 w-4"
                  />
                  Send me email notifications for new applications
                </label>
              </Section>
            )}

            {tab === 'media' && (
              <Section maxWidth={640}>
                <Field
                  label="Hero image URL"
                  hint="Shown in the bold / editorial templates above the job list"
                >
                  <div className="flex gap-2">
                    <Input value={heroImageUrl} onChange={setHeroImageUrl} type="url" />
                    <button
                      type="button"
                      onClick={() => heroInputRef.current?.click()}
                      disabled={uploadingField === 'hero'}
                      className="inline-flex items-center gap-2 rounded-md border border-white/10 px-3 text-sm font-semibold text-white/80 hover:bg-white/5 disabled:opacity-60"
                    >
                      <Upload size={14} />
                      {uploadingField === 'hero' ? 'Uploading...' : 'Upload'}
                    </button>
                    <input
                      ref={heroInputRef}
                      type="file"
                      accept="image/*"
                      className="hidden"
                      onChange={(e) => {
                        void uploadImage('hero', e.target.files?.[0]);
                      }}
                    />
                  </div>
                  {heroImageUrl && (
                    <img
                      src={cdnImageUrl(heroImageUrl)}
                      alt="Hero preview"
                      className="mt-3 h-40 w-full rounded-md object-cover"
                      style={{ background: '#1f1f1f' }}
                    />
                  )}
                </Field>

                <Field label="Favicon URL" hint="Square image, ≥32px. Shown in browser tabs.">
                  <Input value={faviconUrl} onChange={setFaviconUrl} type="url" />
                  {faviconUrl && (
                    <img
                      src={cdnImageUrl(faviconUrl)}
                      alt="Favicon preview"
                      className="mt-3 h-10 w-10 rounded"
                      style={{ background: '#1f1f1f' }}
                    />
                  )}
                </Field>

                {/* OG card preview - a live PNG rendered by the backend
                  from whatever is saved. The admin sees what LinkedIn
                  will scrape when someone shares their URL. */}
                {slug && (
                  <div className="mt-6">
                    <div className="mb-2 text-xs font-medium uppercase tracking-wider text-white/70">
                      Social preview
                    </div>
                    <img
                      key={`${brandName}-${tagline}-${primaryColor}`}
                      src={`${APP_URLS.api}/recruit/public/og?slug=${slug}&_=${Date.now()}`}
                      alt="OG card"
                      className="w-full rounded-md border"
                      style={{ borderColor: '#1f1f1f' }}
                    />
                    <p className="mt-1.5 text-xs text-white/40">
                      Refreshes after you save. This is the card LinkedIn, Slack, and iMessage use.
                    </p>
                  </div>
                )}
              </Section>
            )}
          </div>

          <div
            className="sticky bottom-0 z-10 flex items-center justify-end gap-3 border-t px-10 py-4"
            style={{ borderColor: '#1f1f1f', background: '#0a0a0a' }}
          >
            <button
              type="submit"
              disabled={save.isPending}
              className="rounded-md bg-white px-5 py-2.5 text-sm font-semibold text-black hover:bg-white/90 disabled:opacity-50"
            >
              {save.isPending ? 'Saving…' : 'Save changes'}
            </button>
          </div>
        </form>
      </div>
    </>
  );
};

// ── small building-block components ───────────────────────────────

const TabButton: React.FC<{
  active: boolean;
  onClick: () => void;
  icon: React.ReactNode;
  children: React.ReactNode;
}> = ({ active, onClick, icon, children }) => (
  <button
    type="button"
    onClick={onClick}
    className={cn(
      'inline-flex items-center gap-2 rounded-md px-3 py-1.5 transition',
      active ? 'bg-white/10 text-white' : 'text-white/60 hover:bg-white/5 hover:text-white'
    )}
  >
    <span className="[&>svg]:h-4 [&>svg]:w-4">{icon}</span>
    {children}
  </button>
);

const Section: React.FC<{ children: React.ReactNode; maxWidth?: number }> = ({
  children,
  maxWidth = 640,
}) => (
  <div style={{ maxWidth }} className="space-y-5">
    {children}
  </div>
);

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

const ColorField: React.FC<{
  label: string;
  value: string;
  onChange: (v: string) => void;
  hint?: string;
}> = ({ label, value, onChange, hint }) => (
  <Field label={label} hint={hint}>
    <div className="flex gap-2">
      <input
        type="color"
        value={value || '#000000'}
        onChange={(e) => onChange(e.target.value)}
        className="h-10 w-16 rounded-md border"
        style={{ borderColor: '#2a2a2a' }}
      />
      <Input value={value} onChange={onChange} />
    </div>
  </Field>
);

/**
 * PalettePreview renders a tiny faux careers-page hero using the live
 * form values. Not pixel-accurate to any one template - the point is
 * showing the admin whether their colors read together before they hit
 * save and alt-tab to their public page.
 */
const PalettePreview: React.FC<{
  primary: string;
  accent: string;
  bg: string;
  fg: string;
  muted: string;
  brandName: string;
}> = ({ primary, accent, bg, fg, muted, brandName }) => (
  <div
    className="rounded-lg border p-5 shadow-[0_0_0_1px_rgba(255,255,255,0.04)]"
    style={{ borderColor: '#1f1f1f', background: bg }}
  >
    <div className="text-[10px] uppercase tracking-[0.3em]" style={{ color: accent }}>
      Careers · {brandName}
    </div>
    <div
      className="mt-3 text-2xl font-semibold leading-tight"
      style={{ color: fg, fontFamily: 'var(--recruit-font-display)' }}
    >
      Come build the future of small-business software.
    </div>
    <div className="mt-2 text-sm" style={{ color: muted }}>
      Remote-first · Competitive pay · Team you actually want to work with
    </div>
    <div className="mt-4 space-y-1.5">
      {['Senior Frontend Engineer', 'Product Designer', 'Growth Manager'].map((role) => (
        <div
          key={role}
          className="flex items-center justify-between rounded-md px-3 py-2 text-xs"
          style={{ background: 'rgba(0,0,0,0.04)', color: fg }}
        >
          <span>{role}</span>
          <span className="inline-block h-1.5 w-1.5 rounded-full" style={{ background: primary }} />
        </div>
      ))}
    </div>
  </div>
);

export default CustomizationAdmin;
