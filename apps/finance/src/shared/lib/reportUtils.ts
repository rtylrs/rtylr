export function grossMarginPct(revenue: number, cogs: number): number {
  if (revenue === 0) return 0;
  return ((revenue - cogs) / revenue) * 100;
}

export function netMarginPct(revenue: number, netProfit: number): number {
  if (revenue === 0) return 0;
  return (netProfit / revenue) * 100;
}

export function operatingExpenseRatio(operatingExpenses: number, revenue: number): number {
  if (revenue === 0) return 0;
  return (operatingExpenses / revenue) * 100;
}

export function returnOnAssets(netProfit: number, totalAssets: number): number {
  if (totalAssets === 0) return 0;
  return (netProfit / totalAssets) * 100;
}

export function debtToEquityRatio(totalLiabilities: number, totalEquity: number): number {
  if (totalEquity === 0) return 0;
  return totalLiabilities / totalEquity;
}

export function currentRatio(currentAssets: number, currentLiabilities: number): number {
  if (currentLiabilities === 0) return 0;
  return currentAssets / currentLiabilities;
}

export function yoyGrowthPct(current: number, prior: number): number {
  if (prior === 0) return 0;
  return ((current - prior) / Math.abs(prior)) * 100;
}

export function ebitda(
  net: number,
  interest: number,
  taxes: number,
  dep: number,
  amort: number
): number {
  return net + interest + taxes + dep + amort;
}
export function quickRatio(cash: number, si: number, ar: number, cl: number): number {
  if (cl === 0) return 0;
  return (cash + si + ar) / cl;
}
export function workingCapital(ca: number, cl: number): number {
  return ca - cl;
}
