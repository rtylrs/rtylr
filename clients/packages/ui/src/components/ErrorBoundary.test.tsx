import React, { lazy, Suspense } from "react";
import { render, screen, act, cleanup } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  ErrorBoundary,
  isDynamicImportError,
  recoverFromDynamicImportError,
} from "./ErrorBoundary";

describe("isDynamicImportError", () => {
  it("detects Vite dynamic import fetch failures", () => {
    expect(
      isDynamicImportError(
        new Error(
          "Failed to fetch dynamically imported module: https://dash.rtylr.com/assets/Account-Bu2-9eGC.js",
        ),
      ),
    ).toBe(true);
  });

  it("detects common chunk load failures", () => {
    expect(isDynamicImportError(new Error("Loading chunk 42 failed."))).toBe(
      true,
    );
    expect(
      isDynamicImportError(
        new Error("ChunkLoadError: Loading chunk settings failed."),
      ),
    ).toBe(true);
  });

  it("ignores ordinary runtime errors", () => {
    expect(
      isDynamicImportError(new Error("Cannot read properties of undefined")),
    ).toBe(false);
  });
});

afterEach(() => {
  cleanup();
  window.sessionStorage.clear();
  vi.restoreAllMocks();
});

describe("dynamic import recovery", () => {
  it("keeps the reload guard while a lazy page is pending and after it fails", async () => {
    const key = `${window.location.origin}${window.location.pathname}${window.location.search}`;
    window.sessionStorage.setItem("rtylr:dynamic-import-reload", key);
    let rejectPage: (error: Error) => void = () => {
      throw new Error("Page was not initialized");
    };
    const page = new Promise<{ default: React.ComponentType }>((_, reject) => {
      rejectPage = reject;
    });
    const Page = lazy(() => page);
    const consoleError = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});
    const onError = vi.fn();
    render(
      <ErrorBoundary onError={onError}>
        <Suspense fallback={<p>Loading page</p>}>
          <Page />
        </Suspense>
      </ErrorBoundary>,
    );
    expect(screen.getByText("Loading page")).toBeVisible();
    expect(window.sessionStorage.getItem("rtylr:dynamic-import-reload")).toBe(
      key,
    );
    const error = new Error(
      "Failed to fetch dynamically imported module: /assets/page.js",
    );
    await act(async () => {
      rejectPage(error);
    });
    expect(onError).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("alert")).toBeVisible();
    expect(recoverFromDynamicImportError(error)).toBe(false);
    expect(window.sessionStorage.getItem("rtylr:dynamic-import-reload")).toBe(
      key,
    );
    consoleError.mockRestore();
  });
});
