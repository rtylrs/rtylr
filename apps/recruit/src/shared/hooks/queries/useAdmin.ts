import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import type { Customization, FormField, RecruitCareer, Section, Submission } from '../../lib/types';

// Admin hooks — require auth. Every mutation invalidates the matching
// list query so the HR team sees edits live without a manual refresh.

interface ListCareersResponse {
  careers: RecruitCareer[];
}

export function useAdminCareers() {
  return useQuery({
    queryKey: ['recruit', 'admin', 'careers'],
    queryFn: () => unwrap(safe.get<ListCareersResponse>('/recruit/careers')),
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
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CareerBody) =>
      unwrap(safe.post<CreateCareerResponse>('/recruit/careers', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
  });
}

export function useUpdateCareer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CareerBody) =>
      unwrap(safe.put<{ message: string }>('/recruit/careers', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
  });
}

export function useDeleteCareer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      unwrap(safe.delete<{ message: string }>('/recruit/careers', { data: { id } })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
  });
}

interface GetCustomizationResponse {
  customization: Customization | null;
}

export function useAdminCustomization() {
  return useQuery({
    queryKey: ['recruit', 'admin', 'customization'],
    queryFn: () => unwrap(safe.get<GetCustomizationResponse>('/recruit/customization')),
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
}

export function useUpsertCustomization() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CustomizationBody) =>
      unwrap(safe.put<{ message: string }>('/recruit/customization', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'customization'] }),
  });
}

interface ListSubmissionsResponse {
  submissions: Submission[];
}

export function useAdminSubmissions(status?: string) {
  return useQuery({
    queryKey: ['recruit', 'admin', 'submissions', status ?? ''],
    queryFn: () =>
      unwrap(
        safe.get<ListSubmissionsResponse>('/recruit/submissions', {
          params: status ? { status } : undefined,
        })
      ),
  });
}

export function useUpdateSubmissionStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { id: number; status: string }) =>
      unwrap(safe.put<{ message: string }>('/recruit/submissions/status', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'submissions'] }),
  });
}

// Resume links are signed on demand — we never expose the raw R2 URL
// in the list payload, so a screenshot of the admin table doesn't
// expose anything persistent. The returned URL is good for one hour.
interface ResumeDownloadResponse {
  url: string;
  expiresAt: string;
}

export async function fetchResumeURL(submissionId: number): Promise<string> {
  const res = await unwrap(
    safe.get<ResumeDownloadResponse>('/recruit/submissions/resume-url', {
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
    queryFn: () => unwrap(safe.get<ListSectionsResponse>('/recruit/sections')),
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
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: SectionBody) =>
      unwrap(safe.post<{ section: Section }>('/recruit/sections', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
  });
}

export function useUpdateSection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: SectionBody) =>
      unwrap(safe.put<{ message: string }>('/recruit/sections', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
  });
}

export function useDeleteSection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      unwrap(safe.delete<{ message: string }>('/recruit/sections', { data: { id } })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
  });
}

export function useReorderSections() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (orderedIds: number[]) =>
      unwrap(safe.put<{ message: string }>('/recruit/sections/reorder', { orderedIds })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
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
        safe.get<ListFormFieldsResponse>('/recruit/form-fields', {
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
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: FormFieldBody) =>
      unwrap(safe.post<{ field: FormField }>('/recruit/form-fields', body)),
    onSuccess: (_, vars) =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'form-fields', vars.careerId] }),
  });
}

export function useUpdateFormField() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: FormFieldBody) =>
      unwrap(safe.put<{ message: string }>('/recruit/form-fields', body)),
    onSuccess: (_, vars) =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'form-fields', vars.careerId] }),
  });
}

export function useDeleteFormField(careerId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      unwrap(safe.delete<{ message: string }>('/recruit/form-fields', { data: { id } })),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'form-fields', careerId] }),
  });
}
