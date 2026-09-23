/**
 * The contrast sweep lives here, and it is the test that makes the design's
 * accessibility claim real rather than asserted.
 *
 * It is a real loop over the real data: every palette in PALETTES times every
 * accent in ACCENT_SWATCHES plus every hex in STRESS_ACCENTS. 8 times 14 is 112
 * combinations, and the count itself is asserted so that deleting a fixture
 * cannot quietly shrink the sweep.
 *
 * Measured results, from running this file:
 *   112 combinations, zero failures
 *   worst --accent-fg     4.6010:1  (linen card #fffefb, accent #8b8cf5)
 *   worst --flow-on-accent 6.6117:1 (accent #b57ce8, near-black)
 * The design predicted a worst case of about 4.60:1. It is 4.6010:1.
 *
 * Mutation probe on the derivation gate, run and recorded because the repo
 * convention requires it. The gate is the `>= CONTRAST_TARGET` comparison in
 * deriveAccentFg, forced to a constant in both directions:
 *   - forced true : the walk never runs, so --accent-fg is the raw accent. 56 of
 *                   112 fall below the floor, worst 1.0625:1 (graphite card
 *                   #202027, accent #2d1b4e). The `failures` assertion below
 *                   fails, and 4 tests fail in total across both files.
 *   - forced false: the walk always exhausts and returns pure black or white.
 *                   The `failures` assertion STILL PASSES, because black on a
 *                   dark card is 16.19:1. What catches it is the recorded
 *                   worst-case bound below, and the hue-preservation test in
 *                   colorMath.test.ts, which measures 172.6 degrees of drift.
 *                   6 tests fail in total.
 * So the floor sweep alone does not prove the gate is alive: it passes over a
 * dead branch in one direction. That is exactly the failure mode the convention
 * exists to catch, and it is why the hue test and the worst-case bound are
 * load-bearing rather than decorative.
 */

import { describe, it, expect } from "vitest";
import { appearanceTokens } from "./appearanceTokens";
import {
  compositeOver,
  contrastRatio,
  hexToHsl,
  isValidHex,
  NEAR_BLACK,
  NEAR_WHITE,
  rgbToHsl,
} from "./colorMath";
import {
  ACCENT_DEFAULT,
  ACCENT_SWATCHES,
  PALETTES,
  STRESS_ACCENTS,
} from "./palettes";
import { CONTRAST_FLOOR } from "./types";
import type {
  AppearanceContext,
  AppearancePrefs,
  PalettePreset,
} from "./types";

const BASE_PREFS: AppearancePrefs = {
  mode: "dark",
  palette: "midnight",
  accent: null,
  density: "cozy",
  textSize: "m",
  corners: "soft",
  reduceMotion: false,
};

/** The 14 fixtures. Curated swatches alone would pass while proving nothing. */
const ALL_ACCENTS: readonly string[] = [
  ...ACCENT_SWATCHES.map((swatch) => swatch.hex),
  ...STRESS_ACCENTS,
];

function tokensFor(
  palette: PalettePreset,
  accent: string | null,
): Record<string, string> {
  return appearanceTokens(
    { ...BASE_PREFS, mode: palette.mode, palette: palette.id, accent },
    { resolvedMode: palette.mode, workspaceAccent: ACCENT_DEFAULT },
  );
}

// ── The contrast sweep ──────────────────────────────────────────────

describe("the contrast sweep", () => {
  it("clears the floor on both derived foregrounds for all 112 combinations", () => {
    const failures: string[] = [];
    let combinations = 0;
    let worstFg = Infinity;
    let worstFgAt = "";
    let worstOn = Infinity;
    let worstOnAt = "";

    for (const palette of PALETTES) {
      for (const accent of ALL_ACCENTS) {
        combinations += 1;
        const tokens = tokensFor(palette, accent);

        // --accent-fg is measured against the card, because that is the ground
        // the text using it actually sits on.
        const fgRatio = contrastRatio(
          tokens["--accent-fg"],
          tokens["--flow-card"],
        );
        // --flow-on-accent is measured against the accent fill itself.
        const onRatio = contrastRatio(
          tokens["--flow-on-accent"],
          tokens["--accent"],
        );

        if (fgRatio < worstFg) {
          worstFg = fgRatio;
          worstFgAt = `${palette.id} card ${palette.card}, accent ${accent}`;
        }
        if (onRatio < worstOn) {
          worstOn = onRatio;
          worstOnAt = `accent ${accent}, on ${tokens["--flow-on-accent"]}`;
        }
        if (fgRatio < CONTRAST_FLOOR) {
          failures.push(
            `--accent-fg ${tokens["--accent-fg"]} on ${palette.id} card: ${fgRatio.toFixed(4)}`,
          );
        }
        if (onRatio < CONTRAST_FLOOR) {
          failures.push(
            `--flow-on-accent ${tokens["--flow-on-accent"]} on ${accent}: ${onRatio.toFixed(4)}`,
          );
        }
      }
    }

    // The count is asserted so that shrinking the fixture list breaks the test
    // rather than quietly narrowing what it proves.
    expect(combinations).toBe(112);
    expect(PALETTES).toHaveLength(8);
    expect(ALL_ACCENTS).toHaveLength(14);
    expect(failures).toEqual([]);

    // Recorded worst cases. These are upper bounds on the real measurements, so
    // a regression that erodes the margin shows up here before it reaches AA.
    expect(worstFg).toBeGreaterThanOrEqual(CONTRAST_FLOOR);
    expect(worstFg).toBeCloseTo(4.601, 2);
    expect(worstFgAt).toContain("#8b8cf5");
    expect(worstOn).toBeGreaterThanOrEqual(CONTRAST_FLOOR);
    expect(worstOn).toBeCloseTo(6.612, 2);
    expect(worstOnAt).toContain("#b57ce8");
  });

  it("picks white on accent exactly where near-black would have failed", () => {
    // The four stress hexes are the only fixtures that exercise the
    // --flow-on-accent failure at all. Every other fixture keeps near-black.
    const flipped = ALL_ACCENTS.filter(
      (accent) =>
        tokensFor(PALETTES[0], accent)["--flow-on-accent"] === NEAR_WHITE,
    );
    expect([...flipped].sort()).toEqual([...STRESS_ACCENTS].sort());

    for (const accent of ACCENT_SWATCHES) {
      expect(tokensFor(PALETTES[0], accent.hex)["--flow-on-accent"]).toBe(
        NEAR_BLACK,
      );
    }
  });
});

// ── Reproducing globals.css ─────────────────────────────────────────

describe("the shipped dark defaults", () => {
  const tokens = tokensFor(
    PALETTES.find((p) => p.id === "midnight") as PalettePreset,
    ACCENT_DEFAULT,
  );

  it("reproduces apps/flow/src/globals.css exactly", () => {
    expect(tokens["--flow-bg"]).toBe("#0a0a0a");
    expect(tokens["--flow-surface"]).toBe("#111111");
    expect(tokens["--flow-card"]).toBe("#161616");
    expect(tokens["--flow-border"]).toBe("rgba(255, 255, 255, 0.07)");
    expect(tokens["--flow-text"]).toBe("rgba(240, 240, 238, 0.92)");
    expect(tokens["--flow-text-muted"]).toBe("rgba(240, 240, 238, 0.58)");
    expect(tokens["--flow-text-dim"]).toBe("rgba(240, 240, 238, 0.28)");
    expect(tokens["--flow-hover"]).toBe("rgba(255, 255, 255, 0.05)");
    expect(tokens["--flow-input-bg"]).toBe("rgba(255, 255, 255, 0.04)");
  });

  it("reproduces the accentTints family exactly", () => {
    expect(tokens["--accent"]).toBe("#64c5b9");
    expect(tokens["--accent-muted"]).toBe("rgba(100, 197, 185, 0.12)");
    expect(tokens["--accent-border"]).toBe("rgba(100, 197, 185, 0.3)");
    expect(tokens["--accent-focus"]).toBe("rgba(100, 197, 185, 0.5)");
  });

  it("leaves the hardcoded dark pair where it already was", () => {
    // Dark mode is a no-op for both derived tokens with the shipped teal: the
    // accent already clears the target at 8.81:1, and near-black wins on it.
    // Deriving these must not re-theme the app that ships today.
    expect(tokens["--accent-fg"]).toBe("#64c5b9");
    expect(tokens["--flow-on-accent"]).toBe("#0a0a0a");
  });
});

describe("the shipped light defaults", () => {
  const tokens = tokensFor(
    PALETTES.find((p) => p.id === "paper") as PalettePreset,
    ACCENT_DEFAULT,
  );

  it("reproduces apps/flow/src/globals.css light mode", () => {
    expect(tokens["--flow-bg"]).toBe("#f4f4f5");
    expect(tokens["--flow-surface"]).toBe("#ffffff");
    expect(tokens["--flow-card"]).toBe("#ffffff");
    expect(tokens["--flow-border"]).toBe("rgba(0, 0, 0, 0.08)");
    // globals.css writes this one as `rgb(9, 9, 11)`, with no alpha channel at
    // all. An explicit alpha of 1 is the same colour, and keeping one format for
    // all three text tokens is worth more than matching that literal.
    expect(tokens["--flow-text"]).toBe("rgba(9, 9, 11, 1)");
    expect(tokens["--flow-text-muted"]).toBe("rgba(9, 9, 11, 0.62)");
    expect(tokens["--flow-text-dim"]).toBe("rgba(9, 9, 11, 0.35)");
    expect(tokens["--flow-hover"]).toBe("rgba(0, 0, 0, 0.05)");
    // Light mode declares the surface rather than a wash over it.
    expect(tokens["--flow-input-bg"]).toBe("#ffffff");
  });

  it("runs the accent washes heavier, as accentTints does", () => {
    expect(tokens["--accent-muted"]).toBe("rgba(100, 197, 185, 0.18)");
    // 0.42, from accentTints. globals.css also declares --accent-border at 0.3
    // in its mode-agnostic :root block and never overrides it for light, but
    // that declaration is dead: accentTints writes the property inline on
    // documentElement, which outranks any stylesheet rule.
    expect(tokens["--accent-border"]).toBe("rgba(100, 197, 185, 0.42)");
    expect(tokens["--accent-focus"]).toBe("rgba(100, 197, 185, 0.5)");
  });

  it("darkens the accent for light mode instead of shipping a hardcoded teal", () => {
    // The shipped light value is #1f7d72 at 4.96:1. The walk lands one stop
    // brighter at the same hue, which is the cost of stopping at the first
    // crossing rather than hand-picking.
    expect(tokens["--accent-fg"]).toBe("#308177");
    expect(
      contrastRatio(tokens["--accent-fg"], "#ffffff"),
    ).toBeGreaterThanOrEqual(CONTRAST_FLOOR);
  });
});

// ── Accent precedence ───────────────────────────────────────────────

describe("accent precedence", () => {
  const midnight = PALETTES.find((p) => p.id === "midnight") as PalettePreset;

  it("follows the workspace colour when the user accent is null", () => {
    const tokens = appearanceTokens(
      { ...BASE_PREFS, accent: null },
      { resolvedMode: "dark", workspaceAccent: "#a3d84b" },
    );
    expect(tokens["--accent"]).toBe("#a3d84b");
  });

  it("lets a user accent override the workspace colour", () => {
    const tokens = appearanceTokens(
      { ...BASE_PREFS, accent: "#f56d8f" },
      { resolvedMode: "dark", workspaceAccent: "#a3d84b" },
    );
    expect(tokens["--accent"]).toBe("#f56d8f");
  });

  it("falls back to ACCENT_DEFAULT when neither source is usable", () => {
    const tokens = appearanceTokens(
      { ...BASE_PREFS, accent: null },
      { resolvedMode: "dark", workspaceAccent: "" },
    );
    expect(tokens["--accent"]).toBe(ACCENT_DEFAULT);
    expect(midnight.card).toBe("#161616");
  });
});

// ── Palette resolution ─────────────────────────────────────────────

describe("palette resolution", () => {
  it("falls back to the mode family when the stored id belongs to the other one", () => {
    // Switching mode does not rewrite the stored palette id, so a light id in
    // dark mode is a normal state rather than corruption.
    const tokens = appearanceTokens(
      { ...BASE_PREFS, palette: "paper" },
      { resolvedMode: "dark", workspaceAccent: ACCENT_DEFAULT },
    );
    expect(tokens["--flow-bg"]).toBe("#0a0a0a");
  });

  it("falls back for an id that matches nothing at all", () => {
    const tokens = appearanceTokens(
      { ...BASE_PREFS, palette: "no-such-palette" },
      { resolvedMode: "light", workspaceAccent: ACCENT_DEFAULT },
    );
    expect(tokens["--flow-bg"]).toBe("#f4f4f5");
  });
});

// ── Shape tokens ───────────────────────────────────────────────────

describe("shape tokens", () => {
  it("emits a unitless multiplier for density and text size, a length for corners", () => {
    const at = (over: Partial<AppearancePrefs>): Record<string, string> =>
      appearanceTokens(
        { ...BASE_PREFS, ...over },
        { resolvedMode: "dark", workspaceAccent: ACCENT_DEFAULT },
      );

    // Unitless on purpose. A consumer multiplies its own rem base by these, so
    // spacing keeps following a browser font-size setting. Emitting px here
    // would remove that silently, because px cannot become rem in calc().
    expect(at({ density: "compact" })["--flow-u"]).toBe("0.75");
    expect(at({ density: "cozy" })["--flow-u"]).toBe("1");
    expect(at({ density: "roomy" })["--flow-u"]).toBe("1.25");
    expect(at({ textSize: "s" })["--flow-fs"]).toBe("0.92");
    expect(at({ textSize: "m" })["--flow-fs"]).toBe("1");
    expect(at({ textSize: "l" })["--flow-fs"]).toBe("1.1");
    // Corners stay a length. A radius has no reason to track font size.
    expect(at({ corners: "sharp" })["--flow-r"]).toBe("4px");
    expect(at({ corners: "soft" })["--flow-r"]).toBe("8px");
    expect(at({ corners: "round" })["--flow-r"]).toBe("14px");
  });

  it("cannot scale an input below the 16px iOS floor on its own", () => {
    // This token is a multiplier now, so it cannot express a font size at all,
    // let alone an unsafe one. The floor is therefore the consumer's problem:
    // an input sized `calc(16px * var(--flow-fs))` would still drop to 14.7px
    // at the small step and auto-zoom on iOS. The panel's hex field is sized
    // with a literal 16px for exactly that reason. Recording the constraint
    // where the token is defined, rather than asserting a safety this token
    // does not provide.
    const smallest = appearanceTokens(
      { ...BASE_PREFS, textSize: "s" },
      { resolvedMode: "dark", workspaceAccent: ACCENT_DEFAULT },
    )["--flow-fs"];
    expect(Number(smallest)).toBeLessThan(1);
    expect(Number.isNaN(Number(smallest))).toBe(false);
  });

  it("falls back rather than emitting undefined for a corrupt stored value", () => {
    const corrupt = {
      ...BASE_PREFS,
      density: "huge",
      textSize: "xl",
      corners: "blob",
    } as unknown as AppearancePrefs;
    const tokens = appearanceTokens(corrupt, {
      resolvedMode: "dark",
      workspaceAccent: ACCENT_DEFAULT,
    });
    expect(tokens["--flow-u"]).toBe("1");
    expect(tokens["--flow-fs"]).toBe("1");
    expect(tokens["--flow-r"]).toBe("8px");
  });

  it("falls back to dark tokens for a corrupt persisted mode", () => {
    const tokens = appearanceTokens(BASE_PREFS, {
      resolvedMode: "sepia",
      workspaceAccent: ACCENT_DEFAULT,
    } as unknown as AppearanceContext);

    expect(tokens["--flow-bg"]).toBe("#0a0a0a");
    expect(tokens["--flow-text"]).toBe("rgba(240, 240, 238, 0.92)");
  });
});

// ── Invalid input ──────────────────────────────────────────────────

describe("invalid accent input", () => {
  // '64c5b9' is in this list because the task called it invalid. It is not:
  // Flow's own isValidHex accepts a missing hash, and this mirrors it, so the
  // value is normalized rather than rejected. Recorded as an assertion so the
  // behaviour is deliberate rather than incidental.
  const cases: Array<[string, string]> = [
    ["", ACCENT_DEFAULT],
    ["nonsense", ACCENT_DEFAULT],
    ["#ggg", ACCENT_DEFAULT],
    ["#abc", ACCENT_DEFAULT],
    ["64c5b9", "#64c5b9"],
  ];

  for (const [input, expected] of cases) {
    it(`produces a complete, valid token set for ${JSON.stringify(input)}`, () => {
      const tokens = appearanceTokens(
        { ...BASE_PREFS, accent: input },
        { resolvedMode: "dark", workspaceAccent: ACCENT_DEFAULT },
      );

      expect(tokens["--accent"]).toBe(expected);
      expect(isValidHex(tokens["--accent"])).toBe(true);
      expect(isValidHex(tokens["--accent-fg"])).toBe(true);
      expect(isValidHex(tokens["--flow-on-accent"])).toBe(true);

      // Nothing may be missing, empty, or stringified garbage: a custom
      // property set to "undefined" or "NaN" is silently dropped by the
      // browser, which is the failure mode hardest to notice.
      for (const [key, value] of Object.entries(tokens)) {
        expect(key.startsWith("--")).toBe(true);
        expect(value.length).toBeGreaterThan(0);
        expect(value).not.toContain("undefined");
        expect(value).not.toContain("NaN");
      }

      // A fallback accent still has to clear the bar.
      expect(
        contrastRatio(tokens["--accent-fg"], tokens["--flow-card"]),
      ).toBeGreaterThanOrEqual(CONTRAST_FLOOR);
      expect(
        contrastRatio(tokens["--flow-on-accent"], tokens["--accent"]),
      ).toBeGreaterThanOrEqual(CONTRAST_FLOOR);
    });
  }

  it("does not throw for a null accent with an unusable workspace colour", () => {
    expect(() =>
      appearanceTokens(
        { ...BASE_PREFS, accent: null },
        { resolvedMode: "light", workspaceAccent: "#ggg" },
      ),
    ).not.toThrow();
  });
});

// ── Purity ─────────────────────────────────────────────────────────

describe("purity", () => {
  it("returns the same tokens for the same inputs and mutates neither argument", () => {
    const prefs: AppearancePrefs = { ...BASE_PREFS, accent: "#8b8cf5" };
    const ctx = {
      resolvedMode: "light",
      workspaceAccent: ACCENT_DEFAULT,
    } as const;
    const frozenPrefs = JSON.stringify(prefs);
    const frozenCtx = JSON.stringify(ctx);

    const first = appearanceTokens(prefs, ctx);
    const second = appearanceTokens(prefs, ctx);

    expect(first).toEqual(second);
    expect(JSON.stringify(prefs)).toBe(frozenPrefs);
    expect(JSON.stringify(ctx)).toBe(frozenCtx);
  });

  it("derives the focus ring from the accent rather than holding a brand colour", () => {
    // This token was hardcoded teal and was missed twice: once when --accent-fg
    // and --flow-on-accent were derived, and again until a browser check found a
    // teal ring drawn around a coral-accented card. The hue check is what makes
    // this a real assertion: emitting the token is not the same as deriving it.
    for (const palette of PALETTES) {
      for (const swatch of [
        ...ACCENT_SWATCHES.map((a) => a.hex),
        ...STRESS_ACCENTS,
      ]) {
        const tokens = appearanceTokens(
          { ...BASE_PREFS, palette: palette.id, accent: swatch },
          { resolvedMode: palette.mode, workspaceAccent: ACCENT_DEFAULT },
        );
        const ring = tokens["--flow-focus-ring"];
        expect(ring).toMatch(/^rgba\(/);
        // Hue must track the accent, which is the whole defect being fixed.
        const [r, g, b] = ring
          .match(/[\d.]+/g)!
          .slice(0, 3)
          .map(Number);
        const ringHue = rgbToHsl({ r, g, b }).h;
        const accentHue = hexToHsl(swatch).h;
        const drift = Math.min(
          Math.abs(ringHue - accentHue),
          360 - Math.abs(ringHue - accentHue),
        );
        expect(drift).toBeLessThan(8);
      }
    }
  });

  it("emits every token the design lists, and nothing undeclared", () => {
    const tokens = appearanceTokens(BASE_PREFS, {
      resolvedMode: "dark",
      workspaceAccent: ACCENT_DEFAULT,
    });
    expect(Object.keys(tokens).sort()).toEqual(
      [
        "--accent",
        "--accent-border",
        "--accent-fg",
        "--accent-focus",
        "--accent-muted",
        "--accent-on-muted",
        "--flow-bg",
        "--flow-border",
        "--flow-card",
        "--flow-focus-ring",
        "--flow-fs",
        "--flow-hover",
        "--flow-input-bg",
        "--flow-on-accent",
        "--flow-r",
        "--flow-surface",
        "--flow-text",
        "--flow-text-dim",
        "--flow-text-muted",
        "--flow-u",
      ].sort(),
    );
  });
});

/**
 * rtylr#970. --accent-fg was derived against palette.card and then used as text
 * on the --accent-muted tint, where 40 of 56 light-mode combinations measured
 * below 4.5:1. The guarantee did not transfer because the background was not the
 * one it was derived for.
 *
 * This sweeps every palette against every accent, including the stress hexes a
 * free-form picker allows, and checks the tint composited over all three grounds
 * it can sit on. Light mode is the half that fails, so a dark-only check would
 * have reported this feature as clean.
 */
describe("--accent-on-muted holds up on the tint, not just the card", () => {
  const MUTED_ALPHA: Record<"dark" | "light", number> = {
    dark: 0.12,
    light: 0.18,
  };
  const accents = [
    ...ACCENT_SWATCHES.map((a) => ({ id: a.id, hex: a.hex })),
    ...STRESS_ACCENTS.map((hex, i) => ({ id: `stress-${i}`, hex })),
  ];

  const cases = PALETTES.flatMap((palette) =>
    accents.map((accent) => ({ palette, accent })),
  );

  it("sweeps a meaningful number of combinations", () => {
    // Without this a bad fixture makes every assertion below vacuous.
    expect(cases.length).toBeGreaterThanOrEqual(100);
    expect(STRESS_ACCENTS.length).toBeGreaterThan(0);
  });

  it.each(cases)(
    "clears 4.5:1 on every ground: $palette.id / $accent.id",
    ({ palette, accent }) => {
      const tokens = appearanceTokens(
        {
          mode: palette.mode,
          palette: palette.id,
          accent: accent.hex,
          density: "cozy",
          textSize: "m",
          corners: "soft",
          reduceMotion: false,
        },
        { resolvedMode: palette.mode, workspaceAccent: accent.hex },
      );

      for (const ground of [palette.bg, palette.surface, palette.card]) {
        const tint = compositeOver(
          accent.hex,
          MUTED_ALPHA[palette.mode],
          ground,
        );
        expect(
          contrastRatio(tokens["--accent-on-muted"], tint),
          `${palette.id}/${accent.id} on a tint over ${ground}`,
        ).toBeGreaterThanOrEqual(CONTRAST_FLOOR);
      }
    },
  );

  it("is not simply --accent-fg under another name", () => {
    // If the two tokens were identical the sweep above would be re-testing
    // --accent-fg and would have failed, but assert the distinction directly so
    // a future refactor that aliases them fails here with a clear reason.
    const light = PALETTES.find((p) => p.mode === "light");
    expect(light).toBeTruthy();
    const tokens = appearanceTokens(
      {
        mode: "light",
        palette: light!.id,
        accent: "#e71046",
        density: "cozy",
        textSize: "m",
        corners: "soft",
        reduceMotion: false,
      },
      { resolvedMode: "light", workspaceAccent: "#e71046" },
    );
    expect(tokens["--accent-on-muted"]).not.toBe(tokens["--accent-fg"]);
  });
});
