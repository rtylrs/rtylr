// Benchmark and comparison helpers for the insights feature.

export interface BenchmarkResult {
  value: number;
  benchmark: number;
  delta: number;
  pctDelta: number;
  isAbove: boolean;
}

/**
 * compareToBenchmark computes how a value compares to a benchmark.
 */
export function compareToBenchmark(value: number, benchmark: number): BenchmarkResult {
  const delta = value - benchmark;
  const pctDelta = benchmark === 0 ? 0 : (delta / Math.abs(benchmark)) * 100;
  return {
    value,
    benchmark,
    delta,
    pctDelta: Math.round(pctDelta * 10) / 10,
    isAbove: value > benchmark,
  };
}

/**
 * benchmarkLabel returns a human-readable label for the benchmark comparison.
 */
export function benchmarkLabel(result: BenchmarkResult, unit = ''): string {
  const sign = result.delta >= 0 ? '+' : '';
  const pct = `${sign}${result.pctDelta}%`;
  return unit ? `${pct} vs benchmark (${unit})` : `${pct} vs benchmark`;
}

/**
 * topPerformers returns items whose value exceeds the given threshold.
 */
export function topPerformers<T extends { value: number }>(items: T[], threshold: number): T[] {
  return items.filter((item) => item.value >= threshold);
}

/**
 * underPerformers returns items whose value is below the given threshold.
 */
export function underPerformers<T extends { value: number }>(items: T[], threshold: number): T[] {
  return items.filter((item) => item.value < threshold);
}

/**
 * rankByValue returns items sorted by value descending.
 */
export function rankByValue<T extends { value: number }>(items: T[]): T[] {
  return [...items].sort((a, b) => b.value - a.value);
}
