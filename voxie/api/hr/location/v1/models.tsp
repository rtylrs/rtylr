import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.location.v1;

/** LocationPayload is a physical work location for a tenant. */
model LocationPayload {
  /** Unique identifier for the location. */
  @encode(string)
  id: int64;

  /** Tenant owner this location belongs to. */
  @encode(string)
  owner_id: int64;

  /** Human-readable name of the location (e.g. "Downtown Office"). */
  name: string;

  /** Full street address of the location. */
  address: string;

  /** IANA timezone identifier (e.g. "America/New_York"); absent if using tenant default. */
  timezone?: string;

  /** Whether this location is currently active and assignable. */
  is_active: boolean;

  /** Unix timestamp when this location was created. */
  @encode(string)
  created_at: int64;

  /** Unix timestamp when this location was last updated. */
  @encode(string)
  updated_at: int64;

  /**
   * URL-safe slug, scoped per business. Powers public URLs like
   * portal.rtylr.com/<business-slug>/menu/<location-slug>.
   */
  slug: string;
}

/**
 * ListLocationsResponse lists all locations for a tenant.
 *
 * NOTE: this message has no corresponding request message or RPC anywhere
 * in the source proto (service.proto declares only CreateLocation and
 * UpdateLocation) - it appears to be dead/unused. Carried over unchanged
 * and flagged for manual follow-up.
 */
model ListLocationsResponse {
  /** All locations for the tenant. */
  locations: LocationPayload[];
}

model CreateLocationRequest {
  /** Human-readable name of the location. */
  @Onkt.required
  name: string;

  /** Full street address of the location. */
  address: string;

  /** IANA timezone identifier; absent to use tenant default. */
  timezone?: string;

  /** Optional URL-safe slug; auto-generated from name if absent. */
  slug?: string;
}

model CreateLocationResponse {
  /** The newly created location. */
  location: LocationPayload;
}

/** UpdateLocationBody carries updatable location fields bound from the HTTP payload. */
model UpdateLocationBody {
  /** New name; absent to keep current. */
  name?: string;

  /** New address; absent to keep current. */
  address?: string;

  /** New timezone; absent to keep current. */
  timezone?: string;

  /** New active status; absent to keep current. */
  is_active?: boolean;

  /** New slug; absent to keep current. */
  slug?: string;
}

model UpdateLocationRequest {
  /** ID of the location to update. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  @body
  body: UpdateLocationBody;
}

model UpdateLocationResponse {
  /** The updated location. */
  location: LocationPayload;
}
