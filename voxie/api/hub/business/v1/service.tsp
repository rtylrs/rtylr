import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.business.v1;

/** BusinessService defines hub business endpoints. */
@route("/hub/business/v1")
interface BusinessService {
  /** Returns the businesses owned by the current user. */
  @get
  @route("/businesses/mine")
  listMyBusinesses(...ListMyBusinessesRequest): ListMyBusinessesResponse;

  /** Creates a sub-business under a parent business. */
  @post
  @route("/businesses")
  createSubBusiness(...CreateSubBusinessRequest): CreateSubBusinessResponse;

  /** Creates the owner's first business during onboarding. */
  @post
  @route("/businesses/onboard")
  onboardBusiness(...OnboardBusinessRequest): OnboardBusinessResponse;

  /** Probes whether a business name is already taken. */
  @get
  @route("/businesses/name-exists")
  checkBusinessNameExists(
    ...CheckBusinessNameExistsRequest,
  ): CheckBusinessNameExistsResponse;

  /** Returns a business by ID. */
  @get
  @route("/businesses/{business_id}")
  getBusiness(...GetBusinessRequest): GetBusinessResponse;

  /** Resolves a business by public slug. */
  @get
  @route("/businesses/by-slug")
  lookupBusinessBySlug(
    ...LookupBusinessBySlugRequest,
  ): LookupBusinessBySlugResponse;

  /** Returns public-safe business branding for anonymous menu surfaces. */
  @get
  @route("/businesses/public")
  getPublicBusiness(...GetPublicBusinessRequest): GetPublicBusinessResponse;

  /** Updates only the business slug. */
  @patch
  @route("/businesses/slug")
  setSlug(...SetSlugRequest): SetSlugResponse;

  /** Updates an existing business. */
  @put
  @route("/businesses/{business_id}")
  updateBusiness(...UpdateBusinessRequest): UpdateBusinessResponse;

  /** Renames an active business owned by the current user. */
  @patch
  @route("/businesses/{business_id}/name")
  renameBusiness(...RenameBusinessRequest): RenameBusinessResponse;

  /** Sets the logo for an active business owned by the current user. */
  @patch
  @route("/businesses/{business_id}/logo")
  setLogo(...SetLogoRequest): SetLogoResponse;

  /** Removes an active business owned by the current user. */
  @delete
  @route("/businesses/{business_id}")
  deleteBusiness(...DeleteBusinessRequest): DeleteBusinessResponse;
}
