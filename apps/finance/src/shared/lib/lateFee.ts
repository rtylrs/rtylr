// Automatic late fee calculation utilities.
import { daysBetween } from './date';

export interface LateFeeInput {
  amount: number;
  dueDate: string | Date;
  today?: Date;
  feePercent: number;
  graceDays: number;
  compoundingMonthly?: boolean;
}

export interface LateFeeResult {
  isLate: boolean;
  daysLate: number;
  fee: number;
  totalOwed: number;
}

export function computeLateFee({
  amount,
  dueDate,
  today = new Date(),
  feePercent,
  graceDays,
  compoundingMonthly = true,
}: LateFeeInput): LateFeeResult {
  const days = daysBetween(dueDate, today);
  const daysLate = days - graceDays;
  if (daysLate <= 0) {
    return { isLate: false, daysLate: 0, fee: 0, totalOwed: amount };
  }
  if (compoundingMonthly) {
    const months = Math.ceil(daysLate / 30);
    const fee = amount * Math.pow(1 + feePercent / 100, months) - amount;
    return { isLate: true, daysLate, fee, totalOwed: amount + fee };
  }
  const fee = amount * (feePercent / 100);
  return { isLate: true, daysLate, fee, totalOwed: amount + fee };
}
