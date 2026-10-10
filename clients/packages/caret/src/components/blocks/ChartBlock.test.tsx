import React from "react";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import Markdown from "./Markdown";
import { parseChartSpec } from "./ChartBlock";

describe("chart blocks", () => {
  it("draws a chart from a chart code block", () => {
    const source = [
      "Revenue rose all week.",
      "```chart",
      '{"type":"bar","title":"Daily revenue","unit":"USD","labels":["Mon","Tue","Wed"],"series":[{"name":"Revenue","values":[120,180,240]}]}',
      "```",
    ].join("\n");
    const { container } = render(<Markdown source={source} />);
    expect(screen.getByRole("img", { name: "Daily revenue" })).not.toBeNull();
    expect(container.querySelectorAll("rect").length).toBe(3);
    expect(container.querySelector("pre")).toBeNull();
  });

  it("falls back to code for a malformed or partial chart", () => {
    const { container } = render(<Markdown source={'```chart\n{"type":"bar","labels":["a"\n```'} />);
    expect(container.querySelector("pre")).not.toBeNull();
  });

  it("rejects series that don't match the labels", () => {
    expect(parseChartSpec('{"labels":["a","b"],"series":[{"name":"x","values":[1]}]}')).toBeNull();
    expect(parseChartSpec('{"type":"line","labels":["a","b"],"series":[{"name":"x","values":[1,"2"]}]}')?.type).toBe("line");
  });
});
