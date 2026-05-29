import { describe, it, expect } from 'vitest';
import {
  maskEnd,
  maskStart,
  maskMiddle,
  maskEmail,
  maskPhone,
  maskCard,
  maskApiKey,
  isMasked,
} from './maskUtils';

describe('maskEnd', () => {
  it('masks all but last 4 chars', () => expect(maskEnd('abcdefgh', 4)).toBe('****efgh'));
  it('returns value unchanged when short', () => expect(maskEnd('abc', 5)).toBe('abc'));
});

describe('maskStart', () => {
  it('masks after first 3 chars', () => expect(maskStart('abcdefgh', 3)).toBe('abc*****'));
});

describe('maskMiddle', () => {
  it('keeps first and last visible', () => expect(maskMiddle('abcdefghij', 2)).toBe('ab******ij'));
  it('returns value unchanged when too short', () => expect(maskMiddle('abcd', 3)).toBe('abcd'));
});

describe('maskEmail', () => {
  it('masks local part', () =>
    expect(maskEmail('john.doe@example.com')).toBe('j******e@example.com'));
  it('handles short local part', () => expect(maskEmail('ab@example.com')).toBe('a*@example.com'));
  it('returns unchanged if no at sign', () => expect(maskEmail('notanemail')).toBe('notanemail'));
});

describe('maskPhone', () => {
  it('shows only last 4 digits', () => expect(maskPhone('555-867-5309')).toBe('***-***-5309'));
});

describe('maskCard', () => {
  it('shows only last 4 digits', () =>
    expect(maskCard('4111111111111111')).toBe('**** **** **** 1111'));
});

describe('maskApiKey', () => {
  it('shows first and last 4', () =>
    expect(maskApiKey('sk-abcdefghijklmnop')).toBe('sk-a***********mnop'));
  it('fully masks short keys', () => expect(maskApiKey('short')).toBe('*****'));
});

describe('isMasked', () => {
  it('returns true for masked', () => expect(isMasked('***-***-1234')).toBe(true));
  it('returns false for unmasked', () => expect(isMasked('hello')).toBe(false));
});
