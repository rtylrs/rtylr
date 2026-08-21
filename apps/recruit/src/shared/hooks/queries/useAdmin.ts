import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { recruitBody, recruitCompatibilityClient, recruitResult } from '../../lib/clients';
import type { Customization, FormField, RecruitCareer, Section, Submission } from '../../lib/types';

// Admin hooks - require auth. Every mutation invalidates the matching
// list query so the HR team sees edits live without a manual refresh.

interface ListCareersResponse {
  careers: RecruitCareer[];
}

export function useAdminCareers() {
  return useQuery({
    queryKey: ['recruit', 'admin', 'careers'],
    queryFn: () =>
      unwrap(recruitResult<ListCareersResponse>(() => recruitCompatibilityClient.listCareers({}))),
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
    mutationFn: (body: CareerBody) => {
      const { id: _id, ...payload } = body;
      return unwrap(
        recruitResult<CreateCareerResponse>(() =>
          recruitCompatibilityClient.createCareer(recruitBody(payload))
        )
      );
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
    onError: () => toast.error('Failed to create career'),
  });
}

export function useUpdateCareer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CareerBody) =>
      unwrap(
        recruitResult<{ message: string }>(() =>
          recruitCompatibilityClient.updateCareer(recruitBody(body))
        )
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
    onError: () => toast.error('Failed to update career'),
  });
}

export function useDeleteCareer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      unwrap(
        recruitResult<{ message: string }>(() =>
          recruitCompatibilityClient.deleteCareer({ id: String(id) })
        )
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
    onError: () => toast.error('Failed to delete career'),
  });
}

interface GetCustomizationResponse {
  customization: Customization | null;
}

export function useAdminCustomization() {
  return useQuery({
    queryKey: ['recruit', 'admin', 'customization'],
    queryFn: () =>
      unwrap(
        recruitResult<GetCustomizationResponse>(() =>
          recruitCompatibilityClient.getCustomization({})
        )
      ),
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
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CustomizationBody) =>
      unwrap(
        recruitResult<{ message: string }>(() =>
          recruitCompatibilityClient.updateCustomization(recruitBody(body))
        )
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'customization'] }),
    onError: () => toast.error('Failed to update customization'),
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
        recruitResult<ListSubmissionsResponse>(() =>
          recruitCompatibilityClient.listSubmissions({
            ...(status ? { status } : {}),
            limit,
            offset,
          })
        )
      ),
  });
}

export function useUpdateSubmissionStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { id: number; status: string }) =>
      unwrap(
        recruitResult<{ message: string }>(() =>
          recruitCompatibilityClient.updateSubmissionStatus(recruitBody(body))
        )
      ),
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'submissions'], exact: false }),
    onError: () => toast.error('Failed to update submission status'),
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
    recruitResult<ResumeDownloadResponse>(() =>
      recruitCompatibilityClient.getResumeURL({ submissionId })
    )
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
    queryFn: () =>
      unwrap(
        recruitResult<ListSectionsResponse>(() => recruitCompatibilityClient.listSections({}))
      ),
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
    mutationFn: (body: SectionBody) => {
      const { id: _id, ...payload } = body;
      return unwrap(
        recruitResult<{ section: Section }>(() =>
          recruitCompatibilityClient.createSection(recruitBody(payload))
        )
      );
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error('Failed to create section'),
  });
}

export function useUpdateSection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: SectionBody) =>
      unwrap(
        recruitResult<{ message: string }>(() =>
          recruitCompatibilityClient.updateSection(recruitBody(body))
        )
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error('Failed to update section'),
  });
}

export function useDeleteSection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      unwrap(
        recruitResult<{ message: string }>(() =>
          recruitCompatibilityClient.deleteSection({ id: String(id) })
        )
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error('Failed to delete section'),
  });
}

export function useReorderSections() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (orderedIds: number[]) =>
      unwrap(
        recruitResult<{ message: string }>(() =>
          recruitCompatibilityClient.reorderSections(recruitBody({ orderedIds }))
        )
      ),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error('Failed to reorder sections'),
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
        recruitResult<ListFormFieldsResponse>(() =>
          recruitCompatibilityClient.listFormFields({ careerId })
        )
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
    mutationFn: (body: FormFieldBody) => {
      const { id: _id, ...payload } = body;
      return unwrap(
        recruitResult<{ field: FormField }>(() =>
          recruitCompatibilityClient.createFormField(recruitBody(payload))
        )
      );
    },
    onSuccess: (_, vars) =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'form-fields', vars.careerId] }),
    onError: () => toast.error('Failed to create form field'),
  });
}

export function useUpdateFormField() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: FormFieldBody) =>
      unwrap(
        recruitResult<{ message: string }>(() =>
          recruitCompatibilityClient.updateFormField(recruitBody(body))
        )
      ),
    onSuccess: (_, vars) =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'form-fields', vars.careerId] }),
    onError: () => toast.error('Failed to update form field'),
  });
}

export function useDeleteFormField() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (payload: { id: number; careerId: number }) =>
      unwrap(
        recruitResult<{ message: string }>(() =>
          recruitCompatibilityClient.deleteFormField({ id: String(payload.id) })
        )
      ),
    onSuccess: (_, vars) =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'form-fields', vars.careerId] }),
    onError: () => toast.error('Failed to delete form field'),
  });
}
