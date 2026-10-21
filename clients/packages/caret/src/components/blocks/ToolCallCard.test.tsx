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

  it("shows safe web search results with source links and snippets", () => {
    const call: ToolCallBlock = {
      id: "web-search-1",
      type: "tool_call",
      name: "web.search",
      input: { query: "public question" },
      status: "success",
      output: {
        sources: [
          {
            title: "Search result",
            url: "https://example.com/source",
            snippet: "A short public snippet.",
          },
          { title: "Unsafe source", url: "javascript:alert(1)", snippet: "" },
        ],
      },
    };

    render(<ToolCallCard call={call} />);
    const toggle = screen.getByRole("button", {
      name: /Completed: search the web/i,
    });
    expect(toggle.getAttribute("aria-expanded")).toBe("false");
    expect(screen.queryByRole("link", { name: "Search result" })).toBeNull();
    fireEvent.click(toggle);

    expect(
      screen.getByRole("link", { name: "Search result" }).getAttribute("href"),
    ).toBe("https://example.com/source");
    expect(screen.getByText("A short public snippet.")).not.toBeNull();
    expect(screen.queryByRole("link", { name: "Unsafe source" })).toBeNull();
  });

  it("names the search query and lists each sub-agent's report", () => {
    const search: ToolCallBlock = {
      id: "web-search-2",
      type: "tool_call",
      name: "web.search",
      input: { query: "Saudi Arabia news today" },
      status: "success",
      output: { sources: [] },
    };
    const { unmount } = render(<ToolCallCard call={search} />);
    expect(
      screen.getByRole("button", {
        name: 'Completed: search the web for "Saudi Arabia news today"',
      }),
    ).not.toBeNull();
    unmount();

    const agents: ToolCallBlock = {
      id: "agents-1",
      type: "tool_call",
      name: "agents.run",
      input: {},
      status: "success",
      output: {
        reports: [
          { task: "Compare supplier A", status: "done", report: "A is cheaper." },
          { task: "Compare supplier B", status: "error", report: "" },
        ],
      },
    };
    render(<ToolCallCard call={agents} />);
    fireEvent.click(screen.getByRole("button", { name: /work on it in parallel/i }));
    expect(screen.getByText("Done: Compare supplier A")).not.toBeNull();
    expect(screen.getByText("A is cheaper.")).not.toBeNull();
    expect(screen.getByText("Could not finish: Compare supplier B")).not.toBeNull();
  });
});
