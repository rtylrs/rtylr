import { describe, it, expect } from "vitest";
import { isPosEvent, KDS_EVENT_TYPES } from "./realtimeEvents";

describe("isPosEvent", () => {
  it("returns true when the event type matches", () => {
    const event = { topic: "owner:1", type: "pos:order:paid" };
    expect(isPosEvent(event, "pos:order:paid")).toBe(true);
  });

  it("returns false when the event type does not match", () => {
    const event = { topic: "owner:1", type: "pos:order:created" };
    expect(isPosEvent(event, "pos:order:paid")).toBe(false);
  });

  it("returns false for an unrelated event type", () => {
    const event = { topic: "owner:1", type: "some:other:event" };
    expect(isPosEvent(event, "pos:order:created")).toBe(false);
  });

  it("filters a mixed event list to the requested type", () => {
    const events = [
      { topic: "owner:1", type: "pos:order:created", entity_id: 1 },
      { topic: "owner:1", type: "pos:order:paid", entity_id: 2 },
    ];
    const created = events.filter((e) => isPosEvent(e, "pos:order:created"));
    expect(created).toHaveLength(1);
    expect(created[0]?.entity_id).toBe(1);
  });
});

describe("KDS_EVENT_TYPES", () => {
  it("contains pos:order:created", () => {
    expect(KDS_EVENT_TYPES.has("pos:order:created")).toBe(true);
  });

  it("contains pos:order:preparing", () => {
    expect(KDS_EVENT_TYPES.has("pos:order:preparing")).toBe(true);
  });

  it("contains pos:kds:item:updated", () => {
    expect(KDS_EVENT_TYPES.has("pos:kds:item:updated")).toBe(true);
  });

  it("contains order lifecycle events that remove tickets from the KDS", () => {
    expect(KDS_EVENT_TYPES.has("pos:order:voided")).toBe(true);
    expect(KDS_EVENT_TYPES.has("pos:order:paid")).toBe(true);
  });

  it("does not contain unrelated events", () => {
    expect(KDS_EVENT_TYPES.has("pos:table:status")).toBe(false);
    expect(KDS_EVENT_TYPES.has("pos:stock:low")).toBe(false);
  });

  it("has exactly 5 members", () => {
    expect(KDS_EVENT_TYPES.size).toBe(5);
  });
});
