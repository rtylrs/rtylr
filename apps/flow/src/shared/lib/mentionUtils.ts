import type { Member } from '../stores/flowStore';

/** Extracts @mention handles from a text string. */
export function extractMentions(text: string): string[] {
  const matches = text.match(/@[\w.-]+/g) ?? [];
  return [...new Set(matches)];
}

/** Returns true if the text contains at least one @mention. */
export function hasMentions(text: string): boolean {
  return /@[\w.-]+/.test(text);
}

/** Replaces @handle mentions in text with the member's display name. */
export function resolveMentions(text: string, members: Member[]): string {
  const byName = new Map(members.map((m) => [m.name.toLowerCase().replace(/\s+/g, '.'), m]));
  return text.replace(/@([\w.-]+)/g, (match, handle) => {
    const member = byName.get(handle.toLowerCase());
    return member ? `@${member.name}` : match;
  });
}

/** Returns members whose name matches any mention in the text. */
export function membersForMentions(text: string, members: Member[]): Member[] {
  const mentions = extractMentions(text).map((m) => m.slice(1).toLowerCase());
  return members.filter((m) => mentions.includes(m.name.toLowerCase().replace(/\s+/g, '.')));
}

/** Builds an @mention string for a member (first.last format). */
export function buildMention(member: Member): string {
  return `@${member.name.replace(/\s+/g, '.').toLowerCase()}`;
}

/** Returns a list of members sorted by name for a mention dropdown. */
export function mentionCandidates(members: Member[], query: string): Member[] {
  const q = query.toLowerCase();
  return members
    .filter((m) => m.name.toLowerCase().includes(q) || m.email.toLowerCase().startsWith(q))
    .sort((a, b) => a.name.localeCompare(b.name));
}
