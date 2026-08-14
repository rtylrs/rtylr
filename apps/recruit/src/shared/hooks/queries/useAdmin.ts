import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { safe } from '../../lib/api';
import type {
  Customization,
  FormField,
  RecruitCareer,
  Section,
  Submission,
  SubmissionInterview,
  SubmissionNote,
  SubmissionOffer,
  SubmissionTag,
  SubmissionEvent,
  TalentPool,
  TalentPoolMember,
} from '../../lib/types';

// Admin hooks - require auth. Every mutation invalidates the matching
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
    mutationFn: (body: CareerBody) => {
      const { id: _id, ...payload } = body;
      return unwrap(safe.post<CreateCareerResponse>('/recruit/careers', payload));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
    onError: () => toast.error('Failed to create career'),
  });
}

export function useUpdateCareer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CareerBody) =>
      unwrap(safe.put<{ message: string }>('/recruit/careers', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'careers'] }),
    onError: () => toast.error('Failed to update career'),
  });
}

export function useDeleteCareer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      unwrap(safe.delete<{ message: string }>('/recruit/careers', { data: { id } })),
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
      unwrap(safe.put<{ message: string }>('/recruit/customization', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'customization'] }),
    onError: () => toast.error('Failed to update customization'),
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
    onSuccess: () =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'submissions'], exact: false }),
    onError: () => toast.error('Failed to update submission status'),
  });
}

interface ListSubmissionEventsResponse {
  events: SubmissionEvent[];
}

export function useSubmissionEvents(submissionId: number | undefined) {
  return useQuery({
    queryKey: ['recruit', 'admin', 'submission', submissionId, 'events'],
    enabled: !!submissionId,
    queryFn: () =>
      unwrap(safe.get<ListSubmissionEventsResponse>(`/recruit/submissions/${submissionId}/events`)),
  });
}

interface ListSubmissionTagsResponse {
  tags: SubmissionTag[];
}

export function useSubmissionTags(submissionId: number | undefined) {
  return useQuery({
    queryKey: ['recruit', 'admin', 'submission', submissionId, 'tags'],
    enabled: !!submissionId,
    queryFn: () =>
      unwrap(safe.get<ListSubmissionTagsResponse>(`/recruit/submissions/${submissionId}/tags`)),
  });
}

export function useCreateSubmissionTag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ submissionId, name }: { submissionId: number; name: string }) =>
      unwrap(
        safe.post<{ tagId: number }>(`/recruit/submissions/${submissionId}/tags`, { name })
      ),
    onSuccess: (_data, variables) =>
      qc.invalidateQueries({
        queryKey: ['recruit', 'admin', 'submission', variables.submissionId, 'tags'],
      }),
    onError: () => toast.error('Failed to add candidate tag'),
  });
}

export function useDeleteSubmissionTag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ submissionId, tagId }: { submissionId: number; tagId: number }) =>
      unwrap(
        safe.delete<{ success: boolean }>(
          `/recruit/submissions/${submissionId}/tags/${tagId}`
        )
      ),
    onSuccess: (_data, variables) =>
      qc.invalidateQueries({
        queryKey: ['recruit', 'admin', 'submission', variables.submissionId, 'tags'],
      }),
    onError: () => toast.error('Failed to remove candidate tag'),
  });
}

interface ListTalentPoolsResponse {
  pools: TalentPool[];
}

interface ListTalentPoolMembersResponse {
  members: TalentPoolMember[];
}

export function useTalentPools() {
  return useQuery({
    queryKey: ['recruit', 'admin', 'talent-pools'],
    queryFn: () => unwrap(safe.get<ListTalentPoolsResponse>('/recruit/talent-pools')),
  });
}

export function useCreateTalentPool() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: { name: string; description?: string }) =>
      unwrap(safe.post<{ poolId: number }>('/recruit/talent-pools', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'talent-pools'] }),
    onError: () => toast.error('Failed to create talent pool'),
  });
}

export function useUpdateTalentPool() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ poolId, name, description }: { poolId: number; name: string; description?: string }) =>
      unwrap(safe.put<{ message: string }>(`/recruit/talent-pools/${poolId}`, { name, description })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'talent-pools'] }),
    onError: () => toast.error('Failed to update talent pool'),
  });
}

export function useDeleteTalentPool() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (poolId: number) =>
      unwrap(safe.delete<{ success: boolean }>(`/recruit/talent-pools/${poolId}`)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'talent-pools'] }),
    onError: () => toast.error('Failed to delete talent pool'),
  });
}

export function useTalentPoolMembers(poolId: number | undefined) {
  return useQuery({
    queryKey: ['recruit', 'admin', 'talent-pool', poolId, 'members'],
    enabled: !!poolId,
    queryFn: () =>
      unwrap(safe.get<ListTalentPoolMembersResponse>(`/recruit/talent-pools/${poolId}/members`)),
  });
}

export function useAddTalentPoolMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ poolId, submissionId }: { poolId: number; submissionId: number }) =>
      unwrap(
        safe.post<{ memberId: number }>(`/recruit/talent-pools/${poolId}/members`, {
          submissionId,
        })
      ),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'talent-pools'] });
      qc.invalidateQueries({
        queryKey: ['recruit', 'admin', 'talent-pool', variables.poolId, 'members'],
      });
    },
    onError: () => toast.error('Failed to add candidate to talent pool'),
  });
}

export function useRemoveTalentPoolMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ poolId, submissionId }: { poolId: number; submissionId: number }) =>
      unwrap(
        safe.delete<{ success: boolean }>(
          `/recruit/talent-pools/${poolId}/members/${submissionId}`
        )
      ),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'talent-pools'] });
      qc.invalidateQueries({
        queryKey: ['recruit', 'admin', 'talent-pool', variables.poolId, 'members'],
      });
    },
    onError: () => toast.error('Failed to remove candidate from talent pool'),
  });
}

interface ListInterviewsResponse {
  interviews: SubmissionInterview[];
}

export function useSubmissionInterviews(submissionId: number | undefined) {
  return useQuery({
    queryKey: ['recruit', 'admin', 'submission', submissionId, 'interviews'],
    enabled: !!submissionId,
    queryFn: () =>
      unwrap(
        safe.get<ListInterviewsResponse>(`/recruit/submissions/${submissionId}/interviews`)
      ),
  });
}

export interface SubmissionInterviewBody {
  interviewType: string;
  title: string;
  scheduledAt: string;
  durationMinutes: number;
  timezone: string;
  meetingUrl?: string;
  interviewerName?: string;
  status?: string;
  feedback?: string;
}

export function useCreateSubmissionInterview() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ submissionId, ...body }: SubmissionInterviewBody & { submissionId: number }) =>
      unwrap(
        safe.post<{ interviewId: number }>(
          `/recruit/submissions/${submissionId}/interviews`,
          body
        )
      ),
    onSuccess: (_data, variables) =>
      qc.invalidateQueries({
        queryKey: ['recruit', 'admin', 'submission', variables.submissionId, 'interviews'],
      }),
    onError: () => toast.error('Failed to schedule interview'),
  });
}

interface ListNotesResponse {
  notes: SubmissionNote[];
}

export function useSubmissionNotes(submissionId: number | undefined) {
  return useQuery({
    queryKey: ['recruit', 'admin', 'submission', submissionId, 'notes'],
    enabled: !!submissionId,
    queryFn: () =>
      unwrap(safe.get<ListNotesResponse>(`/recruit/submissions/${submissionId}/notes`)),
  });
}

export function useCreateSubmissionNote() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ submissionId, body }: { submissionId: number; body: string }) =>
      unwrap(
        safe.post<{ noteId: number }>(`/recruit/submissions/${submissionId}/notes`, { body })
      ),
    onSuccess: (_data, variables) =>
      qc.invalidateQueries({
        queryKey: ['recruit', 'admin', 'submission', variables.submissionId, 'notes'],
      }),
    onError: () => toast.error('Failed to save note'),
  });
}

interface ListOffersResponse {
  offers: SubmissionOffer[];
}

export function useSubmissionOffers(submissionId: number | undefined) {
  return useQuery({
    queryKey: ['recruit', 'admin', 'submission', submissionId, 'offers'],
    enabled: !!submissionId,
    queryFn: () =>
      unwrap(safe.get<ListOffersResponse>(`/recruit/submissions/${submissionId}/offers`)),
  });
}

export interface SubmissionOfferBody {
  title: string;
  salary: number;
  salaryCurrency: string;
  startDate?: string;
  expiresAt?: string;
  status?: string;
  terms?: string;
}

export function useCreateSubmissionOffer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ submissionId, ...body }: SubmissionOfferBody & { submissionId: number }) =>
      unwrap(
        safe.post<{ offerId: number }>(`/recruit/submissions/${submissionId}/offers`, body)
      ),
    onSuccess: (_data, variables) =>
      qc.invalidateQueries({
        queryKey: ['recruit', 'admin', 'submission', variables.submissionId, 'offers'],
      }),
    onError: () => toast.error('Failed to create offer'),
  });
}

export function useUpdateSubmissionOfferStatus() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ submissionId, offerId, status }: { submissionId: number; offerId: number; status: string }) =>
      unwrap(
        safe.put<{ message: string }>(
          `/recruit/submissions/${submissionId}/offers/${offerId}/status`,
          { status }
        )
      ),
    onSuccess: (_data, variables) =>
      qc.invalidateQueries({
        queryKey: ['recruit', 'admin', 'submission', variables.submissionId, 'offers'],
      }),
    onError: () => toast.error('Failed to update offer'),
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
    mutationFn: (body: SectionBody) => {
      const { id: _id, ...payload } = body;
      return unwrap(safe.post<{ section: Section }>('/recruit/sections', payload));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error('Failed to create section'),
  });
}

export function useUpdateSection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: SectionBody) =>
      unwrap(safe.put<{ message: string }>('/recruit/sections', body)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error('Failed to update section'),
  });
}

export function useDeleteSection() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      unwrap(safe.delete<{ message: string }>('/recruit/sections', { data: { id } })),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'sections'] }),
    onError: () => toast.error('Failed to delete section'),
  });
}

export function useReorderSections() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (orderedIds: number[]) =>
      unwrap(safe.put<{ message: string }>('/recruit/sections/reorder', { orderedIds })),
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
    mutationFn: (body: FormFieldBody) => {
      const { id: _id, ...payload } = body;
      return unwrap(safe.post<{ field: FormField }>('/recruit/form-fields', payload));
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
      unwrap(safe.put<{ message: string }>('/recruit/form-fields', body)),
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
        safe.delete<{ message: string }>('/recruit/form-fields', { data: { id: payload.id } })
      ),
    onSuccess: (_, vars) =>
      qc.invalidateQueries({ queryKey: ['recruit', 'admin', 'form-fields', vars.careerId] }),
    onError: () => toast.error('Failed to delete form field'),
  });
}
