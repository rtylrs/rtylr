import React, { useEffect, useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';
import { ArrowLeft, Check } from 'lucide-react';

import Header from '../../templates/shared/Header';
import { usePublicCareer, useSubmitApplication } from '../../shared/hooks/queries/usePublic';
import { applyBrand } from '../../templates/resolver';
import NotFound from './NotFound';

/**
 * ApplyPage — form at recruit.rtylr.com/:slug/:careerSlug/apply.
 *
 * MVP scope: name, email, phone, resume URL (linked from wherever the
 * candidate already hosts it — Drive, Dropbox, personal site), cover
 * letter. Native multipart resume upload is a follow-up; storing a URL
 * keeps the schema stable across that change.
 *
 * The form renders the shared Header so branding carries across the
 * apply flow. It intentionally doesn't route through the template
 * resolver — templates diverge on marketing surfaces, but the apply
 * form is a utility and benefits from a single consistent shape.
 */
const ApplyPage: React.FC = (): React.JSX.Element => {
  const { slug = '', careerSlug = '' } = useParams<{ slug: string; careerSlug: string }>();
  const navigate = useNavigate();

  const { data, isLoading, error } = usePublicCareer(slug, careerSlug);
  const submit = useSubmitApplication();

  useEffect(() => {
    if (data?.customization) applyBrand(data.customization.primaryColor);
  }, [data?.customization]);

  const [candidateName, setCandidateName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [resumeUrl, setResumeUrl] = useState('');
  const [coverLetter, setCoverLetter] = useState('');
  const [done, setDone] = useState(false);

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
  if (error || !data) return <NotFound />;

  const { career, customization } = data;

  const onSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    submit.mutate(
      {
        slug,
        careerSlug,
        candidateName,
        email,
        phone: phone || undefined,
        resumeUrl: resumeUrl || undefined,
        coverLetter: coverLetter || undefined,
      },
      {
        onSuccess: () => {
          setDone(true);
          toast.success('Application received');
        },
        onError: () => toast.error("Couldn't submit — try again"),
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
          <h1 className="mt-6 text-3xl font-semibold tracking-tight">You're in.</h1>
          <p className="mt-4 text-base" style={{ color: 'var(--recruit-muted)' }}>
            Thanks for applying to <span className="font-medium">{career.title}</span>. The{' '}
            {customization.brandName} team will review and reach out if there's a fit.
          </p>
          <button
            onClick={() => navigate(`/${slug}`)}
            className="mt-8 inline-flex items-center gap-2 rounded-lg border px-5 py-2.5 text-sm font-medium"
            style={{ borderColor: 'var(--recruit-border)' }}
          >
            <ArrowLeft className="h-4 w-4" />
            Back to open roles
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
          to={`/${slug}/${careerSlug}`}
          className="inline-flex items-center gap-2 text-sm font-medium hover:underline"
          style={{ color: 'var(--recruit-muted)' }}
        >
          <ArrowLeft className="h-4 w-4" />
          Back to role
        </Link>

        <h1 className="mt-6 text-3xl font-semibold tracking-tight">Apply for {career.title}</h1>
        <p className="mt-2 text-sm" style={{ color: 'var(--recruit-muted)' }}>
          {customization.brandName} — {career.location || 'Remote'}
        </p>

        <form onSubmit={onSubmit} className="mt-8 space-y-5">
          <Field label="Full name" required>
            <input
              required
              value={candidateName}
              onChange={(e) => setCandidateName(e.target.value)}
              className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
              style={{ borderColor: 'var(--recruit-border)' }}
            />
          </Field>
          <Field label="Email" required>
            <input
              required
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
              style={{ borderColor: 'var(--recruit-border)' }}
            />
          </Field>
          <Field label="Phone">
            <input
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
              style={{ borderColor: 'var(--recruit-border)' }}
            />
          </Field>
          <Field
            label="Resume link"
            hint="Paste a link to your resume (Google Drive, Dropbox, personal site, etc.)"
          >
            <input
              type="url"
              placeholder="https://…"
              value={resumeUrl}
              onChange={(e) => setResumeUrl(e.target.value)}
              className="w-full rounded-md border bg-white px-3 py-2.5 text-base outline-none focus:ring-2"
              style={{ borderColor: 'var(--recruit-border)' }}
            />
          </Field>
          <Field label="Why you">
            <textarea
              rows={6}
              value={coverLetter}
              onChange={(e) => setCoverLetter(e.target.value)}
              placeholder="A short note on why you'd be a great fit."
              className="w-full rounded-md border bg-white px-3 py-2.5 text-base leading-relaxed outline-none focus:ring-2"
              style={{ borderColor: 'var(--recruit-border)' }}
            />
          </Field>

          <button
            type="submit"
            disabled={submit.isPending}
            className="inline-flex w-full items-center justify-center gap-2 rounded-lg px-6 py-3 text-sm font-semibold text-white transition hover:opacity-90 disabled:opacity-50"
            style={{ background: 'var(--recruit-accent)' }}
          >
            {submit.isPending ? 'Submitting…' : 'Submit application'}
          </button>
        </form>
      </main>
    </div>
  );
};

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
