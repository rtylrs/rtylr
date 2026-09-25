import { describe, expect, it } from "vitest";
import { buildCaretRouteContext } from "./routeContext";

describe("buildCaretRouteContext", () => {
  it("does not label an unmatched route with the root route", () => {
    const context = buildCaretRouteContext({
      appId: "dash",
      pathname: "/audit",
      routeLabels: {
        "/": "Dashboard home",
        "/customers": "Customers",
      },
      routeModules: {
        "/": "dashboard",
        "/customers": "customers",
      },
    });

    expect(context.page).toBe("Audit");
    expect(context.module).toBe("audit");
  });
});
