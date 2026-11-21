import { describe, expect, it, vi } from 'vitest';
import {
  claimPager,
  isPhoneValid,
  isSlugUsable,
  normalisePhone,
  venueSlugCandidates,
} from './claimPager';

const taken = { statusCode: 409, code: 'conflict' };

function steps() {
  const calls: string[] = [];
  return {
    calls,
    savePhone: vi.fn(async (phone: string) => calls.push(`phone ${phone}`)),
    setSlug: vi.fn(async (slug: string) => calls.push(`slug ${slug}`)),
    enablePager: vi.fn(async () => calls.push('enable')),
  };
}

describe('venue slug candidates', () => {
  it('starts from the venue name and falls back to a short random ending', () => {
    expect(venueSlugCandidates("Café L'Étoile & Co", () => 'ab12')).toEqual([
      'cafe-l-etoile-co',
      'cafe-l-etoile-co-ab12',
      'cafe-l-etoile-co-ab12',
      'cafe-l-etoile-co-ab12',
      'cafe-l-etoile-co-ab12',
    ]);
  });

  it('never offers a reserved slug or an empty one', () => {
    expect(venueSlugCandidates('Menu', () => 'x1y2')[0]).toBe('menu-x1y2');
    expect(venueSlugCandidates('مقهى الحي', () => 'x1y2')[0]).toBe('venue-x1y2');
    for (const slug of venueSlugCandidates('مقهى الحي')) {
      expect(slug).toMatch(/^[a-z0-9][a-z0-9-]{0,78}[a-z0-9]$/);
    }
  });
});

describe('claiming the pager', () => {
  it('saves the phone, gives the business a slug, then switches the pager on', async () => {
    const claim = steps();
    await claimPager(claim, { phone: '+961 3 123 456', venueName: 'Bean Bar', slug: '' });
    expect(claim.calls).toEqual(['phone +961 3 123 456', 'slug bean-bar', 'enable']);
  });

  it('keeps a slug the business already has', async () => {
    const claim = steps();
    await claimPager(claim, { phone: '03123456', venueName: 'Bean Bar', slug: 'beans' });
    expect(claim.setSlug).not.toHaveBeenCalled();
    expect(claim.enablePager).toHaveBeenCalledOnce();
  });

  it('moves on to the next slug when one is taken', async () => {
    const claim = steps();
    claim.setSlug.mockRejectedValueOnce(taken);
    await claimPager(claim, { phone: '03123456', venueName: 'Bean Bar', slug: '' });
    expect(claim.setSlug).toHaveBeenCalledTimes(2);
    expect(claim.setSlug.mock.calls[1][0]).toMatch(/^bean-bar-[a-z0-9]{4}$/);
    expect(claim.enablePager).toHaveBeenCalledOnce();
  });

  it('uses only the slug the owner confirmed, and fails rather than pick another', async () => {
    const claim = steps();
    await claimPager(claim, {
      phone: '03123456',
      venueName: 'Brio Coffee',
      slug: '',
      chosenSlug: 'brio',
    });
    expect(claim.calls).toEqual(['phone 03123456', 'slug brio', 'enable']);

    const takenClaim = steps();
    takenClaim.setSlug.mockRejectedValueOnce(taken);
    await expect(
      claimPager(takenClaim, {
        phone: '03123456',
        venueName: 'Brio Coffee',
        slug: '',
        chosenSlug: 'brio',
      })
    ).rejects.toBe(taken);
    expect(takenClaim.setSlug).toHaveBeenCalledOnce();
    expect(takenClaim.enablePager).not.toHaveBeenCalled();
  });

  it('stops without switching the pager on when a step fails for another reason', async () => {
    const claim = steps();
    claim.setSlug.mockRejectedValue({ statusCode: 403 });
    await expect(
      claimPager(claim, { phone: '03123456', venueName: 'Bean Bar', slug: '' })
    ).rejects.toMatchObject({ statusCode: 403 });
    expect(claim.setSlug).toHaveBeenCalledOnce();
    expect(claim.enablePager).not.toHaveBeenCalled();
  });
});

describe('a slug the owner types', () => {
  it.each(['brio', 'brio-coffee', 'b2'])('accepts %s', (slug) => {
    expect(isSlugUsable(slug)).toBe(true);
  });

  it.each(['', '-brio', 'brio-', 'Brio', 'brio coffee', 'board', 'menu', 'a'.repeat(61)])(
    'refuses %j',
    (slug) => {
      expect(isSlugUsable(slug)).toBe(false);
    }
  );
});

describe('venue phone number', () => {
  it('turns Arabic-Indic digits into ASCII and drops anything that is not part of a number', () => {
    expect(normalisePhone(' +٩٦١ ٣ (123) 456 ')).toBe('+961 3 123 456');
    expect(normalisePhone('۰۳۱۲۳۴۵۶')).toBe('03123456');
  });

  it('needs between 6 and 15 digits', () => {
    expect(isPhoneValid('12345')).toBe(false);
    expect(isPhoneValid('+961 3 123 456')).toBe(true);
    expect(isPhoneValid('1234567890123456')).toBe(false);
  });
});
