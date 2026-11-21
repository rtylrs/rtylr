import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace pos.audit.v1;

/** AuditService defines POS audit log endpoints. */
@route("/pos/audit/v1")
interface AuditService {
  /** Returns audit log entries filtered by action, entity, or date range. */
  @get
  @route("/audit")
  listAudit(...ListAuditRequest): ListAuditResponse;

  /** Records a new audit log entry. */
  @post
  @route("/audit")
  createAudit(...CreateAuditRequest): CreateAuditResponse;
}
