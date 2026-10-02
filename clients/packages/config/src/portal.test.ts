import { describe, expect, it } from "vitest";
import { PORTAL_URL } from "./apps";
import { resolvePortalReturn } from "./portal";

describe("Portal sign-in return destination", () => {
  it("preserves an encoded page, repeated query parameters and fragment after sign-in", () => {
    const path = "/team/15?tab=access&filter=a%26b&filter=c#licenses";
    expect(resolvePortalReturn(path)).toBe(PORTAL_URL + path);
  });
  it.each([
    "https://untrusted.example",
    "//untrusted.example",
    "/\\untrusted.example",
    "/\nuntrusted.example",
    "",
    null,
    undefined,
  ])("rejects untrusted return destination %j", (path) => {
    expect(resolvePortalReturn(path)).toBe(PORTAL_URL);
  });
});
