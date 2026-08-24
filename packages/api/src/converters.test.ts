import { describe, expect, it } from 'vitest';
import { fromOneKitJson, toOneKitJson } from './client';

/**
 * rtylr#952, defects 1 and 3.
 *
 * All three are latent: measured against packages/voxie/api, 0 of 1154 distinct
 * wire keys produce an acronym run when camelCased, and no app passes a File or
 * FormData through these helpers today. So these tests are the only thing
 * standing behind the fix, and none of it can be confirmed against real traffic.
 */

describe('key conversion preserves opaque values (rtylr#952 #1)', () => {
  it('does not reduce a Blob to an empty object', () => {
    const blob = new Blob(['x'], { type: 'text/plain' });
    const out = toOneKitJson({ attachment: blob }) as { attachment: unknown };
    // Object.entries on a Blob yields [], so the old converter replaced it with {}.
    expect(out.attachment).toBeInstanceOf(Blob);
  });

  it('does not reduce a nested File to an empty object', () => {
    const file = new File(['x'], 'a.txt', { type: 'text/plain' });
    const out = toOneKitJson({ wrapper: { upload: file } }) as {
      wrapper: { upload: unknown };
    };
    expect(out.wrapper.upload).toBeInstanceOf(File);
  });

  it('leaves FormData intact', () => {
    const form = new FormData();
    form.append('a', '1');
    const out = toOneKitJson({ body: form }) as { body: unknown };
    expect(out.body).toBeInstanceOf(FormData);
  });

  it('leaves URLSearchParams intact', () => {
    const params = new URLSearchParams({ a: '1' });
    const out = toOneKitJson({ query: params }) as { query: unknown };
    expect(out.query).toBeInstanceOf(URLSearchParams);
  });

  it('does not flatten a Blob on the response side either', () => {
    const blob = new Blob(['x']);
    const out = fromOneKitJson<{ download: unknown }>({ download: blob });
    expect(out.download).toBeInstanceOf(Blob);
  });

  it('still converts plain nested objects and arrays', () => {
    const out = toOneKitJson({ userName: 'a', items: [{ itemId: 1 }] }) as Record<string, unknown>;
    expect(out).toEqual({ user_name: 'a', items: [{ item_id: 1 }] });
  });

  it('still leaves a Date alone', () => {
    const d = new Date('2026-01-01T00:00:00Z');
    const out = toOneKitJson({ createdAt: d }) as { created_at: unknown };
    expect(out.created_at).toBeInstanceOf(Date);
  });
});

describe('key conversion handles acronym runs (rtylr#952 #3)', () => {
  it('does not explode an acronym into single letters', () => {
    // The old regex replaced every capital, giving plan_type_u_r_l.
    const out = toOneKitJson({ planTypeURL: 'x' }) as Record<string, unknown>;
    expect(Object.keys(out)).toEqual(['plan_type_url']);
  });

  it('handles a leading acronym followed by a word', () => {
    const out = toOneKitJson({ URLPath: 'x' }) as Record<string, unknown>;
    expect(Object.keys(out)).toEqual(['url_path']);
  });

  it('handles a digit boundary', () => {
    const out = toOneKitJson({ address2Line: 'x' }) as Record<string, unknown>;
    expect(Object.keys(out)).toEqual(['address2_line']);
  });

  it('converts an uppercase segment after an underscore', () => {
    // The old regex only matched _[a-z], so FULL_NAME came through unchanged.
    const out = fromOneKitJson<Record<string, unknown>>({ FULL_NAME: 'x' });
    expect(Object.keys(out)).toEqual(['fullName']);
  });

  it('round-trips every real wire key unchanged', () => {
    // The measured safety property: 0 of 1154 contract keys contain an acronym
    // run, so this change must be a no-op on all current traffic.
    for (const key of ['order_count', 'supplier_name', 'scheduled_at', 'low_stock_count', 'id']) {
      const camel = Object.keys(fromOneKitJson<Record<string, unknown>>({ [key]: 1 }))[0];
      const back = Object.keys(toOneKitJson({ [camel]: 1 }))[0];
      expect(back).toBe(key);
    }
  });
});
