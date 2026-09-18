/**
 * The browser tab title for a screen.
 *
 * Thirty-nine screens spelled out `"<name> · rtylr"` as a literal, which put
 * the product name in thirty-nine places and the screen name in English in all
 * of them. Most of those names already existed, translated, in `nav.json`, so
 * the title is built from the same key the sidebar uses and the two cannot
 * drift apart (rtylr#1266).
 */
export function pageTitle(name: string): string {
  return `${name} · rtylr`;
}
