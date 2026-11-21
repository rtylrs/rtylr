import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.onboarding.v1;

/** OnboardingService defines HR onboarding checklist endpoints. */
@route("/hr/onboarding/v1")
interface OnboardingService {
  /** ListChecklistTemplates returns checklist templates with optional filters. */
  @get
  @route("/templates")
  listChecklistTemplates(
    ...ListChecklistTemplatesRequest,
  ): ListChecklistTemplatesResponse;

  /** CreateChecklistTemplate creates a checklist template. */
  @post
  @route("/templates")
  createChecklistTemplate(
    ...CreateChecklistTemplateRequest,
  ): CreateChecklistTemplateResponse;

  /** UpdateChecklistTemplate updates a checklist template. */
  @put
  @route("/templates/{role_id}")
  updateChecklistTemplate(
    ...UpdateChecklistTemplateRequest,
  ): UpdateChecklistTemplateResponse;

  /** AddItemToChecklistTemplate adds an item to a checklist template. */
  @post
  @route("/templates/{role_id}/items")
  addItemToChecklistTemplate(
    ...AddItemToChecklistTemplateRequest,
  ): AddItemToChecklistTemplateResponse;

  /** ListEmployeeChecklists returns employee checklists with optional filters. */
  @get
  @route("/checklists")
  listEmployeeChecklists(
    ...ListEmployeeChecklistsRequest,
  ): ListEmployeeChecklistsResponse;

  /** GetEmployeeChecklist returns a single employee checklist. */
  @get
  @route("/employees/{employee_id}")
  getEmployeeChecklist(
    ...GetEmployeeChecklistRequest,
  ): GetEmployeeChecklistResponse;

  /** TriggerChecklist triggers onboarding for an employee. */
  @post
  @route("/checklists/trigger")
  triggerChecklist(...TriggerChecklistRequest): TriggerChecklistResponse;

  /** CompleteChecklistItem marks a checklist item complete. */
  @post
  @route("/checklists/items/complete")
  completeChecklistItem(
    ...CompleteChecklistItemRequest,
  ): CompleteChecklistItemResponse;
}
