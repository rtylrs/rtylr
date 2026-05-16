import { describe, expect, it } from 'vitest';
import { buildContextBlock } from './contextBlock';

describe('buildContextBlock', () => {
  it('includes the timezone line', () => {
    const out = buildContextBlock({ pathname: '/' });
    expect(out).toContain('Timezone:');
  });

  it('passes through the locale override', () => {
    const out = buildContextBlock({ pathname: '/', locale: 'fr-FR' });
    expect(out).toContain('Locale: fr-FR');
  });
});
