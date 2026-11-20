import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.setup.v1;

/** SetupService defines HR setup wizard endpoints. */
@route("/hr/setup/v1")
interface SetupService {
  /** Marks a setup step complete. */
  @post
  @route("/steps")
  completeSetupStep(...CompleteSetupStepRequest): CompleteSetupStepResponse;
}
