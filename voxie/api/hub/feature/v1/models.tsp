import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.feature.v1;

model FeatureResolution {
  key: string;
  enabled: boolean;
  variant: string;
}

model FeatureFlag {
  key: string;
  description: string;
  enabled: boolean;
  rollout_percentage: int32;
  variant_on: string;
  variant_off: string;
}

model FeatureTarget {
  feature_key: string;

  @encode(string)
  owner_id: int64;

  enabled: boolean;
  variant: string;
}

model FeatureCondition {
  attribute: string;
  operator: string;
  value: string;
}

model FeatureRule {
  @encode(string)
  id: int64;

  feature_key: string;
  priority: int32;
  description: string;
  variant: string;
  rollout_percentage: int32;
  enabled: boolean;
  conditions: FeatureCondition[];
}

model ListFlagsRequest {}

model ListFlagsResponse {
  flags: FeatureFlag[];
}

model UpsertFlagRequest {
  @Onkt.required
  key: string;

  description: string;
  enabled: boolean;
  rollout_percentage: int32;
  variant_on: string;
  variant_off: string;
}

model UpsertFlagResponse {
  message: string;
}

model DeleteFlagRequest {
  @Onkt.required
  key: string;
}

model DeleteFlagResponse {
  message: string;
}

model ResolveFeaturesRequest {
  subject: string;
  attributes: Record<string>;
}

model ResolveFeaturesResponse {
  features: Record<FeatureResolution>;
  variant: string;
}

model ResolveMyFeaturesRequest {
  attributes: Record<string>;
}

model ResolveMyFeaturesResponse {
  features: Record<FeatureResolution>;
  variant: string;
}

model ListRulesRequest {
  @Onkt.required
  @query("feature_key")
  feature_key: string;
}

model ListRulesResponse {
  rules: FeatureRule[];
}

model CreateRuleRequest {
  @Onkt.required
  feature_key: string;

  priority: int32;
  description: string;
  variant: string;
  rollout_percentage: int32;
  enabled: boolean;
  conditions: FeatureCondition[];
}

model CreateRuleResponse {
  @encode(string)
  rule_id: int64;

  message: string;
}

model UpdateRuleRequest {
  @encode(string)
  id: int64;

  priority: int32;
  description: string;
  variant: string;
  rollout_percentage: int32;
  enabled: boolean;
  conditions: FeatureCondition[];
}

model UpdateRuleResponse {
  message: string;
}

model DeleteRuleRequest {
  @encode(string)
  id: int64;
}

model DeleteRuleResponse {
  message: string;
}

model ListTargetsRequest {
  @Onkt.required
  @query("feature_key")
  feature_key: string;
}

model ListTargetsResponse {
  targets: FeatureTarget[];
}

model UpsertTargetRequest {
  @Onkt.required
  feature_key: string;

  @encode(string)
  owner_id: int64;

  enabled: boolean;
  variant: string;
}

model UpsertTargetResponse {
  message: string;
}

model DeleteTargetRequest {
  @Onkt.required
  feature_key: string;

  @encode(string)
  owner_id: int64;
}

model DeleteTargetResponse {
  message: string;
}
