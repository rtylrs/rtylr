import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;
using hub.business.v1;

namespace hub.config.v1;

model ConfigPayload {
  /**
   * Identifies the business that owns this config. Serialized as a string
   * in the legacy payload.
   */
  owner_id: string;

  site_name: string;
  logo_url: string;
  favicon_url: string;
  meta_tags: string[];
  contact_email: string;
  contact_number: string;
  sender_email: string;
  email_provider: string;
  sender_name: string;
  reply_to_email: string;
  email_notifications_enabled: boolean;
  smtp_host: string;
  smtp_port: int32;
  smtp_username: string;
  smtp_password_configured: boolean;
  website: string;
  maps_url: string;
  address: string;
  city: string;
  zipcode: string;
  country: string;
  instagram: string;
  maintenance_mode: boolean;
  maintenance_notice: string;
  addons_enabled: boolean;
  enable_reviews: boolean;
  enable_wishlist: boolean;
  enable_newsletter_signup: boolean;
  about_content: string;
  footer_sub_text: string;
  slug: string;
}

model CreateConfigRequest {
  /**
   * Accepted for legacy compatibility; the server derives the owning
   * business from the authenticated user.
   */
  owner: string;

  @Onkt.required
  key: string;

  @Onkt.required
  value: string;

  description: string;
}

model CreateConfigResponse {
  /** Identifies the created config. */
  @encode(string)
  config_id: int64;

  message: string;
}

model GetConfigRequest {}

model GetConfigResponse {
  /** business is declared in hub.business.v1 - a cross-package reference. */
  business: Business;

  maintenance: boolean;
}

model IndustryType {
  slug: string;
  label: string;
  icon: string;
}

model IndustryGroup {
  category: string;
  types: IndustryType[];
}

model GetIndustriesRequest {}

model GetIndustriesResponse {
  groups: IndustryGroup[];
}

model UpdateConfigRequest {
  /** Absent means unchanged. */
  site_name?: string;

  logo_url?: string;
  favicon_url?: string;
  contact_email?: string;
  contact_number?: string;
  sender_email?: string;
  email_provider?: string;
  sender_name?: string;
  reply_to_email?: string;
  email_notifications_enabled?: boolean;
  smtp_host?: string;
  smtp_port?: int32;
  smtp_username?: string;
  smtp_password?: string;
  website?: string;
  maps_url?: string;
  address?: string;
  city?: string;
  zipcode?: string;
  country?: string;
  instagram?: string;
  maintenance_mode?: boolean;
  maintenance_notice?: string;
  addons_enabled?: boolean;
  enable_reviews?: boolean;
  enable_wishlist?: boolean;
  enable_newsletter_signup?: boolean;
  about_content?: string;
  footer_sub_text?: string;
}

model UpdateConfigResponse {
  message: string;
  config: ConfigPayload;
}

model SendTestEmailRequest {
  @Onkt.required
  recipient: string;
}

model SendTestEmailResponse {
  message: string;
  queued: boolean;
}

model CustomDomain {
  @encode(string)
  id: int64;

  @encode(string)
  business_id: int64;

  domain: string;
  status: string;
  verification_token: string;
  ssl_status: string;
  verified_at: string;
  created_at: string;
  updated_at: string;
}

model GetCustomDomainRequest {}

model GetCustomDomainResponse {
  domain: CustomDomain;
}

model RegisterCustomDomainRequest {
  @Onkt.required
  domain: string;
}

model RegisterCustomDomainResponse {
  domain: CustomDomain;
}

model VerifyCustomDomainRequest {
  @Onkt.required
  verification_token: string;
}

model VerifyCustomDomainResponse {
  domain: CustomDomain;
}

model RemoveCustomDomainRequest {}

model RemoveCustomDomainResponse {
  message: string;
}
