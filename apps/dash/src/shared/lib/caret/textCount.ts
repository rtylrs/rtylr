/** Quick counts for composer footer (chars / words). */
export interface TextCounts {
  chars: number;
  words: number;
}

export function countText(text: string): TextCounts {
  const trimmed = text.trim();
  return {
    chars: text.length,
    words: trimmed ? trimmed.split(/\s+/).length : 0,
  };
}
