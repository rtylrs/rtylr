import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace support.inbox.v1;

@route("/support/inbox/v1")
interface InboxService {
  @get
  @route("/inbox")
  listInbox(...ListInboxRequest): ListInboxResponse;

  @get
  @route("/inbox/unread-count")
  getUnreadCount(...GetUnreadCountRequest): GetUnreadCountResponse;
}
