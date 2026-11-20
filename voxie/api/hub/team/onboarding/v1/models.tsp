import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hub.team.onboarding.v1;

model MarkOnboardedRequest {}

model MarkOnboardedResponse {
  message: string;
}
