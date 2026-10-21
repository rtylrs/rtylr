/**
 * The browser tab title for a screen.
 *
 * Seventeen screens spelled out `"<name> · rtylr"` as a literal, which put the
 * product name in seventeen places and the screen name in English in all of
 * them. Where the sidebar already names a screen, the title reads the same key
 * so the two cannot drift apart (rtylr#1266).
 */
export function pageTitle(name: string): string {
  return `${name} · rtylr`;
}
