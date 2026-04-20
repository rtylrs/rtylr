// Wire types for the recruit API. These mirror the Go payload structs
// in onion's careers module; keep them in sync until we promote them
// to @voxire/contracts protos.

export interface RecruitCareer {
  id: number;
  ownerId: number;
  slug: string;
  title: string;
  department: string;
  location: string;
  employmentType: string;
  workMode: string;
  seniority: string;
  details: string;
  requirements: string;
  salaryMin: number;
  salaryMax: number;
  salaryCurrency: string;
  isPublished: boolean;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export type TemplateId = 'minimal' | 'editorial' | 'bold';

// FontPairingId matches the keys in FONT_PAIRINGS (templates/fonts.ts).
// Empty string means "use the template default"; anything else must be
// a known pairing or the frontend ignores it.
export type FontPairingId =
  | ''
  | 'jakarta-fraunces'
  | 'space-grotesk-inter'
  | 'playfair-inter'
  | 'dmserif-dmsans'
  | 'default';

export interface Customization {
  id: number;
  ownerId: number;
  slug: string;
  brandName: string;
  tagline: string;
  aboutBlurb: string;
  logoUrl: string;
  primaryColor: string;
  accentColor: string;
  bgColor: string;
  textColor: string;
  mutedTextColor: string;
  fontPairing: FontPairingId | string;
  heroImageUrl: string;
  faviconUrl: string;
  template: TemplateId | string;
  headerHtml: string;
  createdAt: string;
  updatedAt: string;
}

export interface Submission {
  id: number;
  careerId: number;
  ownerId: number;
  candidateName: string;
  email: string;
  phone: string;
  resumeUrl: string;
  coverLetter: string;
  status: string;
  careerTitle: string;
  createdAt: string;
  updatedAt: string;
}
