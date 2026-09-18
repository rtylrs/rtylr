// Wire types for the recruit API. These mirror the Go payload structs
// in onion's careers module; keep them in sync until we promote them
// to @voxire/voxie protos.

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
  ownerId?: number;
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
  metaTitle: string;
  metaDescription: string;
  privacyUrl: string;
  termsUrl: string;
  gaMeasurementId: string;
  metaPixelId: string;
  requireConsent: boolean;
  notificationEmail?: string;
  emailNotifications?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

// SectionKind is the discriminator for content sections. Templates
// render a given section by switching on this value; unknown kinds
// fall through to a plain-text renderer so older clients still show
// something reasonable if a new kind is shipped before they update.
export type SectionKind =
  | 'perks'
  | 'culture'
  | 'testimonial'
  | 'video'
  | 'faq'
  | 'values'
  | 'cta'
  | string;

export interface Section {
  id: number;
  ownerId: number;
  kind: SectionKind;
  title: string;
  body: string;
  mediaUrl: string;
  // itemsJson is a pre-serialised JSON string. Templates parse it per
  // kind; the wire type stays opaque so adding new structured shapes
  // doesn't require a schema migration on the frontend.
  itemsJson: string;
  position: number;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

// Parsed shapes of items_json per kind - used by both the admin editor
// and the public template renderers.
export interface PerkItem {
  icon?: string; // lucide icon name
  label: string;
  blurb?: string;
}
export interface FaqItem {
  question: string;
  answer: string;
}
export interface ValueItem {
  label: string;
  blurb?: string;
}

// FormFieldKind is the discriminator for custom question types. Kept
// narrow - the apply form renderer and the builder both switch on
// this, and allowing arbitrary strings on the client would turn into
// silent "nothing renders" bugs.
export type FormFieldKind =
  | 'short_text'
  | 'long_text'
  | 'single_select'
  | 'multi_select'
  | 'url'
  | 'boolean';

export interface FormField {
  id: number;
  careerId: number;
  ownerId: number;
  kind: FormFieldKind | string;
  label: string;
  placeholder: string;
  required: boolean;
  optionsJson: string;
  position: number;
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
  answersJson: string;
  careerTitle: string;
  createdAt: string;
  updatedAt: string;
}
