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

export interface Customization {
  id: number;
  ownerId: number;
  slug: string;
  brandName: string;
  tagline: string;
  aboutBlurb: string;
  logoUrl: string;
  primaryColor: string;
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
