/**
 * Nutritional information helpers for the Menu app.
 */

export interface NutritionInfo {
  calories: number;
  protein?: number;
  carbs?: number;
  fat?: number;
  fiber?: number;
}

/** Return the calorie tier label for a menu item. */
export function calorieTier(calories: number): 'low' | 'medium' | 'high' | 'very-high' {
  if (calories < 300) return 'low';
  if (calories < 600) return 'medium';
  if (calories < 900) return 'high';
  return 'very-high';
}

/** Return the percentage of daily recommended calories (2000 kcal reference). */
export function dailyCaloriePct(calories: number, dailyTarget = 2000): number {
  return Math.round((calories / dailyTarget) * 100);
}

/** Return the total calories from macronutrients (protein: 4, carbs: 4, fat: 9 kcal/g). */
export function caloriesFromMacros(protein: number, carbs: number, fat: number): number {
  return Math.round(protein * 4 + carbs * 4 + fat * 9);
}

/** Return true if the nutrition info is within the given calorie limit. */
export function isWithinCalorieLimit(info: NutritionInfo, limit: number): boolean {
  return info.calories <= limit;
}

/** Sum nutrition values across multiple items. */
export function sumNutrition(items: NutritionInfo[]): NutritionInfo {
  return items.reduce(
    (acc, item) => ({
      calories: acc.calories + item.calories,
      protein: (acc.protein ?? 0) + (item.protein ?? 0),
      carbs: (acc.carbs ?? 0) + (item.carbs ?? 0),
      fat: (acc.fat ?? 0) + (item.fat ?? 0),
      fiber: (acc.fiber ?? 0) + (item.fiber ?? 0),
    }),
    { calories: 0, protein: 0, carbs: 0, fat: 0, fiber: 0 }
  );
}
