import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import Markdown from "./Markdown";
import { CaretHostProvider, type CaretPageSpec } from "../../host";

describe("Markdown tables", () => {
  it("replaces a header-only table with an empty result message", () => {
    render(
      <Markdown
        source={`| Customer | Total Spend |
| --- | --- |`}
      />,
    );

    expect(screen.getByText("No matching records.")).toBeTruthy();
    expect(document.querySelector("table")).toBeNull();
  });

  it("keeps tables that contain data rows", () => {
    render(
      <Markdown
        source={`| Customer | Total Spend |
| --- | --- |
| Acme | $100 |`}
      />,
    );

    expect(screen.getByText("Acme")).toBeTruthy();
    expect(document.querySelector("table.caret-table")).toBeTruthy();
  });
});

describe("Markdown code highlighting", () => {
  const fence = (language: string, code: string): string =>
    "```" + language + "\n" + code + "\n```";

  it("colours the tokens of a fenced block and keeps its text intact", async () => {
    const code = "def total(xs):\n    return round(sum(xs), 2)  # sum";
    render(<Markdown source={fence("python", code)} />);

    await waitFor(() =>
      expect(document.querySelector(".hljs-keyword")).toBeTruthy(),
    );
    expect(document.querySelector(".hljs-keyword")?.textContent).toBe("def");
    expect(document.querySelector(".hljs-comment")?.textContent).toBe("# sum");
    expect(document.querySelector("pre code")?.textContent).toBe(code + "\n");
  });

  it("leaves unknown languages and unlabelled blocks plain", async () => {
    render(
      <Markdown
        source={fence("nonsense-lang", "def x") + "\n\n" + fence("", "def y")}
      />,
    );
    // Wait for any highlighter load to settle before asserting nothing coloured.
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(document.querySelector('[class*="hljs-"]')).toBeNull();
    expect(document.querySelectorAll("pre code")).toHaveLength(2);
  });

  it("shows markup inside code as text, never as elements", async () => {
    render(<Markdown source={fence("html", "<img src=x onerror=alert(1)>")} />);
    await waitFor(() =>
      expect(document.querySelector(".hljs-tag")).toBeTruthy(),
    );
    expect(document.querySelector("pre img")).toBeNull();
    expect(document.querySelector("pre code")?.textContent).toContain(
      "<img src=x onerror=alert(1)>",
    );
  });
});

describe("Markdown code block copy button", () => {
  it("copies the full text of a fenced block, even once highlighted", async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      value: { writeText },
      configurable: true,
    });
    const code = "def total(xs):\n    return sum(xs)  # sum";
    render(<Markdown source={"```python\n" + code + "\n```"} />);

    await waitFor(() =>
      expect(document.querySelector(".hljs-keyword")).toBeTruthy(),
    );
    fireEvent.click(screen.getByRole("button", { name: "Copy code" }));

    await waitFor(() => expect(writeText).toHaveBeenCalledWith(code));
    expect(await screen.findByRole("button", { name: "Copied" })).toBeTruthy();
  });

  it("does not add a copy button to inline code", () => {
    render(<Markdown source="Call `summarize` first." />);
    expect(screen.queryByRole("button", { name: "Copy code" })).toBeNull();
  });
});

describe("Markdown page blocks", () => {
  const block = '```page\n{"path":"/team","action":"Invite member"}\n```';

  it("lets the host app render the page it names", () => {
    const renderPage = vi.fn(({ path, action }: CaretPageSpec) => (
      <span>
        card {path} {action}
      </span>
    ));
    render(
      <CaretHostProvider value={{ renderPage }}>
        <Markdown source={`Team members are added here:\n\n${block}`} />
      </CaretHostProvider>,
    );
    expect(screen.getByText("card /team Invite member")).toBeTruthy();
    expect(document.querySelector("pre")).toBeNull();
  });

  it("falls back to a link without a host, and refuses other sites", () => {
    render(<Markdown source={block} />);
    expect(
      screen.getByRole("link", { name: "/team" }).getAttribute("href"),
    ).toBe("/team");
    render(<Markdown source={'```page\n{"path":"//evil.example"}\n```'} />);
    expect(screen.queryByRole("link", { name: "//evil.example" })).toBeNull();
  });

  it("opens in-app links through the host so the chat stays open", () => {
    const openPage = vi.fn();
    render(
      <CaretHostProvider value={{ openPage }}>
        <Markdown source="Go to [Team](/team)." />
      </CaretHostProvider>,
    );
    fireEvent.click(screen.getByRole("link", { name: "Team" }));
    expect(openPage).toHaveBeenCalledWith({ path: "/team" });
  });
});
