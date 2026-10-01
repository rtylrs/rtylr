import { describe, expect, it } from "vitest";
import { hasEffectiveAppAccess, parseEffectiveAccess } from "./effectiveAccess";

const validAccess = {
  businessId: "business-1",
  dashPermissions: [],
  subscribedApps: ["flow"],
  licenses: [],
  apps: [
    {
      appId: "flow",
      licensed: true,
      hasAccess: true,
      permissions: ["flow:tasks:view"],
      permissionScopes: [{ permissionKey: "flow:tasks:view", scopes: [] }],
      sources: ["explicit"],
    },
  ],
  isOwner: false,
  isAdmin: false,
  authorizationVersion: 3,
};

describe("parseEffectiveAccess", () => {
  it("parses a complete server access document", () => {
    const access = parseEffectiveAccess(validAccess);
    expect(access?.businessId).toBe("business-1");
    expect(access?.apps[0]?.appId).toBe("flow");
    expect(hasEffectiveAppAccess(access, "flow")).toBe(true);
    expect(hasEffectiveAppAccess(access, "finance")).toBe(false);
  });

  it("recognizes the baseline Dash launcher permission for an active member", () => {
    const access = parseEffectiveAccess({
      ...validAccess,
      dashPermissions: ["hub:dashboard:view"],
    });
    expect(hasEffectiveAppAccess(access, "dash")).toBe(true);
  });

  it("parses the resource scopes emitted by normalized authorization", () => {
    const access = parseEffectiveAccess({
      ...validAccess,
      apps: [
        {
          ...validAccess.apps[0],
          permissionScopes: [
            { permissionKey: "flow:tasks:view", scopes: [{ scopeType: "flow_project", scopeId: "41" }] },
          ],
        },
      ],
    });

    expect(access?.apps[0]?.permissionScopes["flow:tasks:view"]).toEqual([
      { scopeType: "flow_project", scopeId: "41" },
    ]);
  });

  it("fails closed when a required field is malformed", () => {
    const malformed = {
      ...validAccess,
      apps: [{ ...validAccess.apps[0], hasAccess: "true" }],
    };
    expect(parseEffectiveAccess(malformed)).toBeNull();
  });

  it("accepts omitted zero-value fields emitted by the server without granting access", () => {
    const access = parseEffectiveAccess({
      business_id: "business-1",
      authorization_version: "3",
      apps: [{ app_id: "hr" }],
    });

    expect(access).not.toBeNull();
    expect(access?.isOwner).toBe(false);
    expect(access?.isAdmin).toBe(false);
    expect(access?.dashPermissions).toEqual([]);
    expect(access?.subscribedApps).toEqual([]);
    expect(access?.apps[0]?.licensed).toBe(false);
    expect(access?.apps[0]?.hasAccess).toBe(false);
    expect(access?.apps[0]?.permissions).toEqual([]);
    expect(access?.apps[0]?.permissionScopes).toEqual({});
    expect(access?.apps[0]?.sources).toEqual([]);
    expect(hasEffectiveAppAccess(access, "hr")).toBe(false);
  });

  it("does not accept unknown applications", () => {
    const malformed = {
      ...validAccess,
      subscribedApps: ["unknown"],
    };
    expect(parseEffectiveAccess(malformed)).toBeNull();
  });
});
