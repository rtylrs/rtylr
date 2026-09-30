// Single source of truth for the POS operator PIN policy, mirrored from the
// server: services/onion/internal/modules/pos_users/internal/domain/validate.go
// sets MinPINLength=6, MaxPINLength=12 and ASCII digits on Create, Update and
// SetPIN. A change to that policy must land here too or the client will offer
// PINs the server refuses (onboarding fails) or refuse PINs the server accepts.

/** Shortest PIN that may be set or rotated (server MinPINLength). */
export const PIN_MIN_LENGTH = 6;
/** Longest PIN that may be set or rotated (server MaxPINLength). */
export const PIN_MAX_LENGTH = 12;
/**
 * Length of the PINs issued before the policy above existed. The server
 * deliberately does not enforce the new minimum when a PIN is *verified*, so
 * these keep working; the entry screen must therefore still accept them, or a
 * cashier on shift cannot clock in.
 */
export const LEGACY_PIN_LENGTH = 4;

const DIGITS_ONLY = /^\d+$/;

/** True for a PIN the server will accept when one is set or rotated. */
export function isNewPin(pin: string): boolean {
  return (
    DIGITS_ONLY.test(pin) && pin.length >= PIN_MIN_LENGTH && pin.length <= PIN_MAX_LENGTH
  );
}

/**
 * True for a PIN worth sending to verify: a legacy 4-digit one, or any PIN of
 * the current 6-12 digit policy. Anything shorter was never valid and anything
 * longer cannot have been set, so neither is worth a round trip.
 */
export function isVerifiablePin(pin: string): boolean {
  return (
    DIGITS_ONLY.test(pin) && pin.length >= LEGACY_PIN_LENGTH && pin.length <= PIN_MAX_LENGTH
  );
}
