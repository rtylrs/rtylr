import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { AppIconTile, AutomationsIcon, CrmIcon, FlowIcon, appIconAccentStyle } from "./index";

const accentGroup = (container: HTMLElement) => container.querySelector("svg > g") as SVGGElement;

describe("app icons", () => {
  it("draws the accent as a quiet tint of the icon colour by default", () => {
    const { container } = render(<CrmIcon size={20} />);
    const svg = container.querySelector("svg")!;
    expect(svg.getAttribute("width")).toBe("20");
    expect(svg.getAttribute("aria-hidden")).toBe("true");
    expect(accentGroup(container).style.fill).toBe("var(--app-icon-accent, currentColor)");
  });

  it("colours the accent when given the app colour", () => {
    const { container } = render(<CrmIcon accent="#8b5cf6" />);
    expect(accentGroup(container).style.fill).toBe("#8b5cf6");
    expect(accentGroup(container).style.fillOpacity).toBe("1");
  });

  it("strokes line accents instead of filling them", () => {
    const { container } = render(<FlowIcon accent="#ef4444" />);
    expect(accentGroup(container).style.stroke).toBe("#ef4444");
    expect(accentGroup(container).style.fill).toBe("");
  });

  it("gives Automations its own glyph", () => {
    const flow = render(<FlowIcon />).container.innerHTML;
    const automations = render(<AutomationsIcon />).container.innerHTML;
    expect(automations).not.toBe(flow);
  });

  it("sets the accent variables a parent passes down", () => {
    expect(appIconAccentStyle("var(--app-hr)")).toMatchObject({
      "--app-icon-accent": "var(--app-hr)",
      "--app-icon-accent-opacity": "1",
    });
    const { container } = render(<AppIconTile icon={CrmIcon} color="var(--app-crm)" size={32} />);
    const tile = container.firstElementChild as HTMLElement;
    expect(tile.style.background).toBe("var(--app-crm)");
    expect(tile.style.color).toBe("#fff");
  });
});
