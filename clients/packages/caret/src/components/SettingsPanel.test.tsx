import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { useCaretSettingsStore } from "../stores/settingsStore";
import SettingsPanel from "./SettingsPanel";

describe("SettingsPanel", () => {
  beforeEach(() => {
    useCaretSettingsStore.setState({
      mode: "ask",
      voiceEnabled: true,
      autoApproveReadOnly: true,
      showThinking: true,
    });
  });

  it("does not expose agent types as a user-selectable mode", () => {
    render(<SettingsPanel />);

    expect(
      screen.queryByRole("combobox", { name: "Response style" }),
    ).toBeNull();
    expect(screen.getByRole("heading", { name: "Behavior" })).toBeTruthy();
  });

  it("uses full-row accessible switches for behavior settings", () => {
    render(<SettingsPanel />);

    const voiceSwitch = screen.getByRole("switch", { name: /Voice input/ });
    expect(voiceSwitch.getAttribute("aria-checked")).toBe("true");

    fireEvent.click(voiceSwitch);

    expect(useCaretSettingsStore.getState().voiceEnabled).toBe(false);
    expect(voiceSwitch.getAttribute("aria-checked")).toBe("false");
  });
});
