import type { ContentBlock } from "../types/message";

export interface SearchSource {
  title: string;
  url: string;
  host: string;
  snippet?: string;
}

/** Parse one web.search source, accepting only plain https links. */
export function toSearchSource(value: unknown): SearchSource | null {
  if (!value || typeof value !== "object") return null;
  const item = value as Record<string, unknown>;
  if (typeof item.title !== "string" || typeof item.url !== "string")
    return null;
  if (item.url.length > 2048) return null;
  let url: URL;
  try {
    url = new URL(item.url);
  } catch {
    return null;
  }
  if (
    url.protocol !== "https:" ||
    !url.hostname ||
    url.username ||
    url.password
  )
    return null;
  return {
    title: item.title,
    url: url.href,
    host: url.hostname.replace(/^www\./, ""),
    snippet:
      typeof item.snippet === "string" && item.snippet.trim()
        ? item.snippet
        : undefined,
  };
}

/** Sources returned by a web.search tool output. */
export function searchOutputSources(output: unknown): SearchSource[] {
  if (!output || typeof output !== "object") return [];
  const sources = (output as Record<string, unknown>).sources;
  if (!Array.isArray(sources)) return [];
  return sources
    .map(toSearchSource)
    .filter((source): source is SearchSource => source !== null);
}

/**
 * Every source a reply's searches returned, deduplicated, with the ones the
 * answer actually links or names first.
 */
export function replySources(blocks: ContentBlock[]): SearchSource[] {
  const text = blocks
    .map((block) => (block.type === "text" ? block.content : ""))
    .join("\n")
    .toLowerCase();
  const seen = new Set<string>();
  const sources: SearchSource[] = [];
  for (const block of blocks) {
    if (
      block.type !== "tool_call" ||
      block.name !== "web.search" ||
      block.status !== "success"
    )
      continue;
    for (const source of searchOutputSources(block.output)) {
      if (seen.has(source.url)) continue;
      seen.add(source.url);
      sources.push(source);
    }
  }
  const cited = (source: SearchSource) =>
    text.includes(source.url.toLowerCase()) || text.includes(source.host);
  return [
    ...sources.filter(cited),
    ...sources.filter((source) => !cited(source)),
  ];
}
