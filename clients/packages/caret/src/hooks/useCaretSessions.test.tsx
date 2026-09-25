import React from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useCaretSessions } from "./useCaretSessions";

const memoryClient = vi.hoisted(() => ({
  listChats: vi.fn(),
}));

vi.mock("../lib/memoryClient", () => ({
  createMemoryClient: () => memoryClient,
}));

describe("useCaretSessions", () => {
  it("keeps a failed history request distinguishable from an empty history and allows retry", async () => {
    const requestError = new Error("history service unavailable");
    memoryClient.listChats.mockRejectedValueOnce(requestError);
    memoryClient.listChats.mockResolvedValueOnce({ chats: [] });

    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const wrapper = ({
      children,
    }: React.PropsWithChildren): React.JSX.Element => (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );

    const { result } = renderHook(
      () => useCaretSessions({ userId: 7, apiBaseUrl: "http://127.0.0.1:9" }),
      { wrapper },
    );

    await waitFor(() => expect(result.current.error).toBe(requestError));
    expect(result.current.sessions).toEqual([]);
    expect(result.current.isLoading).toBe(false);

    await act(async () => {
      await result.current.refetch();
    });

    expect(memoryClient.listChats).toHaveBeenCalledTimes(2);
    await waitFor(() => expect(result.current.error).toBeNull());
    expect(result.current.sessions).toEqual([]);
  });
});
