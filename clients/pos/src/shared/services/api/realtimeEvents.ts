/**
 * Typed definitions for the realtime events the POS backend actually emits
 * over the WebSocket bus (onion `wsbus.Event`). The wire format is a lightweight
 * invalidation hint - `{ topic, type, entity_id, actor_id }` - with NO payload:
 * consumers refetch canonical data via REST when they see a relevant `type`.
 */

import type { RealtimeEvent } from "@voxire/api";

// The event types currently published by onion (grep `Type:.*"pos:` there):
//   pos:order:created      owner:{userId}
//   pos:order:voided       owner:{userId}
//   pos:order:paid         owner:{userId}
//   pos:order:preparing    owner:{userId}
//   pos:kds:item:updated   business:{businessId} (KDS module) and owner topic (orders module)
//   pos:table:status       business:{businessId}
//
// NOTE: `pos:stock:low` / `pos:stock:out` are NOT emitted yet. useStockAlerts
// subscribes defensively so alerts start working the moment the backend adds
// them, but today no such events arrive.

export type PosEventType =
  | "pos:order:created"
  | "pos:order:updated"
  | "pos:order:paid"
  | "pos:order:voided"
  | "pos:order:preparing"
  | "pos:kds:item:updated"
  | "pos:table:status"
  // Not emitted by the backend yet; kept so clients are already wired.
  | "pos:stock:low"
  | "pos:stock:out";

/** Narrow a generic realtime event to a known POS event type. */
export function isPosEvent(event: RealtimeEvent, type: PosEventType): boolean {
  return event.type === type;
}

/** All KDS-relevant event types - used by useKdsStream to filter the bus. */
export const KDS_EVENT_TYPES = new Set<PosEventType>([
  "pos:order:created",
  "pos:order:voided",
  "pos:order:paid",
  "pos:order:preparing",
  "pos:kds:item:updated",
]);
