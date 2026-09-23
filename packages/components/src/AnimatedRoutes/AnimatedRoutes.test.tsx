import React from "react";
import { render } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const presenceState = vi.hoisted(() => ({
  modes: [] as string[],
  keys: [] as string[],
}));

vi.mock("motion/react", () => {
  const Presence = ({
    children,
    mode,
  }: {
    children: React.ReactNode;
    mode?: string;
  }): React.JSX.Element => {
    const child = React.isValidElement(children) ? children : null;
    presenceState.modes.push(mode ?? "");
    presenceState.keys.push(child?.key == null ? "" : String(child.key));
    return <div data-testid="animate-presence">{children}</div>;
  };

  const MotionDiv = ({
    children,
  }: React.PropsWithChildren): React.JSX.Element => <div>{children}</div>;

  return {
    AnimatePresence: Presence,
    motion: { div: MotionDiv },
  };
});

import AnimatedRoutes from "./index";

describe("AnimatedRoutes", () => {
  beforeEach(() => {
    presenceState.modes.length = 0;
    presenceState.keys.length = 0;
  });

  it("keeps route content in presence and remounts it when the animation key changes", () => {
    const view = render(
      <AnimatedRoutes animationKey="/login">Login</AnimatedRoutes>,
    );

    expect(presenceState.modes).toEqual(["wait"]);
    expect(presenceState.keys).toEqual(["/login"]);

    view.rerender(
      <AnimatedRoutes animationKey="/register">Register</AnimatedRoutes>,
    );

    expect(presenceState.keys).toEqual(["/login", "/register"]);
  });
});
