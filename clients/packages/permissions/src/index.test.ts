import { describe, expect, it } from "vitest";
import type { EffectiveAccess } from "@voxire/types";
import {
  canAllEffectiveAccess,
  canAnyEffectiveAccess,
  canEffectiveAccess,
  canEffectiveAccessInScope,
  hasEffectiveAppAccess,
} from "./index";

function access(permissions: string[], permissionScopes: Record<string, { scopeType: "flow_project"; scopeId: string }[]> = {}): EffectiveAccess {
  return {
    businessId: "12",
    dashPermissions: [],
    subscribedApps: ["flow"],
    licenses: [],
    apps: [{
      appId: "flow",
      licensed: true,
      hasAccess: true,
      permissions,
      permissionScopes: Object.fromEntries(permissions.map((permission) => [permission, permissionScopes[permission] ?? []])),
      sources: ["explicit"],
    }],
    isOwner: false,
    isAdmin: false,
    authorizationVersion: 1,
  };
}

describe("effective permission checks", () => {
  it("keeps read and write grants independent", () => {
    const readOnly = access(["flow:tasks:read"]);
    const writeOnly = access(["flow:tasks:write"]);

    expect(canEffectiveAccess(readOnly, "flow:tasks:read")).toBe(true);
    expect(canEffectiveAccess(readOnly, "flow:tasks:write")).toBe(false);
    expect(canEffectiveAccess(writeOnly, "flow:tasks:write")).toBe(true);
    expect(canEffectiveAccess(writeOnly, "flow:tasks:read")).toBe(false);
  });

  it("requires both an active app grant and a license", () => {
    const unlicensed = access(["flow:tasks:read"]);
    unlicensed.apps[0]!.licensed = false;
    const unassigned = access(["flow:tasks:read"]);
    unassigned.apps[0]!.hasAccess = false;

    expect(canEffectiveAccess(unlicensed, "flow:tasks:read")).toBe(false);
    expect(hasEffectiveAppAccess(unlicensed, "flow")).toBe(false);
    expect(canEffectiveAccess(unassigned, "flow:tasks:read")).toBe(false);
  });

  it("checks the scopes attached to the requested permission", () => {
    const scoped = access(["flow:tasks:read", "flow:tasks:write"], {
      "flow:tasks:read": [{ scopeType: "flow_project", scopeId: "41" }],
      "flow:tasks:write": [{ scopeType: "flow_project", scopeId: "82" }],
    });

    expect(canEffectiveAccessInScope(scoped, "flow:tasks:read", { scopeType: "flow_project", scopeId: "41" })).toBe(true);
    expect(canEffectiveAccessInScope(scoped, "flow:tasks:read", { scopeType: "flow_project", scopeId: "82" })).toBe(false);
    expect(canEffectiveAccessInScope(scoped, "flow:tasks:write", { scopeType: "flow_project", scopeId: "82" })).toBe(true);
  });

  it("treats an empty scope list as business-wide and missing access as deny", () => {
    const broad = access(["flow:tasks:read"]);
    expect(canEffectiveAccessInScope(broad, "flow:tasks:read", { scopeType: "flow_project", scopeId: "41" })).toBe(true);
    expect(canEffectiveAccess(null, "flow:tasks:read")).toBe(false);
    expect(canAllEffectiveAccess(broad, ["flow:tasks:read", "flow:tasks:write"])).toBe(false);
    expect(canAnyEffectiveAccess(broad, ["flow:tasks:read", "flow:tasks:write"])).toBe(true);
  });
});
