import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { ToolCallBlock } from "../../types/message";
import ToolCallCard from "./ToolCallCard";

describe("ToolCallCard", () => {
  it("keeps technical tool metadata out of the user-facing details", () => {
    const call: ToolCallBlock = {
      id: "tool-1",
      type: "tool_call",
      name: "insights.get_product_kpis",
      input: { business_id: 42 },
      status: "success",
      output: { product_kpis: { total_skus: 10 } },
      summary: "Checked product data.",
    };

    render(<ToolCallCard call={call} />);
    fireEvent.click(
      screen.getByRole("button", { name: /Completed: check business data/i }),
    );

    expect(screen.queryByText("technical name")).toBeNull();
    expect(screen.queryByText("request details")).toBeNull();
    expect(screen.queryByText("insights.get_product_kpis")).toBeNull();
    expect(screen.queryByText(/business_id/)).toBeNull();
    expect(screen.getByText("Checked product data.")).not.toBeNull();
  });

  it("animates state-appropriate orbs for pending and running tools", () => {
    const pending: ToolCallBlock = {
      id: "tool-pending",
      type: "tool_call",
      name: "erp.create_product",
      input: {},
      status: "pending",
      requiresApproval: true,
    };
    const { unmount } = render(<ToolCallCard call={pending} />);
    expect(screen.getByRole("img", { name: "Shaping…" })).not.toBeNull();
    unmount();

    const running: ToolCallBlock = {
      id: "tool-running",
      type: "tool_call",
      name: "crm.list_customers",
      input: {},
      status: "running",
    };
    render(<ToolCallCard call={running} />);
    expect(screen.getByRole("img", { name: "Searching…" })).not.toBeNull();
  });

  it("shows the exact grounded web answer, safe sources, and associated Google suggestions", () => {
    const call: ToolCallBlock = {
      id: "web-search-1",
      type: "tool_call",
      name: "web.search",
      input: { query: "public question" },
      status: "success",
      output: {
        answer: "A sourced answer from Google.",
        sources: [
          {
            title: "Google AI docs",
            url: "https://ai.google.dev/gemini-api/docs/google-search",
          },
          { title: "Unsafe source", url: "javascript:alert(1)" },
        ],
        search_suggestions_html:
          '<div><a href="https://www.google.com/search?q=gemini">Gemini</a><a href="javascript:alert(1)">Unsafe</a><a href="https://user@google.com/search?q=bad">Credential URL</a><a href="https://evil.google.com/search?q=bad">Lookalike host</a></div>',
      },
    };

    render(<ToolCallCard call={call} />);

    expect(screen.getByText("A sourced answer from Google.")).not.toBeNull();
    expect(
      screen.getByRole("link", { name: "Google AI docs" }).getAttribute("href"),
    ).toBe("https://ai.google.dev/gemini-api/docs/google-search");
    expect(
      screen.getByRole("link", { name: "Gemini" }).getAttribute("href"),
    ).toBe("https://www.google.com/search?q=gemini");
    expect(screen.queryByRole("link", { name: "Unsafe source" })).toBeNull();
    expect(screen.queryByRole("link", { name: "Unsafe" })).toBeNull();
    expect(screen.queryByRole("link", { name: "Credential URL" })).toBeNull();
    expect(screen.queryByRole("link", { name: "Lookalike host" })).toBeNull();
    expect(screen.queryByTitle("Google Search suggestions")).toBeNull();
  });
});
