import { describe, expect, it } from "vitest";
import { hasEffectiveAppAccess, parseEffectiveAccess } from "./effectiveAccess";

const validAccess = {
  businessId: "business-1",
  dashRole: "member",
  dashRoleId: null,
  dashPermissions: [],
  subscribedApps: ["flow"],
  licenses: [],
  apps: [
    {
      appId: "flow",
      licensed: true,
      hasAccess: true,
      roleIds: ["flow-contributor"],
      primaryRoleId: "flow-contributor",
      displayRole: "Contributor",
      permissions: ["flow:tasks:view"],
      scopes: [{ scopeType: "business", scopeId: "*" }],
      sources: ["explicit"],
      permissionSources: { "flow:tasks:view": ["flow-contributor"] },
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

  it("fails closed when a required field is malformed", () => {
    const malformed = {
      ...validAccess,
      apps: [{ ...validAccess.apps[0], hasAccess: "true" }],
    };
    expect(parseEffectiveAccess(malformed)).toBeNull();
  });

  it("treats omitted per-app collections as empty without rejecting other apps", () => {
    const access = parseEffectiveAccess({
      ...validAccess,
      apps: [
        validAccess.apps[0],
        {
          appId: "hr",
          licensed: false,
          hasAccess: false,
          roleIds: [],
          primaryRoleId: null,
          displayRole: null,
          scopes: [],
          permissionSources: {},
        },
      ],
    });

    expect(access).not.toBeNull();
    expect(hasEffectiveAppAccess(access, "flow")).toBe(true);
    expect(access?.apps[1]?.permissions).toEqual([]);
    expect(access?.apps[1]?.sources).toEqual([]);
  });

  it("does not accept unknown applications", () => {
    const malformed = {
      ...validAccess,
      subscribedApps: ["unknown"],
    };
    expect(parseEffectiveAccess(malformed)).toBeNull();
  });
});
