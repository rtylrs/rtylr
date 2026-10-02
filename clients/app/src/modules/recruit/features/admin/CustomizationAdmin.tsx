import AdminQueryState from '../../shared/components/AdminQueryState';
import { RECRUIT_PUBLIC_URL } from '@voxire/config';
import React, { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
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

// Exported so the locale suite can assert a name and a blurb exist for every
// template. Both are looked up with a runtime built key.
export const TEMPLATES: { id: TemplateId }[] = [
  { id: 'minimal' },
  { id: 'editorial' },
  { id: 'bold' },
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
  const { t } = useTranslation('recruitCommon');
  const { data, isLoading, isError, refetch } = useAdminCustomization();
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
      toast.error(err instanceof Error ? err.message : t('recruitAdmin.branding.uploadFailed'));
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
    setNotificationEmail(existing.notificationEmail ?? '');
    setEmailNotifications(existing.emailNotifications ?? true);
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
        onSuccess: () => toast.success(t('recruitAdmin.branding.saved')),
        onError: () => toast.error(t('recruitAdmin.branding.saveFailed')),
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
    toast.success(t('recruitAdmin.branding.paletteReset', { template }));
  }

  if (isLoading) return <AdminQueryState loading />;
  if (isError) return <AdminQueryState retry={() => void refetch()} />;

  return (
    <>
      <HelmetContainer title={t('recruitAdmin.branding.metaTitle')} />
      <div className="flex min-h-screen flex-col">
        {/* Header - title + preview link. Stays pinned above the tabs. */}
        <div
          className="sticky top-0 z-20 border-b px-10 py-6"
          style={{
            borderColor: 'var(--recruit-admin-border)',
            background: 'var(--recruit-admin-bg)',
          }}
        >
          <div className="flex items-baseline justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">
                {t('recruitAdmin.branding.heading')}
              </h1>
              <p className="mt-1 text-sm text-[var(--dash-text-muted)]">
                {t('recruitAdmin.branding.subtitle')} {RECRUIT_PUBLIC_URL}/
                <span className="text-[var(--dash-text)]">
                  {slug || t('recruitAdmin.branding.slugPlaceholder')}
                </span>
              </p>
            </div>
            {slug && (
              <a
                href={`${RECRUIT_PUBLIC_URL}/${slug}`}
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-2 rounded-md border border-[var(--dash-border)] px-3 py-2 text-sm hover:bg-[var(--dash-hover)]"
              >
                <ExternalLink className="h-4 w-4" />
                {t('recruitAdmin.branding.preview')}
              </a>
            )}
          </div>

          {/* Tabs. Icon + label so the admin doesn't have to guess what's
            behind each. */}
          <div className="mt-6 flex gap-1 text-sm">
            <TabButton active={tab === 'brand'} onClick={() => setTab('brand')} icon={<Layout />}>
              {t('recruitAdmin.branding.tabBrand')}
            </TabButton>
            <TabButton
              active={tab === 'palette'}
              onClick={() => setTab('palette')}
              icon={<Palette />}
            >
              {t('recruitAdmin.branding.tabPalette')}
            </TabButton>
            <TabButton
              active={tab === 'typography'}
              onClick={() => setTab('typography')}
              icon={<Type />}
            >
              {t('recruitAdmin.branding.tabTypography')}
            </TabButton>
            <TabButton
              active={tab === 'media'}
              onClick={() => setTab('media')}
              icon={<ImageIcon />}
            >
              {t('recruitAdmin.branding.tabMedia')}
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
                  <Field
                    label={t('recruitAdmin.branding.slug')}
                    required
                    hint={t('recruitAdmin.branding.slugHint')}
                  >
                    <Input value={slug} onChange={setSlug} required />
                  </Field>
                  <Field label={t('recruitAdmin.branding.brandName')} required>
                    <Input value={brandName} onChange={setBrandName} required />
                  </Field>
                </Row>
                <Field label={t('recruitAdmin.branding.tagline')}>
                  <Input value={tagline} onChange={setTagline} />
                </Field>
                <Field
                  label={t('recruitAdmin.branding.about')}
                  hint={t('recruitAdmin.branding.aboutHint')}
                >
                  <Textarea value={aboutBlurb} onChange={setAboutBlurb} rows={4} />
                </Field>
                <Field label={t('recruitAdmin.branding.logoUrl')}>
                  <div className="flex gap-2">
                    <Input value={logoUrl} onChange={setLogoUrl} type="url" />
                    <button
                      type="button"
                      onClick={() => logoInputRef.current?.click()}
                      disabled={uploadingField === 'logo'}
                      className="inline-flex items-center gap-2 rounded-md border border-[var(--dash-border)] px-3 text-sm font-semibold text-[var(--dash-text)]/80 hover:bg-[var(--dash-hover)] disabled:opacity-60"
                    >
                      <Upload size={14} />
                      {uploadingField === 'logo'
                        ? t('recruitAdmin.branding.uploading')
                        : t('recruitAdmin.branding.upload')}
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
                  label={t('recruitAdmin.branding.headerHtml')}
                  hint={t('recruitAdmin.branding.headerHtmlHint')}
                >
                  <Textarea value={headerHtml} onChange={setHeaderHtml} rows={3} />
                </Field>
              </Section>
            )}

            {tab === 'palette' && (
              <Section maxWidth={1000}>
                <div className="flex items-baseline justify-between">
                  <p className="text-sm text-[var(--dash-text-muted)]">
                    Pick a template first - it sets the structural layout. Then tune the palette.
                    Empty fields fall back to the template's default.
                  </p>
                  <button
                    type="button"
                    onClick={resetPalette}
                    className="inline-flex items-center gap-1.5 rounded-md border border-[var(--dash-border)] px-3 py-1.5 text-xs hover:bg-[var(--dash-hover)]"
                  >
                    <RotateCcw className="h-3 w-3" />
                    {t('recruitAdmin.branding.resetPalette')}
                  </button>
                </div>

                <div className="mt-6 grid grid-cols-3 gap-3">
                  {TEMPLATES.map((tpl) => (
                    <button
                      key={tpl.id}
                      type="button"
                      onClick={() => setTemplate(tpl.id)}
                      className="block rounded-lg border p-4 text-left transition"
                      style={{
                        borderColor:
                          template === tpl.id
                            ? 'var(--recruit-admin-selected-border)'
                            : 'var(--recruit-admin-border)',
                        background:
                          template === tpl.id
                            ? 'rgba(255,255,255,0.06)'
                            : 'var(--recruit-admin-surface)',
                      }}
                    >
                      <div className="font-semibold">
                        {t(`recruitAdmin.branding.template.${tpl.id}.name`)}
                      </div>
                      <div className="mt-1 text-xs text-[var(--dash-text-muted)]">
                        {t(`recruitAdmin.branding.template.${tpl.id}.blurb`)}
                      </div>
                    </button>
                  ))}
                </div>

                <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-[1fr_360px]">
                  <div className="space-y-4">
                    <ColorField
                      label={t('recruitAdmin.branding.primary')}
                      value={primaryColor}
                      onChange={setPrimaryColor}
                    />
                    <ColorField
                      label={t('recruitAdmin.branding.accent')}
                      value={accentColor}
                      onChange={setAccentColor}
                      hint={t('recruitAdmin.branding.accentHint')}
                    />
                    <ColorField
                      label={t('recruitAdmin.branding.background')}
                      value={bgColor}
                      onChange={setBgColor}
                      hint={t('recruitAdmin.branding.backgroundHint')}
                    />
                    <ColorField
                      label={t('recruitAdmin.branding.text')}
                      value={textColor}
                      onChange={setTextColor}
                    />
                    <ColorField
                      label={t('recruitAdmin.branding.mutedText')}
                      value={mutedTextColor}
                      onChange={setMutedTextColor}
                      hint={t('recruitAdmin.branding.mutedTextHint')}
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
                <p className="text-sm text-[var(--dash-text-muted)]">
                  {t('recruitAdmin.branding.typographyHint')}
                </p>
                <div className="mt-6 space-y-3">
                  {Object.values(FONT_PAIRINGS).map((p) => (
                    <button
                      key={p.id}
                      type="button"
                      onClick={() => setFontPairing(p.id)}
                      className="block w-full rounded-lg border p-5 text-left transition"
                      style={{
                        borderColor:
                          fontPairing === p.id
                            ? 'var(--recruit-admin-selected-border)'
                            : 'var(--recruit-admin-border)',
                        background:
                          fontPairing === p.id
                            ? 'rgba(255,255,255,0.06)'
                            : 'var(--recruit-admin-surface)',
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <div>
                          <div className="text-xs uppercase tracking-wider text-[var(--dash-text-dim)]">
                            {t(`recruitAdmin.branding.fontPairing.${p.id}`)} ({p.fonts})
                          </div>
                          <div className="mt-2 text-3xl" style={{ fontFamily: p.heading }}>
                            The quick brown fox
                          </div>
                          <div
                            className="mt-1 text-sm text-[var(--dash-text-muted)]"
                            style={{ fontFamily: p.body }}
                          >
                            jumps over the lazy dog.
                          </div>
                        </div>
                        {fontPairing === p.id && (
                          <div className="rounded-full bg-[var(--dash-hover)] px-2 py-0.5 text-[10px] font-semibold uppercase text-[var(--dash-text)]">
                            {t('recruitAdmin.branding.selected')}
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
                  <div className="text-sm font-semibold">
                    {t('recruitAdmin.branding.searchAndSocial')}
                  </div>
                  <div className="text-xs text-[var(--dash-text-dim)]">
                    Override the defaults that LinkedIn, Google, and Twitter show. Blank fields fall
                    back to your brand name + tagline.
                  </div>
                </div>
                <Field
                  label={t('recruitAdmin.branding.metaTitleField')}
                  hint={t('recruitAdmin.branding.metaTitleHint')}
                >
                  <Input value={metaTitle} onChange={setMetaTitle} />
                </Field>
                <Field
                  label={t('recruitAdmin.branding.metaDescription')}
                  hint={t('recruitAdmin.branding.metaDescriptionHint')}
                >
                  <Textarea value={metaDescription} onChange={setMetaDescription} rows={3} />
                </Field>

                <div className="pt-4">
                  <div className="text-sm font-semibold">{t('recruitAdmin.branding.legal')}</div>
                  <div className="text-xs text-[var(--dash-text-dim)]">
                    {t('recruitAdmin.branding.legalHint')}
                  </div>
                </div>
                <Row>
                  <Field label={t('recruitAdmin.branding.privacyUrl')}>
                    <Input value={privacyUrl} onChange={setPrivacyUrl} type="url" />
                  </Field>
                  <Field label={t('recruitAdmin.branding.termsUrl')}>
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
                  <div className="text-sm font-semibold">
                    {t('recruitAdmin.branding.analytics')}
                  </div>
                  <div className="text-xs text-[var(--dash-text-dim)]">
                    {t('recruitAdmin.branding.analyticsHint')}
                  </div>
                </div>
                <Row>
                  <Field
                    label={t('recruitAdmin.branding.ga4')}
                    hint={t('recruitAdmin.branding.ga4Hint')}
                  >
                    <Input value={gaMeasurementId} onChange={setGaMeasurementId} />
                  </Field>
                  <Field
                    label={t('recruitAdmin.branding.metaPixel')}
                    hint={t('recruitAdmin.branding.metaPixelHint')}
                  >
                    <Input value={metaPixelId} onChange={setMetaPixelId} />
                  </Field>
                </Row>

                <div className="pt-4">
                  <div className="text-sm font-semibold">
                    {t('recruitAdmin.branding.emailNotifications')}
                  </div>
                  <div className="text-xs text-[var(--dash-text-dim)]">
                    Candidates always get a confirmation when they apply. You get a "new
                    application" ping at the address below, if set.
                  </div>
                </div>
                <Field
                  label={t('recruitAdmin.branding.notificationInbox')}
                  hint={t('recruitAdmin.branding.notificationInboxHint')}
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
                  {t('recruitAdmin.branding.notifyMe')}
                </label>
              </Section>
            )}

            {tab === 'media' && (
              <Section maxWidth={640}>
                <Field
                  label={t('recruitAdmin.branding.heroImageUrl')}
                  hint={t('recruitAdmin.branding.heroImageHint')}
                >
                  <div className="flex gap-2">
                    <Input value={heroImageUrl} onChange={setHeroImageUrl} type="url" />
                    <button
                      type="button"
                      onClick={() => heroInputRef.current?.click()}
                      disabled={uploadingField === 'hero'}
                      className="inline-flex items-center gap-2 rounded-md border border-[var(--dash-border)] px-3 text-sm font-semibold text-[var(--dash-text)]/80 hover:bg-[var(--dash-hover)] disabled:opacity-60"
                    >
                      <Upload size={14} />
                      {uploadingField === 'hero'
                        ? t('recruitAdmin.branding.uploading')
                        : t('recruitAdmin.branding.upload')}
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
                      alt={t('recruitAdmin.branding.heroPreview')}
                      className="mt-3 h-40 w-full rounded-md object-cover"
                      style={{ background: 'var(--recruit-admin-border)' }}
                    />
                  )}
                </Field>

                <Field
                  label={t('recruitAdmin.branding.faviconUrl')}
                  hint={t('recruitAdmin.branding.faviconHint')}
                >
                  <Input value={faviconUrl} onChange={setFaviconUrl} type="url" />
                  {faviconUrl && (
                    <img
                      src={cdnImageUrl(faviconUrl)}
                      alt={t('recruitAdmin.branding.faviconPreview')}
                      className="mt-3 h-10 w-10 rounded"
                      style={{ background: 'var(--recruit-admin-border)' }}
                    />
                  )}
                </Field>

                {/* OG card preview - a live PNG rendered by the backend
                  from whatever is saved. The admin sees what LinkedIn
                  will scrape when someone shares their URL. */}
                {slug && (
                  <div className="mt-6">
                    <div className="mb-2 text-xs font-medium uppercase tracking-wider text-[var(--dash-text-muted)]">
                      {t('recruitAdmin.branding.socialPreview')}
                    </div>
                    <img
                      key={`${brandName}-${tagline}-${primaryColor}`}
                      src={`${APP_URLS.api}/recruit/public/og?slug=${slug}&_=${Date.now()}`}
                      alt={t('recruitAdmin.branding.ogCard')}
                      className="w-full rounded-md border"
                      style={{ borderColor: 'var(--recruit-admin-border)' }}
                    />
                    <p className="mt-1.5 text-xs text-[var(--dash-text-dim)]">
                      {t('recruitAdmin.branding.ogHint')}
                    </p>
                  </div>
                )}
              </Section>
            )}
          </div>

          <div
            className="sticky bottom-0 z-10 flex items-center justify-end gap-3 border-t px-10 py-4"
            style={{
              borderColor: 'var(--recruit-admin-border)',
              background: 'var(--recruit-admin-bg)',
            }}
          >
            <button
              type="submit"
              disabled={save.isPending}
              className="rounded-md bg-[var(--dash-hover)] px-5 py-2.5 text-sm font-semibold text-[var(--dash-text)] hover:bg-[var(--dash-active)] disabled:opacity-50"
            >
              {save.isPending ? t('recruitAdmin.branding.saving') : t('recruitAdmin.branding.save')}
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
      active
        ? 'bg-[var(--dash-hover)] text-[var(--dash-text)]'
        : 'text-[var(--dash-text-muted)] hover:bg-[var(--dash-hover)] hover:text-[var(--dash-text)]'
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
      <span className="text-xs font-medium uppercase tracking-wider text-[var(--dash-text-muted)]">
        {label}
        {required && <span className="text-red-400"> *</span>}
      </span>
      {hint && <span className="text-[10px] text-[var(--dash-text-dim)]">{hint}</span>}
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
    className="w-full rounded-md border bg-[var(--dash-hover)] px-3 py-2.5 text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--dash-text)]"
    style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
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
    className="w-full rounded-md border bg-[var(--dash-hover)] px-3 py-2.5 text-sm leading-relaxed focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--dash-text)]"
    style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
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
        style={{ borderColor: 'var(--recruit-admin-border-strong)' }}
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
}> = ({ primary, accent, bg, fg, muted, brandName }) => {
  const { t } = useTranslation('recruitCommon');
  // Sample content, translated for the same reason the rest of the screen is:
  // an admin judging their colours should read it in their own language.
  const roles = [
    t('recruitAdmin.branding.previewCard.role1'),
    t('recruitAdmin.branding.previewCard.role2'),
    t('recruitAdmin.branding.previewCard.role3'),
  ];
  return (
    <div
      className="rounded-lg border p-5 shadow-[0_0_0_1px_rgba(255,255,255,0.04)]"
      style={{ borderColor: 'var(--recruit-admin-border)', background: bg }}
    >
      <div className="text-[10px] uppercase tracking-[0.3em]" style={{ color: accent }}>
        {t('recruitAdmin.branding.previewCard.careersEyebrow')} · {brandName}
      </div>
      <div
        className="mt-3 text-2xl font-semibold leading-tight"
        style={{ color: fg, fontFamily: 'var(--recruit-font-display)' }}
      >
        {t('recruitAdmin.branding.previewCard.headline')}
      </div>
      <div className="mt-2 text-sm" style={{ color: muted }}>
        {t('recruitAdmin.branding.previewCard.subline')}
      </div>
      <div className="mt-4 space-y-1.5">
        {roles.map((role) => (
          <div
            key={role}
            className="flex items-center justify-between rounded-md px-3 py-2 text-xs"
            style={{ background: 'rgba(0,0,0,0.04)', color: fg }}
          >
            <span>{role}</span>
            <span
              className="inline-block h-1.5 w-1.5 rounded-full"
              style={{ background: primary }}
            />
          </div>
        ))}
      </div>
    </div>
  );
};

export default CustomizationAdmin;
