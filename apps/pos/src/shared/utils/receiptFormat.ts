// Pure helpers for formatting receipt lines and data.

/** Truncate a string to a maximum length, appending an ellipsis if cut. */
export function truncate(text: string, maxLen: number): string {
  if (text.length <= maxLen) return text;
  return `${text.slice(0, maxLen - 1)}.`;
}

/** Pad a string to a fixed width, left-aligning by default. */
export function padEnd(text: string, width: number): string {
  return text.padEnd(width);
}

/** Pad a string to a fixed width, right-aligning (for prices). */
export function padStart(text: string, width: number): string {
  return text.padStart(width);
}

/** Build a two-column receipt line: item name and price separated by spaces. */
export function receiptLine(label: string, price: string, lineWidth = 40): string {
  const maxLabel = lineWidth - price.length - 1;
  const safeLabel = label.length > maxLabel ? truncate(label, maxLabel) : label;
  return `${safeLabel.padEnd(maxLabel)} ${price.padStart(price.length)}`;
}

/** Format a quantity-times-price expression, e.g. "2 x $5.00". */
export function formatQtyPrice(qty: number, unitPrice: string): string {
  return `${qty} x ${unitPrice}`;
}

/** Generate a divider line of repeated characters. */
export function divider(char = '-', width = 40): string {
  return char.repeat(width);
}

/** Center a string within a given width, padding with spaces. */
export function centerText(text: string, width = 40): string {
  if (text.length >= width) return text;
  const totalPad = width - text.length;
  const leftPad = Math.floor(totalPad / 2);
  return text.padStart(text.length + leftPad).padEnd(width);
}
