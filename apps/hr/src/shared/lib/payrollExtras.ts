export type FilingStatus = 'single' | 'married' | 'head_of_household';
export type DeductionType = 'health' | 'dental' | 'vision' | 'retirement' | 'other';

export function estimateFederalTax(annualIncome: number, filingStatus: FilingStatus): number {
  const brackets: Record<FilingStatus, Array<[number, number]>> = {
    single: [
      [11600, 0.1],
      [44725, 0.12],
      [95375, 0.22],
      [201050, 0.24],
      [383900, 0.32],
      [487450, 0.35],
      [Infinity, 0.37],
    ],
    married: [
      [23200, 0.1],
      [89450, 0.12],
      [190750, 0.22],
      [364200, 0.24],
      [462500, 0.32],
      [693750, 0.35],
      [Infinity, 0.37],
    ],
    head_of_household: [
      [16550, 0.1],
      [63100, 0.12],
      [100500, 0.22],
      [191950, 0.24],
      [243700, 0.32],
      [487450, 0.35],
      [Infinity, 0.37],
    ],
  };

  let tax = 0;
  let prev = 0;
  for (const [limit, rate] of brackets[filingStatus]) {
    if (annualIncome <= prev) break;
    const taxable = Math.min(annualIncome, limit) - prev;
    tax += taxable * rate;
    prev = limit;
  }
  return roundCents(tax);
}

export function roundCents(amount: number): number {
  return Math.round(amount * 100) / 100;
}

export function ficaEmployeeContribution(
  grossPay: number,
  ytdWages: number,
  ssWageBase = 160200
): number {
  const ssContribution = Math.max(0, Math.min(grossPay, ssWageBase - ytdWages)) * 0.062;
  const medicareContribution = grossPay * 0.0145;
  const additionalMedicare = ytdWages + grossPay > 200000 ? grossPay * 0.009 : 0;
  return roundCents(ssContribution + medicareContribution + additionalMedicare);
}

export function calculatePayStubNet(
  grossPay: number,
  federalTax: number,
  stateTax: number,
  ficaTax: number,
  benefitDeductions: number
): number {
  return roundCents(grossPay - federalTax - stateTax - ficaTax - benefitDeductions);
}

export function deductionTypeLabel(type: DeductionType): string {
  const labels: Record<DeductionType, string> = {
    health: 'Health Insurance',
    dental: 'Dental Insurance',
    vision: 'Vision Insurance',
    retirement: '401(k) Contribution',
    other: 'Other Deduction',
  };
  return labels[type];
}

export function retirementContribution(grossPay: number, contributionRate: number): number {
  const rate = Math.min(1, Math.max(0, contributionRate));
  return roundCents(grossPay * rate);
}
