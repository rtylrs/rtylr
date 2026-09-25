import { unwrap } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { rest } from '../../lib/clients';
import type { Customization, FormField, RecruitCareer, Section, Submission } from '../../lib/types';

// Admin hooks - require auth. Every mutation invalidates the matching
// list query so the HR team sees edits live without a manual refresh.

interface ListCareersResponse {
  careers: RecruitCareer[];
}

export function useAdminCareers() {
  return useQuery({
    queryKey: ['recruit', 'admin', 'careers'],
    queryFn: () => unwrap(rest.get<ListCareersResponse>('/recruit/careers')),
  });
}

interface CareerBody {
  id?: number;
  slug: string;
  title: string;
  department?: string;
  location?: string;
  employmentType?: string;
  workMode?: string;
  seniority?: string;
  details: string;
  requirements?: string;
  salaryMin?: number;
  salaryMax?: number;
  salaryCurrency?: string;
  isPublished: boolean;
}

interface CreateCareerResponse {
  career: RecruitCareer;
}

export function useCreateCareer() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CareerBody) => {
      const { id: _id, ...payload } = body;
      return unwrap(rest.post<CreateCareerResponse>('/recruit/careers', payload));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
    onError: () => toast.error(t('common.errors.createCareer')),
  });
}

export function useUpdateCareer() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CareerBody) =>
      unwrap(rest.put<{ message: string }>('/recruit/careers', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
    onError: () => toast.error(t('common.errors.updateCareer')),
  });
}

export function useDeleteCareer() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      unwrap(rest.delete<{ message: string }>('/recruit/careers', { data: { id } })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
    onError: () => toast.error(t('common.errors.deleteCareer')),
  });
}

interface GetCustomizationResponse {
  customization: Customization | null;
}

export function useAdminCustomization() {
  return useQuery({
    queryKey: ['recruit', 'admin', 'customization'],
    queryFn: () => unwrap(rest.get<GetCustomizationResponse>('/recruit/customization')),
  });
}

interface CustomizationBody {
  slug: string;
  brandName: string;
  tagline?: string;
  aboutBlurb?: string;
  logoUrl?: string;
  primaryColor?: string;
  accentColor?: string;
  bgColor?: string;
  textColor?: string;
  mutedTextColor?: string;
  fontPairing?: string;
  heroImageUrl?: string;
  faviconUrl?: string;
  template: string;
  headerHtml?: string;
  metaTitle?: string;
  metaDescription?: string;
  privacyUrl?: string;
  termsUrl?: string;
  gaMeasurementId?: string;
  metaPixelId?: string;
  requireConsent?: boolean;
  notificationEmail?: string;
  emailNotifications?: boolean;
}

export function useUpsertCustomization() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CustomizationBody) =>
      unwrap(rest.put<{ message: string }>('/recruit/customization', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'customization'] }),
    onError: () => toast.error(t('common.errors.updateCustomization')),
  });
}

interface ListSubmissionsResponse {
  submissions: Submission[];
}

export function useAdminSubmissions(status?: string, limit: number = 20, offset: number = 0) {
  return useQuery({
    queryKey: ['recruit', 'admin', 'submissions', status ?? '', limit, offset],
    queryFn: () =>
      unwrap(
        rest.get<ListSubmissionsResponse>('/recruit/submissions', {
          params: {
            ...(status ? { status } : {}),
            limit,
            offset,
          },
        })
      ),
  });
}

export function useUpdateSubmissionStatus() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { id: number; status: string }) =>
      unwrap(rest.put<{ message: string }>('/recruit/submissions/status', body)),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'submissions'], exact: false }),
    onError: () => toast.error(t('common.errors.updateSubmission')),
  });
}

// Resume links are signed on demand - we never expose the raw R2 URL
// in the list payload, so a screenshot of the admin table doesn't
// expose anything persistent. The returned URL is good for one hour.
interface ResumeDownloadResponse {
  url: string;
  expiresAt: string;
}

export async function fetchResumeURL(submissionId: number): Promise<string> {
  const res = await unwrap(
    rest.get<ResumeDownloadResponse>('/recruit/submissions/resume-url', {
      params: { submission_id: submissionId },
    })
  );
  return res.url;
}

// ── sections ─────────────────────────────────────────────────────────

interface ListSectionsResponse {
  sections: Section[];
}

export function useAdminSections() {
  return useQuery({
    queryKey: ['recruit', 'admin', 'sections'],
    queryFn: () => unwrap(rest.get<ListSectionsResponse>('/recruit/sections')),
  });
}

export interface SectionBody {
  id?: number;
  kind: string;
  title?: string;
  body?: string;
  mediaUrl?: string;
  itemsJson?: string;
  position?: number;
  enabled: boolean;
}

export function useCreateSection() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: SectionBody) => {
      const { id: _id, ...payload } = body;
      return unwrap(rest.post<{ section: Section }>('/recruit/sections', payload));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error(t('common.errors.createSection')),
  });
}

export function useUpdateSection() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: SectionBody) =>
      unwrap(rest.put<{ message: string }>('/recruit/sections', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error(t('common.errors.updateSection')),
  });
}

export function useDeleteSection() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      unwrap(rest.delete<{ message: string }>('/recruit/sections', { data: { id } })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error(t('common.errors.deleteSection')),
  });
}

export function useReorderSections() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (orderedIds: number[]) =>
      unwrap(rest.put<{ message: string }>('/recruit/sections/reorder', { orderedIds })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error(t('common.errors.reorderSections')),
  });
}

// ── form fields ─────────────────────────────────────────────────────

interface ListFormFieldsResponse {
  fields: FormField[];
}

export function useAdminFormFields(careerId: number | undefined) {
  return useQuery({
    queryKey: ['recruit', 'admin', 'form-fields', careerId],
    enabled: !!careerId,
    queryFn: () =>
      unwrap(
        rest.get<ListFormFieldsResponse>('/recruit/form-fields', {
          params: { career_id: careerId },
        })
      ),
  });
}

export interface FormFieldBody {
  id?: number;
  careerId: number;
  kind: string;
  label: string;
  placeholder?: string;
  required: boolean;
  optionsJson?: string;
  position?: number;
}

export function useCreateFormField() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: FormFieldBody) => {
      const { id: _id, ...payload } = body;
      return unwrap(rest.post<{ field: FormField }>('/recruit/form-fields', payload));
    },
    onSuccess: (_, vars) =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'form-fields', vars.careerId] }),
    onError: () => toast.error(t('common.errors.createField')),
  });
}

export function useUpdateFormField() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: FormFieldBody) =>
      unwrap(rest.put<{ message: string }>('/recruit/form-fields', body)),
    onSuccess: (_, vars) =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'form-fields', vars.careerId] }),
    onError: () => toast.error(t('common.errors.updateField')),
  });
}

export function useDeleteFormField() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: { id: number; careerId: number }) =>
      unwrap(
        rest.delete<{ message: string }>('/recruit/form-fields', { data: { id: payload.id } })
      ),
    onSuccess: (_, vars) =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'form-fields', vars.careerId] }),
    onError: () => toast.error(t('common.errors.deleteField')),
  });
}
