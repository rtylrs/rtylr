import { unwrap } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery } from '@tanstack/react-query';
import { toast } from 'sonner';
import { rest } from '../../lib/clients';
import type { Customization, FormField, RecruitCareer, Section } from '../../lib/types';

// Public (anonymous) hooks - used by the candidate-facing pages.
// Every request is scoped by the URL slug so we don't need auth.

interface ListPublicCareersResponse {
  careers: RecruitCareer[];
  customization: Customization;
  sections: Section[];
}

export function useBusinessCareers(slug: string | undefined) {
  return useQuery({
    queryKey: ['recruit', 'public', 'careers', slug],
    enabled: !!slug,
    queryFn: () =>
      unwrap(
        rest.get<ListPublicCareersResponse>('/recruit/public/careers', {
          params: { slug },
        })
      ),
  });
}

interface GetPublicCareerResponse {
  career: RecruitCareer;
  customization: Customization;
  fields: FormField[];
}

export function usePublicCareer(slug: string | undefined, careerSlug: string | undefined) {
  return useQuery({
    queryKey: ['recruit', 'public', 'career', slug, careerSlug],
    enabled: !!slug && !!careerSlug,
    queryFn: () =>
      unwrap(
        rest.get<GetPublicCareerResponse>('/recruit/public/career', {
          params: { slug, career_slug: careerSlug },
        })
      ),
  });
}

interface SubmitApplicationBody {
  slug: string;
  careerSlug: string;
  candidateName: string;
  email: string;
  phone?: string;
  resumeUrl?: string;
  coverLetter?: string;
  // Stringified JSON object keyed by form_field_id → answer (string
  // or string[] for multi_select / boolean). The server stores this
  // opaquely and the admin decodes it for display.
  answersJson?: string;
  consented?: boolean;
}

interface SubmitApplicationResponse {
  id: number;
  message: string;
}

export function useSubmitApplication() {
  const { t } = useTranslation('recruitCommon');
  return useMutation({
    mutationFn: (body: SubmitApplicationBody) =>
      unwrap(rest.post<SubmitApplicationResponse>('/recruit/public/apply', body)),
    onError: () => toast.error(t('recruitCommon.errors.apply')),
  });
}

// Upload flow: ask the backend for a presigned R2 URL, PUT the file
// directly to R2, then submit the application with the returned key.
// We intentionally expose this as a plain helper (not a useMutation)
// because the PUT step talks to Cloudflare, not our backend - the
// axios client's base URL and auth-cookie logic shouldn't apply.

interface ResumeUploadURLBody {
  slug: string;
  careerSlug: string;
  fileName: string;
  contentType: string;
}

export interface ResumeUploadURLResponse {
  uploadUrl: string;
  key: string;
  expiresAt: string;
  maxBytes: number;
}

async function requestUploadURL(body: ResumeUploadURLBody): Promise<ResumeUploadURLResponse> {
  return unwrap(rest.post<ResumeUploadURLResponse>('/recruit/public/resume-upload-url', body));
}

/**
 * uploadResume asks the backend for a presigned URL, uploads the file
 * to R2, and returns the R2 object key on success. Any step that
 * errors bubbles up with a readable message so the form can render it
 * inline. We deliberately don't do chunked/multipart uploads - the
 * hard ceiling is 10 MB and a single PUT is plenty for resumes.
 */
export async function uploadResume(args: {
  file: File;
  slug: string;
  careerSlug: string;
}): Promise<string> {
  const presigned = await requestUploadURL({
    slug: args.slug,
    careerSlug: args.careerSlug,
    fileName: args.file.name,
    contentType: args.file.type || 'application/pdf',
  });

  if (args.file.size > presigned.maxBytes) {
    throw new Error(
      `File is ${(args.file.size / 1024 / 1024).toFixed(1)} MB - max is ${(
        presigned.maxBytes /
        1024 /
        1024
      ).toFixed(0)} MB`
    );
  }

  // Direct PUT to R2. Axios would add auth headers we don't want here,
  // so we drop to fetch. The Content-Type header must match what we
  // sent to the presigner - R2 signs that value in.
  const putRes = await fetch(presigned.uploadUrl, {
    method: 'PUT',
    body: args.file,
    headers: { 'Content-Type': args.file.type || 'application/pdf' },
  });
  if (!putRes.ok) {
    throw new Error(`Upload failed (${putRes.status})`);
  }

  return presigned.key;
}
