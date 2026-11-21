import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace hr.one_on_one.v1;

@route("/hr/one-on-one/v1")
interface OneOnOneService {
  @get
  @route("/one-on-ones")
  listOneOnOnes(...ListOneOnOnesRequest): ListOneOnOnesResponse;

  @post
  @route("/one-on-ones")
  createOneOnOne(...CreateOneOnOneRequest): CreateOneOnOneResponse;
}
