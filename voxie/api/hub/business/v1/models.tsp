import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.business.v1;

/** Business is a merchant's top-level business entity. */
model Business {
  @encode(string)
  id: int64;

  name: string;

  /** restaurant, retail, services, etc. */
  type: string;

  logo_url: string;
  address: string;
  phone: string;

  /** Email address used for lookup or contact. */
  email: string;

  website: string;
  timezone: string;
  currency: string;
  country: string;
  plan: string;

  /** When the record was created. */
  created_at: string;

  /** URL-safe slug. Powers public surfaces like portal.rtylr.com/<business-slug>/menu. Reserved words are rejected. */
  slug: string;
}

/** BusinessEntry is a compact business row for ownership listings. */
model BusinessEntry {
  @encode(string)
  id: int64;

  name: string;
  logo_url: string;

  /** Identifies the parent business, zero for top-level businesses. */
  @encode(string)
  parent_id: int64;

  is_active: boolean;

  /** True when the current user owns the business; false for team access. */
  is_owned: boolean;
}

/** PublicLocation is a location entry for public menu surfaces. */
model PublicLocation {
  @encode(string)
  id: int64;

  name: string;
  slug: string;
  address: string;
  timezone: string;
  is_active: boolean;
}

/** PublicBusinessSummary resolves a public slug into a business summary. */
model PublicBusinessSummary {
  @encode(string)
  id: int64;

  name: string;
  slug: string;
  logo_url: string;
  currency: string;
  type: string;
  timezone: string;
}

/** PublicBusinessProfile is public-safe business branding for anonymous menu surfaces. */
model PublicBusinessProfile {
  @encode(string)
  id: int64;

  name: string;
  logo_url: string;
  currency: string;
  timezone: string;
}

model GetBusinessRequest {
  @minValueExclusive(0)
  @encode(string)
  business_id: int64;
}

model GetBusinessResponse {
  business: Business;
}

model CreateSubBusinessRequest {
  @Onkt.required
  name: string;

  @Onkt.required
  @minValueExclusive(0)
  @encode(string)
  parent_business_id?: int64;
}

model CreateSubBusinessResponse {
  @encode(string)
  id: int64;
}

model OnboardBusinessRequest {
  /** The first business name created during owner onboarding. */
  @Onkt.required
  name: string;
}

model OnboardBusinessResponse {
  /** A fresh JWT carrying the new business_id. */
  token: string;
}

model CheckBusinessNameExistsRequest {
  @Onkt.required
  @query("name")
  name: string;
}

model CheckBusinessNameExistsResponse {
  exists: boolean;
}

model GetPublicBusinessRequest {
  @minValueExclusive(0)
  @query("business_id")
  @encode(string)
  business_id: int64;
}

model GetPublicBusinessResponse {
  business: PublicBusinessProfile;
}

model ListMyBusinessesRequest {}

model ListMyBusinessesResponse {
  businesses: BusinessEntry[];
}

model LookupBusinessBySlugRequest {
  @Onkt.required
  @pattern("^[a-z0-9]$|^[a-z0-9][a-z0-9-]{0,78}[a-z0-9]$")
  @query("slug")
  slug: string;
}

model LookupBusinessBySlugResponse {
  business: PublicBusinessSummary;
  locations: PublicLocation[];
}

model SetSlugRequest {
  @Onkt.required
  @pattern("^[a-z0-9]$|^[a-z0-9][a-z0-9-]{0,78}[a-z0-9]$")
  slug: string;
}

model SetSlugResponse {
  slug: string;
}

model UpdateBusinessRequest {
  @minValueExclusive(0)
  @encode(string)
  business_id: int64;

  @Onkt.required
  name: string;

  @Onkt.required
  type: string;

  address: string;
  phone: string;

  @Onkt.required
  @format("email")
  email: string;

  website: string;

  @Onkt.required
  timezone: string;

  @Onkt.required
  currency: string;

  slug?: string;
}

model UpdateBusinessResponse {
  business: Business;
}

model RenameBusinessRequest {
  @minValueExclusive(0)
  @encode(string)
  business_id: int64;

  @Onkt.required
  name: string;
}

model RenameBusinessResponse {
  business: Business;
}

model SetLogoRequest {
  @minValueExclusive(0)
  @encode(string)
  business_id: int64;

  /** Public URL of an already-uploaded asset (see \@voxire/api uploadAsset). */
  @Onkt.required
  logo_url: string;
}

model SetLogoResponse {
  business: Business;
}

model DeleteBusinessRequest {
  @minValueExclusive(0)
  @encode(string)
  business_id: int64;
}

model DeleteBusinessResponse {
  message: string;
}
