import { describe, expect, it } from 'vitest';
import { MOODS, moodDataUri, moodSvg } from './moods';
import { renderIllustration } from './illustrated';

describe('RTYLR Moods', () => {
  it('has twelve faces with unique ids the server accepts', () => {
    expect(MOODS).toHaveLength(12);
    expect(new Set(MOODS.map((m) => m.id)).size).toBe(12);
    for (const mood of MOODS) expect(mood.id).toMatch(/^[a-z-]{1,64}$/);
  });

  it('draws every face as a well-formed SVG', () => {
    for (const mood of MOODS) {
      const svg = moodSvg(mood.id) ?? '';
      const doc = new DOMParser().parseFromString(svg, 'image/svg+xml');
      expect(doc.querySelector('parsererror'), mood.id).toBeNull();
      expect(doc.documentElement.getAttribute('viewBox')).toBe('0 0 100 100');
    }
  });

  it('renders through the illustrated path, and nothing for an unknown mood', async () => {
    expect(await renderIllustration('rtylr-moods', 'galaxy-brain')).toBe(
      moodDataUri('galaxy-brain')
    );
    await expect(renderIllustration('rtylr-moods', 'pepe')).rejects.toThrow();
    expect(moodDataUri('pepe')).toBeUndefined();
  });
});
