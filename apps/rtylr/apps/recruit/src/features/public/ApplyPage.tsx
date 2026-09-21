import React, { useEffect, useRef, useState } from 'react';
import { Link, useNavigate, useParams } from '@tanstack/react-router';
import { toast } from 'sonner';
import { Trans, useTranslation } from 'react-i18next';
import { ArrowLeft, Check, Upload, FileText, X } from 'lucide-react';

import Header from '../../templates/shared/Header';
import {
  usePublicCareer,
  useSubmitApplication,
  uploadResume,
} from '../../shared/hooks/queries/usePublic';
import { applyBrand } from '../../templates/resolver';
import { applySEO, loadAnalytics } from '../../shared/lib/seo';
import { FLAG_KEYS, resolveFlag } from '@voxire/config';
import NotFound from './NotFound';
import { cleanMojibake } from '../../shared/lib/sanitize';

/**
 * ApplyPage - form at recruit.rtylr.com/:slug/:careerSlug/apply.
 *
 * MVP scope: name, email, phone, resume URL (linked from wherever the
 * candidate already hosts it - Drive, Dropbox, personal site), cover
 * letter. Native multipart resume upload is a follow-up; storing a URL
 * keeps the schema stable across that change.
 *
 * The form renders the shared Header so branding carries across the
 * apply flow. It intentionally doesn't route through the template
 * resolver - templates diverge on marketing surfaces, but the apply
 * form is a utility and benefits from a single consistent shape.
 */
/**
 * The consent row reads differently depending on which legal links the
 * business has set, and each variant is a complete sentence rather than a
 * fragment, so a translator can place the links where the language wants them.
 */
function consentKey(privacyUrl?: string, termsUrl?: string): string {
  if (privacyUrl && termsUrl) return 'public.apply.consentWithBoth';
  if (privacyUrl) return 'public.apply.consentWithPrivacy';
  if (termsUrl) return 'public.apply.consentWithTerms';
  return 'public.apply.consent';
}

const MAX_ANSWER_LENGTH = 5000;

const ApplyPage: React.FC = (): React.JSX.Element => {
  const { slug: rawSlug, careerSlug: rawCareerSlug } = useParams({ strict: false }) as {
    slug?: string;
    careerSlug?: string;
  };
  const slug = rawSlug ?? '';
  const careerSlug = rawCareerSlug ?? '';
  const navigate = useNavigate();
  const { t } = useTranslation();

  const isFileUploadEnabled = resolveFlag(FLAG_KEYS.APP_RECRUIT_FILE_UPLOAD);

  const { data, isLoading, error } = usePublicCareer(slug, careerSlug);
  const submit = useSubmitApplication();

  useEffect(() => {
    if (!data) return;
    applyBrand(data.customization);
    applySEO(data.customization, { pageKind: 'career', careerTitle: data.career.title });
    loadAnalytics(data.customization);
  }, [data]);

  const [candidateName, setCandidateName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [resumeUrl, setResumeUrl] = useState('');
  const [coverLetter, setCoverLetter] = useState('');
  const [done, setDone] = useState(false);

  const [answers, setAnswers] = useState<Record<string, string | string[]>>({});
  function setAnswer(fieldId: number, value: string | string[]) {
    setAnswers((prev) => ({ ...prev, [String(fieldId)]: value }));
  }

  const [consented, setConsented] = useState(false);
  const [honeypot, setHoneypot] = useState('');

  const [resumeFile, setResumeFile] = useState<File | null>(null);
  const [uploadedKey, setUploadedKey] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  // Bumped on every pick/clear so an upload that was removed (or replaced)
  // can't write its key into state after the candidate's decision.
  const uploadEpochRef = useRef(0);

  if (!slug || !careerSlug) {
    return <div>{t('public.apply.invalidLink')}</div>;
  }

  async function onPickFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    const allowed = [
      'application/pdf',
      'application/msword',
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    ];
    if (!allowed.includes(file.type)) {
      setUploadError(t('public.apply.fileType'));
      if (fileInputRef.current) fileInputRef.current.value = '';
      return;
    }
    setResumeFile(file);
    setUploadedKey(null);
    setUploadError(null);
    setUploading(true);
    const epoch = ++uploadEpochRef.current;
    try {
      const key = await uploadResume({ file, slug, careerSlug });
      if (uploadEpochRef.current !== epoch) return;
      setUploadedKey(key);
      // Clear the manual URL input - we'll use the uploaded key instead.
      setResumeUrl('');
    } catch (err) {
      if (uploadEpochRef.current !== epoch) return;
      const msg = err instanceof Error ? err.message : t('public.apply.uploadFailed');
      setUploadError(msg);
      setResumeFile(null);
      if (fileInputRef.current) fileInputRef.current.value = '';
    } finally {
      if (uploadEpochRef.current === epoch) setUploading(false);
    }
  }

  function clearFile() {
    // Invalidate a still-in-flight upload so its resolved key is dropped.
    uploadEpochRef.current += 1;
    setResumeFile(null);
    setUploadedKey(null);
    setUploadError(null);
    setUploading(false);
    if (fileInputRef.current) fileInputRef.current.value = '';
  }

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div
          className="h-6 w-6 animate-spin rounded-full border-2"
          style={{ borderColor: 'rgba(0,0,0,0.1)', borderTopColor: 'var(--recruit-accent)' }}
        />
      </div>
    );
  }
  if (error && !data) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-white p-6 text-center">
        <h2 className="text-lg font-semibold text-gray-900">{t('public.apply.loadFailed')}</h2>
        <p className="text-sm text-gray-500">{t('public.errors.refreshHint')}</p>
        <button
          type="button"
          onClick={() => window.location.reload()}
          className="rounded-lg bg-gray-900 px-4 py-2 text-sm text-white hover:bg-gray-800"
        >
          {t('public.errors.retry')}
        </button>
      </div>
    );
  }

  if (!data) return <NotFound />;

  const { career, customization } = data;
  const customFields = data.fields ?? [];

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (honeypot.trim() !== '') {
      setDone(true);
      return;
    }
    // Client-side required check before hitting the API - better UX
    // than a round-trip 400. The backend still accepts whatever it's
    // given since we can't yet validate kind-specific rules
    // server-side.
    for (const f of customFields) {
      const v = answers[String(f.id)];
      const empty = v === undefined || v === '' || (Array.isArray(v) && v.length === 0);
      if (f.required && empty) {
        toast.error(t('public.apply.required', { label: f.label }));
        return;
      }
      if (typeof v === 'string' && v.length > MAX_ANSWER_LENGTH) {
        toast.error(t('public.apply.tooLong', { label: f.label, max: MAX_ANSWER_LENGTH }));
        return;
      }
    }

    if (customization.requireConsent && !consented) {
      toast.error(t('public.apply.consentRequired'));
      return;
    }

    // Prefer the uploaded-to-R2 key; fall back to the manually typed
    // URL for candidates who'd rather link out (Drive, personal site).
    const resumeRef = uploadedKey ?? (resumeUrl || undefined);
    const answersJson = Object.keys(answers).length > 0 ? JSON.stringify(answers) : undefined;

    submit.mutate(
      {
        slug,
        careerSlug,
        candidateName,
        email,
        phone: phone || undefined,
        resumeUrl: resumeRef,
        coverLetter: coverLetter || undefined,
        answersJson,
        consented,
      },
      {
        onSuccess: () => {
          setDone(true);
          toast.success(t('public.apply.received'));
          window.history.replaceState(null, '', window.location.href);
        },
        onError: () => toast.error(t('public.apply.submitFailed')),
      }
    );
  };

  if (done) {
    return (
      <div className="min-h-screen">
        <Header customization={customization} compact />
        <main className="mx-auto max-w-xl px-6 py-24 text-center">
          <div
            className="mx-auto flex h-12 w-12 items-center justify-center rounded-full"
            style={{ background: 'var(--recruit-accent)' }}
          >
            <Check className="h-6 w-6 text-white" />
          </div>
          <h1 className="mt-6 text-3xl font-semibold tracking-tight">
            {t('public.apply.doneTitle')}
          </h1>
          <p className="mt-4 text-base" style={{ color: 'var(--recruit-muted)' }}>
            <Trans
              i18nKey="public.apply.doneBody"
              values={{ title: career.title, brand: customization.brandName }}
              components={{ name: <span className="font-medium" /> }}
            />
          </p>
          <button
            type="button"
            onClick={() => navigate({ to: '/$slug', params: { slug } })}
            className="mt-8 inline-flex items-center gap-2 rounded-lg border px-5 py-2.5 text-sm font-medium"
            style={{ borderColor: 'var(--recruit-border)' }}
          >
            <ArrowLeft className="h-4 w-4" />
            {t('public.apply.backToRoles')}
          </button>
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen">
      <Header customization={customization} compact />
      <main className="mx-auto max-w-xl px-6 py-12">
        <Link
          to="/$slug/$careerSlug"
          params={{ slug, careerSlug }}
          className="inline-flex min-h-[44px] items-center gap-2 text-sm font-medium hover:underline"
          style={{ color: 'var(--recruit-muted)' }}
        >
          <ArrowLeft className="h-4 w-4" />
          {t('public.apply.backToRole')}
        </Link>

        <h1 className="mt-6 text-3xl font-semibold tracking-tight">
          {t('public.apply.heading', { title: cleanMojibake(career.title) })}
        </h1>
        <p className="mt-2 text-sm" style={{ color: 'var(--recruit-muted)' }}>
          {cleanMojibake(customization.brandName)} ·{' '}
          {cleanMojibake(career.location) || t('public.apply.remote')}
        </p>

        <form onSubmit={onSubmit} className="mt-8 space-y-5">
          <div style={{ display: 'none' }} aria-hidden="true">
            <input
              type="text"
              name="website_url_hp"
              value={honeypot}
              onChange={(e) => setHoneypot(e.target.value)}
              tabIndex={-1}
              autoComplete="off"
            />
          </div>
          <Field label={t('public.apply.fullName')} required>
            <input
              required
              value={candidateName}
              onChange={(e) => setCandidateName(e.target.value)}
              className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
              style={{ borderColor: 'var(--recruit-border)' }}
            />
          </Field>
          <Field label={t('public.apply.email')} required>
            <input
              required
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
              style={{ borderColor: 'var(--recruit-border)' }}
            />
          </Field>
          <Field label={t('public.apply.phone')}>
            <input
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
              style={{ borderColor: 'var(--recruit-border)' }}
            />
          </Field>
          {isFileUploadEnabled && (
            <Field label={t('public.apply.resume')}>
              {/* Hidden native input; we render a styled dropzone ourselves. */}
              <input
                ref={fileInputRef}
                type="file"
                accept=".pdf,.doc,.docx,application/pdf,application/msword,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
                onChange={onPickFile}
                className="hidden"
              />
              {!resumeFile && !uploadedKey && (
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  className="flex w-full items-center justify-center gap-2 rounded-md border border-dashed bg-white px-4 py-6 text-sm transition hover:bg-gray-50"
                  style={{ borderColor: 'var(--recruit-border)' }}
                >
                  <Upload className="h-4 w-4" style={{ color: 'var(--recruit-accent)' }} />
                  <span>{t('public.apply.uploadCta')}</span>
                  <span className="text-xs" style={{ color: 'var(--recruit-muted)' }}>
                    {t('public.apply.uploadHint')}
                  </span>
                </button>
              )}
              {resumeFile && (
                <div
                  className="flex items-center gap-3 rounded-md border bg-white px-3 py-2.5 text-sm"
                  style={{ borderColor: 'var(--recruit-border)' }}
                >
                  <FileText className="h-4 w-4" style={{ color: 'var(--recruit-accent)' }} />
                  <div className="min-w-0 flex-1">
                    <div className="truncate font-medium">{resumeFile.name}</div>
                    <div className="text-xs" style={{ color: 'var(--recruit-muted)' }}>
                      {uploading
                        ? t('public.apply.uploading')
                        : uploadedKey
                          ? t('public.apply.uploaded')
                          : uploadError
                            ? uploadError
                            : `${(resumeFile.size / 1024 / 1024).toFixed(1)} MB`}
                    </div>
                  </div>
                  {uploadedKey && !uploading && (
                    <Check className="h-4 w-4" style={{ color: 'var(--recruit-accent)' }} />
                  )}
                  <button
                    type="button"
                    onClick={clearFile}
                    className="rounded-md p-1 hover:bg-gray-100"
                    title={t('public.apply.remove')}
                  >
                    <X className="h-4 w-4" style={{ color: 'var(--recruit-muted)' }} />
                  </button>
                </div>
              )}
              {uploadError && !resumeFile && (
                <p className="mt-1 text-xs text-red-600 font-medium">{uploadError}</p>
              )}
            </Field>
          )}

          {/* Fallback for candidates who'd rather link out. Hidden when
              a file is picked - two resume sources on one form is
              confusing. */}
          {!resumeFile && !uploadedKey && (
            <Field label={t('public.apply.linkInstead')} hint={t('public.apply.linkHint')}>
              <input
                type="url"
                placeholder="https://…"
                value={resumeUrl}
                onChange={(e) => setResumeUrl(e.target.value)}
                className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
                style={{ borderColor: 'var(--recruit-border)' }}
              />
            </Field>
          )}
          <Field label={t('public.apply.whyYou')}>
            <textarea
              rows={6}
              value={coverLetter}
              onChange={(e) => setCoverLetter(e.target.value)}
              placeholder={t('public.apply.whyYouPlaceholder')}
              className="w-full rounded-md border bg-white px-3 py-2.5 text-base leading-relaxed outline-none focus:ring-2"
              style={{ borderColor: 'var(--recruit-border)' }}
            />
          </Field>

          {/* Custom per-role questions. Rendered in admin-defined order
              after the baseline fields so the baseline never shifts. */}
          {customFields.map((f) => (
            <CustomFieldRenderer
              key={f.id}
              field={f}
              value={answers[String(f.id)] ?? (f.kind === 'multi_select' ? [] : '')}
              onChange={(v) => setAnswer(f.id, v)}
            />
          ))}

          {/* GDPR / legal consent row. The business's privacy and terms
              URLs, when set, link out in a new tab. */}
          {(customization.requireConsent || customization.privacyUrl || customization.termsUrl) && (
            <label
              className="flex items-start gap-2 rounded-md border p-3 text-sm"
              style={{ borderColor: 'var(--recruit-border)' }}
            >
              <input
                type="checkbox"
                required={customization.requireConsent}
                checked={consented}
                onChange={(e) => setConsented(e.target.checked)}
                className="mt-0.5 h-4 w-4"
              />
              <span style={{ color: 'var(--recruit-muted)' }}>
                <Trans
                  i18nKey={consentKey(customization.privacyUrl, customization.termsUrl)}
                  values={{ brand: customization.brandName }}
                  components={{
                    privacy: (
                      <a
                        href={customization.privacyUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="underline"
                        style={{ color: 'var(--recruit-accent)' }}
                      />
                    ),
                    terms: (
                      <a
                        href={customization.termsUrl}
                        target="_blank"
                        rel="noreferrer"
                        className="underline"
                        style={{ color: 'var(--recruit-accent)' }}
                      />
                    ),
                  }}
                />
              </span>
            </label>
          )}

          <button
            type="submit"
            disabled={submit.isPending || uploading}
            className="inline-flex w-full items-center justify-center gap-2 rounded-lg px-6 py-3 text-sm font-semibold text-white transition hover:opacity-90 disabled:opacity-50"
            style={{ background: 'var(--recruit-accent)' }}
          >
            {uploading
              ? t('public.apply.uploadingResume')
              : submit.isPending
                ? t('public.apply.submitting')
                : t('public.apply.submit')}
          </button>
        </form>
      </main>
    </div>
  );
};

// CustomFieldRenderer dispatches on the field's kind. The value is
// kept as a string or string[] union so the rest of the form doesn't
// need per-kind state. Unknown kinds degrade to a single-line text
// input so a mis-typed kind on the server still lets the candidate
// apply.
interface CustomFieldRendererProps {
  field: import('../../shared/lib/types').FormField;
  value: string | string[];
  onChange: (v: string | string[]) => void;
}
const CustomFieldRenderer: React.FC<CustomFieldRendererProps> = ({ field, value, onChange }) => {
  const { t } = useTranslation();
  const options = parseOptions(field.optionsJson);
  const strValue = Array.isArray(value) ? '' : value;
  const arrValue = Array.isArray(value) ? value : [];

  switch (field.kind) {
    case 'long_text':
      return (
        <Field label={field.label} required={field.required}>
          <textarea
            rows={4}
            required={field.required}
            value={strValue}
            onChange={(e) => onChange(e.target.value)}
            placeholder={field.placeholder}
            className="w-full rounded-md border bg-white px-3 py-2.5 text-base leading-relaxed outline-none focus:ring-2"
            style={{ borderColor: 'var(--recruit-border)' }}
          />
        </Field>
      );
    case 'url':
      return (
        <Field label={field.label} required={field.required}>
          <input
            type="url"
            required={field.required}
            value={strValue}
            onChange={(e) => onChange(e.target.value)}
            placeholder={field.placeholder || 'https://…'}
            className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
            style={{ borderColor: 'var(--recruit-border)' }}
          />
        </Field>
      );
    case 'single_select':
      return (
        <Field label={field.label} required={field.required}>
          <select
            required={field.required}
            value={strValue}
            onChange={(e) => onChange(e.target.value)}
            className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
            style={{ borderColor: 'var(--recruit-border)' }}
          >
            <option value="">{t('public.apply.select')}</option>
            {options.map((o) => (
              <option key={o} value={o}>
                {o}
              </option>
            ))}
          </select>
        </Field>
      );
    case 'multi_select':
      return (
        <Field label={field.label} required={field.required}>
          <div className="flex flex-wrap gap-2">
            {options.map((o) => {
              const active = arrValue.includes(o);
              return (
                <button
                  key={o}
                  type="button"
                  onClick={() =>
                    onChange(active ? arrValue.filter((x) => x !== o) : [...arrValue, o])
                  }
                  className="rounded-full border px-3 py-1.5 text-sm"
                  style={{
                    borderColor: active ? 'var(--recruit-accent)' : 'var(--recruit-border)',
                    background: active ? 'var(--recruit-accent)' : 'transparent',
                    color: active ? 'white' : 'var(--recruit-fg)',
                  }}
                >
                  {o}
                </button>
              );
            })}
          </div>
        </Field>
      );
    case 'boolean':
      return (
        <label className="flex cursor-pointer items-center gap-2 py-1">
          <input
            type="checkbox"
            checked={strValue === '1'}
            onChange={(e) => onChange(e.target.checked ? '1' : '')}
            className="h-4 w-4"
            required={field.required}
          />
          <span className="text-sm">
            {field.label}
            {field.required && <span style={{ color: 'var(--recruit-accent)' }}> *</span>}
          </span>
        </label>
      );
    case 'short_text':
    default:
      return (
        <Field label={field.label} required={field.required}>
          <input
            type="text"
            required={field.required}
            value={strValue}
            onChange={(e) => onChange(e.target.value)}
            placeholder={field.placeholder}
            className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
            style={{ borderColor: 'var(--recruit-border)' }}
          />
        </Field>
      );
  }
};

function parseOptions(json: string): string[] {
  if (!json) return [];
  try {
    const parsed = JSON.parse(json);
    return Array.isArray(parsed) ? parsed.map(String) : [];
  } catch {
    return [];
  }
}

interface FieldProps {
  label: string;
  required?: boolean;
  hint?: string;
  children: React.ReactNode;
}
const Field: React.FC<FieldProps> = ({ label, required, hint, children }) => (
  <label className="block">
    <div className="mb-1.5 flex items-baseline justify-between">
      <span className="text-sm font-medium">
        {label}
        {required && <span style={{ color: 'var(--recruit-accent)' }}> *</span>}
      </span>
      {hint && (
        <span className="text-xs" style={{ color: 'var(--recruit-muted)' }}>
          {hint}
        </span>
      )}
    </div>
    {children}
  </label>
);

export default ApplyPage;
