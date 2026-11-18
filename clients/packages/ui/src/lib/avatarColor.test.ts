import { describe, expect, it } from "vitest";
import { AVATAR_PALETTE, avatarColor, avatarColorIndex } from "./avatarColor";

/** WCAG relative luminance of a #RRGGBB colour. */
function luminance(hex: string): number {
  const channels = [1, 3, 5].map(
    (i) => parseInt(hex.slice(i, i + 2), 16) / 255,
  );
  const [r, g, b] = channels.map((c) =>
    c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4,
  );
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

describe("avatar colour", () => {
  it("is stable for a seed and ignores case and spaces", () => {
    expect(avatarColor("1042")).toEqual(avatarColor("1042"));
    expect(avatarColorIndex(" Abed ")).toBe(avatarColorIndex("abed"));
  });

  it("gives the five people in the reported strip five different colours", () => {
    // The ids behind AA, AH, AT, MH and MI in the Investment Hub strip are not
    // known here, so check the property that matters on realistic small ids.
    const indexes = new Set(
      ["1", "2", "3", "4", "5", "6", "7"].map(avatarColorIndex),
    );
    expect(indexes.size).toBeGreaterThanOrEqual(6);
  });

  it("spreads a team of 120 ids across every colour", () => {
    const counts = new Array(AVATAR_PALETTE.length).fill(0);
    for (let id = 1; id <= 120; id += 1)
      counts[avatarColorIndex(String(id))] += 1;
    expect(Math.min(...counts)).toBeGreaterThan(0);
  });

  it("keeps white initials readable (WCAG AA 4.5:1) on every colour", () => {
    for (const hex of AVATAR_PALETTE) {
      const ratio = 1.05 / (luminance(hex) + 0.05);
      expect(ratio, hex).toBeGreaterThanOrEqual(4.5);
    }
  });
});
