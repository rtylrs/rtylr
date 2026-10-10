import { describe, expect, it } from "vitest";
import { replySources } from "./searchSources";
import type { ContentBlock } from "../types/message";

const search = (urls: string[]): ContentBlock => ({
  id: "tool",
  type: "tool_call",
  name: "web.search",
  input: {},
  status: "success",
  output: {
    sources: urls.map((url, index) => ({ title: `Result ${index}`, url })),
  },
});

describe("replySources", () => {
  it("lists cited sources first, dedupes, and drops unsafe links", () => {
    const sources = replySources([
      search([
        "https://www.example.com/a",
        "https://nypost.com/story",
        "http://insecure.test/x",
        "https://www.example.com/a",
      ]),
      { id: "text", type: "text", content: "Per nypost.com, things happened." },
    ]);

    expect(sources.map((source) => source.host)).toEqual([
      "nypost.com",
      "example.com",
    ]);
  });

  it("ignores failed searches", () => {
    const failed = { ...search(["https://a.test/"]), status: "error" as const };
    expect(replySources([failed])).toEqual([]);
  });
});

describe("replySources from other tools", () => {
  it("collects sources from page reads and sub-agents too", () => {
    const sources = replySources([
      {
        id: "a",
        type: "tool_call",
        name: "agents.run",
        status: "success",
        input: {},
        output: { sources: [{ title: "Report", url: "https://example.com/r" }] },
      },
      {
        id: "b",
        type: "tool_call",
        name: "web.fetch",
        status: "success",
        input: {},
        output: { sources: [{ title: "Page", url: "https://example.org/p" }] },
      },
    ] as never);
    expect(sources.map((source) => source.host)).toEqual([
      "example.com",
      "example.org",
    ]);
  });
});
