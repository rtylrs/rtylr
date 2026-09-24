import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const searchCustomers = vi.hoisted(() => vi.fn());

vi.mock("@voxire/api", () => ({
  createOneKitClient: () => ({ searchCustomers }),
}));

import { useCustomerLookup } from "./useCustomerLookup";

describe("useCustomerLookup", () => {
  beforeEach(() => {
    searchCustomers.mockReset();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("clears a previous failure as soon as the query becomes too short", async () => {
    searchCustomers.mockRejectedValue(new Error("Search unavailable"));
    const { result } = renderHook(() => useCustomerLookup());

    act(() => result.current.setQuery("ab"));
    await waitFor(
      () => expect(result.current.error).toBe("Search unavailable"),
      { timeout: 1000 },
    );

    act(() => result.current.setQuery("a"));

    await waitFor(() => expect(result.current.error).toBeNull());
    expect(result.current.results).toEqual([]);
    expect(result.current.loading).toBe(false);
    expect(searchCustomers).toHaveBeenCalledTimes(1);
  });

  it("does not apply a late response from a superseded query", async () => {
    let resolveFirst: ((value: { customers: unknown[] }) => void) | undefined;
    searchCustomers.mockImplementation((request: { query: string }) => {
      if (request.query === "ab") {
        return new Promise((resolve) => {
          resolveFirst = resolve;
        });
      }
      return Promise.resolve({
        customers: [{ id: "2", firstName: "Current", lastName: "Result" }],
      });
    });

    const { result } = renderHook(() => useCustomerLookup());
    act(() => result.current.setQuery("ab"));
    await new Promise((resolve) => setTimeout(resolve, 350));
    act(() => result.current.setQuery("ac"));
    await waitFor(
      () => expect(result.current.results[0]?.name).toBe("Current Result"),
      { timeout: 1000 },
    );

    await act(async () => {
      resolveFirst?.({
        customers: [{ id: "1", firstName: "Stale", lastName: "Result" }],
      });
      await Promise.resolve();
    });
    expect(result.current.results[0]?.name).toBe("Current Result");
  });
});
