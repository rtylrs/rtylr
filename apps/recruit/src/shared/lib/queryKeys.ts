// Query key factories for @tanstack/react-query used in the recruit app.
// Centralizing keys here ensures cache invalidation is explicit and
// avoids typo-induced stale data bugs.

export const recruitKeys = {
  all: ['recruit'] as const,

  public: () => [...recruitKeys.all, 'public'] as const,
  publicCareers: (slug: string) => [...recruitKeys.public(), 'careers', slug] as const,
  publicCareer: (slug: string, careerSlug: string) =>
    [...recruitKeys.publicCareers(slug), careerSlug] as const,
  publicCustomization: (slug: string) => [...recruitKeys.public(), 'customization', slug] as const,

  admin: () => [...recruitKeys.all, 'admin'] as const,
  adminCareers: () => [...recruitKeys.admin(), 'careers'] as const,
  adminCareer: (id: number) => [...recruitKeys.adminCareers(), id] as const,
  adminSubmissions: (careerId?: number) =>
    [...recruitKeys.admin(), 'submissions', { careerId }] as const,
  adminCustomization: () => [...recruitKeys.admin(), 'customization'] as const,
  adminFormFields: (careerId: number) => [...recruitKeys.admin(), 'form-fields', careerId] as const,
  adminSections: () => [...recruitKeys.admin(), 'sections'] as const,
} as const;
