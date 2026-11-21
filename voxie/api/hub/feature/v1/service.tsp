import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.feature.v1;

@route("/hub/feature/v1")
interface FeatureService {
  @post
  @route("/features")
  resolveFeatures(...ResolveFeaturesRequest): ResolveFeaturesResponse;

  @post
  @route("/features/me")
  resolveMyFeatures(...ResolveMyFeaturesRequest): ResolveMyFeaturesResponse;

  @get
  @route("/features/admin/flags")
  listFlags(...ListFlagsRequest): ListFlagsResponse;

  @put
  @route("/features/admin/flags")
  upsertFlag(...UpsertFlagRequest): UpsertFlagResponse;

  @delete
  @route("/features/admin/flags/{key}")
  deleteFlag(...DeleteFlagRequest): DeleteFlagResponse;

  @get
  @route("/features/admin/targets")
  listTargets(...ListTargetsRequest): ListTargetsResponse;

  @put
  @route("/features/admin/targets")
  upsertTarget(...UpsertTargetRequest): UpsertTargetResponse;

  @delete
  @route("/features/admin/targets/{feature_key}/{owner_id}")
  deleteTarget(...DeleteTargetRequest): DeleteTargetResponse;

  @get
  @route("/features/admin/rules")
  listRules(...ListRulesRequest): ListRulesResponse;

  @post
  @route("/features/admin/rules")
  createRule(...CreateRuleRequest): CreateRuleResponse;

  @put
  @route("/features/admin/rules")
  updateRule(...UpdateRuleRequest): UpdateRuleResponse;

  @delete
  @route("/features/admin/rules/{id}")
  deleteRule(...DeleteRuleRequest): DeleteRuleResponse;
}
