/**
 * Unit conversion and formatting helpers for the ERP inventory app.
 * All functions are pure with no side-effects.
 */

export type WeightUnit = 'g' | 'kg' | 'lb' | 'oz';
export type VolumeUnit = 'ml' | 'l' | 'fl_oz' | 'gal';
export type LengthUnit = 'mm' | 'cm' | 'm' | 'in' | 'ft';

/** Weight conversion factors relative to grams. */
const WEIGHT_TO_GRAMS: Record<WeightUnit, number> = {
  g: 1,
  kg: 1000,
  lb: 453.592,
  oz: 28.3495,
};

/** Volume conversion factors relative to milliliters. */
const VOLUME_TO_ML: Record<VolumeUnit, number> = {
  ml: 1,
  l: 1000,
  fl_oz: 29.5735,
  gal: 3785.41,
};

/** Length conversion factors relative to millimeters. */
const LENGTH_TO_MM: Record<LengthUnit, number> = {
  mm: 1,
  cm: 10,
  m: 1000,
  in: 25.4,
  ft: 304.8,
};

/** Convert a weight value from one unit to another. */
export function convertWeight(value: number, from: WeightUnit, to: WeightUnit): number {
  const grams = value * WEIGHT_TO_GRAMS[from];
  return grams / WEIGHT_TO_GRAMS[to];
}

/** Convert a volume value from one unit to another. */
export function convertVolume(value: number, from: VolumeUnit, to: VolumeUnit): number {
  const ml = value * VOLUME_TO_ML[from];
  return ml / VOLUME_TO_ML[to];
}

/** Convert a length value from one unit to another. */
export function convertLength(value: number, from: LengthUnit, to: LengthUnit): number {
  const mm = value * LENGTH_TO_MM[from];
  return mm / LENGTH_TO_MM[to];
}

/** Round a converted value to a given number of decimal places. */
export function roundTo(value: number, decimals: number): number {
  const factor = Math.pow(10, decimals);
  return Math.round(value * factor) / factor;
}

/** Format a weight quantity with its unit label. */
export function formatWeight(value: number, unit: WeightUnit, decimals = 2): string {
  return `${roundTo(value, decimals)} ${unit}`;
}

/** Format a volume quantity with its unit label. */
export function formatVolume(value: number, unit: VolumeUnit, decimals = 2): string {
  return `${roundTo(value, decimals)} ${unit.replace('_', ' ')}`;
}

/** Return human-readable label for a unit. */
export function unitLabel(unit: WeightUnit | VolumeUnit | LengthUnit): string {
  const labels: Record<string, string> = {
    g: 'Grams',
    kg: 'Kilograms',
    lb: 'Pounds',
    oz: 'Ounces',
    ml: 'Milliliters',
    l: 'Liters',
    fl_oz: 'Fluid Ounces',
    gal: 'Gallons',
    mm: 'Millimeters',
    cm: 'Centimeters',
    m: 'Meters',
    in: 'Inches',
    ft: 'Feet',
  };
  return labels[unit] ?? unit;
}

/** Determine if a number of pieces should be displayed with a unit. */
export function needsUnit(quantity: number, unit?: string): boolean {
  return unit !== undefined && unit !== 'pcs' && quantity !== 1;
}
