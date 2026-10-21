/**
 * The client mirror of the server PIN policy
 * (services/onion/.../pos_users/internal/domain/validate.go: MinPINLength=6,
 * MaxPINLength=12, ASCII digits, enforced on set/rotate but not on verify).
 * If these drift, the client either offers PINs the server refuses or refuses
 * PINs staff are already using.
 */
import { describe, it, expect } from 'vitest';
import { PIN_MIN_LENGTH, PIN_MAX_LENGTH, LEGACY_PIN_LENGTH, isNewPin, isVerifiablePin } from './pinPolicy';

describe('PIN policy constants', () => {
  it('matches the server bounds', () => {
    expect(PIN_MIN_LENGTH).toBe(6);
    expect(PIN_MAX_LENGTH).toBe(12);
    expect(LEGACY_PIN_LENGTH).toBe(4);
  });
});

describe('isNewPin', () => {
  it('accepts a PIN of the new minimum length', () => {
    expect(isNewPin('123456')).toBe(true);
  });

  it('accepts every length up to the server maximum', () => {
    for (let length = PIN_MIN_LENGTH; length <= PIN_MAX_LENGTH; length += 1) {
      expect(isNewPin('1'.repeat(length))).toBe(true);
    }
  });

  it('rejects a legacy 4-digit PIN, which the server no longer sets', () => {
    expect(isNewPin('1234')).toBe(false);
    expect(isNewPin('12345')).toBe(false);
  });

  it('rejects an over-long or empty PIN', () => {
    expect(isNewPin('1'.repeat(PIN_MAX_LENGTH + 1))).toBe(false);
    expect(isNewPin('')).toBe(false);
  });

  it('rejects non-digits, including Arabic-Indic digits awaiting normalisation', () => {
    expect(isNewPin('12a456')).toBe(false);
    expect(isNewPin('123456 ')).toBe(false);
    expect(isNewPin('٤٥٦٧٨٩')).toBe(false);
  });
});

describe('isVerifiablePin', () => {
  it('still accepts a legacy 4-digit PIN so an existing operator can clock in', () => {
    expect(isVerifiablePin('1234')).toBe(true);
  });

  it('accepts every PIN length the server can hold', () => {
    for (let length = LEGACY_PIN_LENGTH; length <= PIN_MAX_LENGTH; length += 1) {
      expect(isVerifiablePin('9'.repeat(length))).toBe(true);
    }
  });

  it('rejects a PIN no one could have been issued', () => {
    expect(isVerifiablePin('123')).toBe(false);
    expect(isVerifiablePin('1'.repeat(PIN_MAX_LENGTH + 1))).toBe(false);
    expect(isVerifiablePin('12a4')).toBe(false);
    expect(isVerifiablePin('')).toBe(false);
  });
});
