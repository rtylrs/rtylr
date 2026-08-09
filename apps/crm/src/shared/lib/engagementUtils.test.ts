import { describe, it, expect } from 'vitest';
import { recencyScore, frequencyScore, monetaryScore, engagementScore } from './engagementUtils';

describe('recencyScore', () => {
  it('returns 5 for visits within 7 days', () => expect(recencyScore(3)).toBe(5));
  it('returns 4 for visits within 30 days', () => expect(recencyScore(20)).toBe(4));
  it('returns 3 for visits within 90 days', () => expect(recencyScore(60)).toBe(3));
  it('returns 2 for visits within 180 days', () => expect(recencyScore(120)).toBe(2));
  it('returns 1 for visits over 180 days', () => expect(recencyScore(200)).toBe(1));
});

describe('frequencyScore', () => {
  it('returns 5 for 20+ visits', () => expect(frequencyScore(25)).toBe(5));
  it('returns 4 for 10-19 visits', () => expect(frequencyScore(12)).toBe(4));
  it('returns 3 for 5-9 visits', () => expect(frequencyScore(7)).toBe(3));
  it('returns 2 for 2-4 visits', () => expect(frequencyScore(3)).toBe(2));
  it('returns 1 for 1 visit', () => expect(frequencyScore(1)).toBe(1));
});

describe('monetaryScore', () => {
  it('returns 5 for spend >= 1000', () => expect(monetaryScore(1500)).toBe(5));
  it('returns 4 for spend >= 500', () => expect(monetaryScore(700)).toBe(4));
  it('returns 3 for spend >= 200', () => expect(monetaryScore(300)).toBe(3));
  it('returns 2 for spend >= 50', () => expect(monetaryScore(75)).toBe(2));
  it('returns 1 for spend < 50', () => expect(monetaryScore(20)).toBe(1));
});

describe('engagementScore', () => {
  it('returns champions tier for top scores', () => {
    const score = engagementScore(3, 25, 1500);
    expect(score.tier).toBe('champions');
    expect(score.total).toBe(15);
  });

  it('returns loyal tier for mid-high scores', () => {
    const score = engagementScore(20, 12, 300);
    expect(score.tier).toBe('loyal');
  });

  it('returns at-risk tier for mid scores', () => {
    const score = engagementScore(60, 3, 75);
    expect(score.tier).toBe('at-risk');
  });

  it('returns lost tier for low scores', () => {
    const score = engagementScore(200, 1, 20);
    expect(score.tier).toBe('lost');
  });
});
