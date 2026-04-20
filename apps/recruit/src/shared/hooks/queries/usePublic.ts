import { unwrap } from '@voxire/api';
import { useMutation, useQuery } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import type { Customization, RecruitCareer } from '../../lib/types';

// Public (anonymous) hooks — used by the candidate-facing pages.
// Every request is scoped by the URL slug so we don't need auth.

interface ListPublicCareersResponse {
  careers: RecruitCareer[];
  customization: Customization;
}

export function useBusinessCareers(slug: string | undefined) {
  return useQuery({
    queryKey: ['recruit', 'public', 'careers', slug],
    enabled: !!slug,
    queryFn: () =>
      unwrap(
        safe.get<ListPublicCareersResponse>('/recruit/public/careers', {
          params: { slug },
        })
      ),
  });
}

interface GetPublicCareerResponse {
  career: RecruitCareer;
  customization: Customization;
}

export function usePublicCareer(slug: string | undefined, careerSlug: string | undefined) {
  return useQuery({
    queryKey: ['recruit', 'public', 'career', slug, careerSlug],
    enabled: !!slug && !!careerSlug,
    queryFn: () =>
      unwrap(
        safe.get<GetPublicCareerResponse>('/recruit/public/career', {
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
}

interface SubmitApplicationResponse {
  id: number;
  message: string;
}

export function useSubmitApplication() {
  return useMutation({
    mutationFn: (body: SubmitApplicationBody) =>
      unwrap(safe.post<SubmitApplicationResponse>('/recruit/public/apply', body)),
  });
}
