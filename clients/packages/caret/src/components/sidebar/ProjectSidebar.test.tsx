import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { CaretProject } from "../../types/chat";
import ProjectSidebar from "./ProjectSidebar";

describe("ProjectSidebar", () => {
  it("saves a project with its selected icon, color, and shared context", async () => {
    const onSave = vi.fn(
      async (input) =>
        ({
          id: input.id || "project-1",
          name: input.name,
          icon: input.icon,
          color: input.color,
          context: input.context,
          updatedAt: "2026-10-03T00:00:00Z",
        }) satisfies CaretProject,
    );
    const onSelect = vi.fn();

    render(
      <ProjectSidebar
        projects={[]}
        activeId={null}
        onSelect={onSelect}
        onSave={onSave}
        onArchive={vi.fn(async () => undefined)}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Create project" }));
    fireEvent.change(screen.getByRole("textbox", { name: "Project name" }), {
      target: { value: "POS launch" },
    });
    fireEvent.click(screen.getByTitle("coffee"));
    fireEvent.click(screen.getByTitle("green"));
    fireEvent.change(
      screen.getByRole("textbox", { name: /Shared project context/ }),
      { target: { value: "Work on terminal setup." } },
    );
    fireEvent.click(screen.getByRole("button", { name: "Save project" }));

    await waitFor(() =>
      expect(onSave).toHaveBeenCalledWith({
        id: "",
        name: "POS launch",
        icon: "coffee",
        color: "green",
        context: "Work on terminal setup.",
      }),
    );
    await waitFor(() => expect(onSelect).toHaveBeenCalledWith("project-1"));
    expect(screen.queryByRole("dialog")).toBeNull();
  });
});
