import { describe, it, expect } from 'vitest';
import {
  calorieTier,
  dailyCaloriePct,
  caloriesFromMacros,
  isWithinCalorieLimit,
  sumNutrition,
} from './nutritionUtils';

describe('calorieTier', () => {
  it('returns low for calories below 300', () => {
    expect(calorieTier(0)).toBe('low');
    expect(calorieTier(299)).toBe('low');
  });

  it('returns medium for 300-599 calories', () => {
    expect(calorieTier(300)).toBe('medium');
    expect(calorieTier(599)).toBe('medium');
  });

  it('returns high for 600-899 calories', () => {
    expect(calorieTier(600)).toBe('high');
    expect(calorieTier(899)).toBe('high');
  });

  it('returns very-high for 900+ calories', () => {
    expect(calorieTier(900)).toBe('very-high');
    expect(calorieTier(1500)).toBe('very-high');
  });
});

describe('dailyCaloriePct', () => {
  it('calculates percentage of daily calories', () => {
    expect(dailyCaloriePct(500)).toBe(25);
    expect(dailyCaloriePct(2000)).toBe(100);
  });

  it('rounds to nearest integer', () => {
    expect(dailyCaloriePct(333)).toBe(17);
  });

  it('accepts a custom daily target', () => {
    expect(dailyCaloriePct(500, 1000)).toBe(50);
  });
});

describe('caloriesFromMacros', () => {
  it('calculates calories using 4-4-9 formula', () => {
    // 10g protein * 4 + 20g carbs * 4 + 5g fat * 9 = 40 + 80 + 45 = 165
    expect(caloriesFromMacros(10, 20, 5)).toBe(165);
  });

  it('returns 0 for all zero macros', () => {
    expect(caloriesFromMacros(0, 0, 0)).toBe(0);
  });

  it('rounds the result', () => {
    // 1g protein * 4 + 1g carbs * 4 + 1g fat * 9 = 17
    expect(caloriesFromMacros(1, 1, 1)).toBe(17);
  });
});

describe('isWithinCalorieLimit', () => {
  it('returns true when calories are within limit', () => {
    expect(isWithinCalorieLimit({ calories: 400 }, 500)).toBe(true);
  });

  it('returns true when calories equal the limit', () => {
    expect(isWithinCalorieLimit({ calories: 500 }, 500)).toBe(true);
  });

  it('returns false when calories exceed the limit', () => {
    expect(isWithinCalorieLimit({ calories: 600 }, 500)).toBe(false);
  });
});

describe('sumNutrition', () => {
  it('sums calories across items', () => {
    const items = [{ calories: 200 }, { calories: 300 }];
    expect(sumNutrition(items).calories).toBe(500);
  });

  it('sums optional macro fields', () => {
    const items = [
      { calories: 200, protein: 10, carbs: 20, fat: 5, fiber: 3 },
      { calories: 300, protein: 15, carbs: 30, fat: 8, fiber: 2 },
    ];
    const result = sumNutrition(items);
    expect(result.protein).toBe(25);
    expect(result.carbs).toBe(50);
    expect(result.fat).toBe(13);
    expect(result.fiber).toBe(5);
  });

  it('treats missing optional fields as zero', () => {
    const items = [{ calories: 100, protein: 10 }, { calories: 200 }];
    const result = sumNutrition(items);
    expect(result.protein).toBe(10);
    expect(result.carbs).toBe(0);
  });

  it('returns zero totals for empty array', () => {
    const result = sumNutrition([]);
    expect(result.calories).toBe(0);
    expect(result.protein).toBe(0);
  });
});
