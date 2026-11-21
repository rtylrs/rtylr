import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.expenses.v1;

/** Expense is a recorded business expense. */
model Expense {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  /** category_id is the category id value for this contract. */
  @encode(string)
  category_id?: int64;

  /** amount is the amount value for this contract. */
  amount: float64;

  /** currency is the currency value for this contract. */
  currency: string;

  /** description is the description value for this contract. */
  description: string;

  /** vendor is who was paid, empty when not recorded. */
  vendor: string;

  /** receipt_url links the uploaded receipt, empty when none is attached. */
  receipt_url: string;

  /** payment_method is the canonical payment method for accounting. */
  payment_method: string;

  /** is_tax_deductible indicates whether the expense is tax deductible. */
  is_tax_deductible: boolean;

  /** tax_amount is the tax amount recorded for the expense. */
  tax_amount: float64;

  /** notes contains accounting notes for the expense. */
  notes: string;

  /** expense_date is the expense date value for this contract. */
  expense_date: string;

  /** status is the current lifecycle state. */
  status: string;

  /** created_at is when the record was created. */
  created_at: string;

  /** updated_at is when the record was last updated. */
  updated_at: string;
}

/** ExpenseCategory groups expenses for reporting. */
model ExpenseCategory {
  /** id is the unique identifier for the resource. */
  @encode(string)
  id: int64;

  /** owner_id identifies the tenant that owns the resource. */
  @encode(string)
  owner_id: int64;

  /** name is the name value for this contract. */
  name: string;

  /** color is the color value for this contract. */
  color: string;

  /** created_at is when the record was created. */
  created_at: string;
}

model CreateExpenseRequest {
  /** category_id is the category id value for this contract. */
  @encode(string)
  category_id?: int64;

  /** amount is the amount value for this contract. */
  @minValue(0)
  amount: float64;

  /** currency is the currency value for this contract. */
  @Onkt.required
  currency: string;

  /** description is the description value for this contract. */
  @Onkt.required
  description: string;

  /** vendor is who was paid, empty when not recorded. */
  vendor: string;

  /** receipt_url links the uploaded receipt, empty when none is attached. */
  receipt_url: string;

  /** payment_method is the canonical payment method for accounting. */
  payment_method: string;

  /** is_tax_deductible indicates whether the expense is tax deductible. */
  is_tax_deductible: boolean;

  /** tax_amount is the tax amount recorded for the expense. */
  tax_amount: float64;

  /** notes contains accounting notes for the expense. */
  notes: string;

  /** expense_date is the expense date value for this contract. */
  @Onkt.required
  expense_date: string;
}

model CreateExpenseResponse {
  /** expense_id is the expense id value for this contract. */
  @encode(string)
  expense_id: int64;

  /** message is a human-readable result message. */
  message: string;
}

model DeleteExpenseRequest {
  /** id is the unique identifier for the expense. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteExpenseResponse {
  /** message is a human-readable result message. */
  message: string;
}

model GetExpenseRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model GetExpenseResponse {
  /** expense is the expense value for this contract. */
  expense: Expense;
}

model CreateExpenseCategoryRequest {
  /** name labels the category; names are unique per business, ignoring case. */
  @Onkt.required
  name: string;

  /** color is the swatch shown next to the category. */
  color: string;
}

model CreateExpenseCategoryResponse {
  /** category is the stored category. */
  category: ExpenseCategory;

  /**
   * created is false when a category with the same name already existed and
   * was returned instead of inserting a duplicate.
   */
  created: boolean;
}

model UpdateExpenseCategoryBody {
  /** name labels the category; names are unique per business, ignoring case. */
  @Onkt.required
  name: string;

  /** color is the swatch shown next to the category. */
  color: string;
}

model UpdateExpenseCategoryRequest {
  /** id is the unique identifier for the category. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable category fields bound from the HTTP payload. */
  @body
  body: UpdateExpenseCategoryBody;
}

model UpdateExpenseCategoryResponse {
  /** category is the stored category after the update. */
  category: ExpenseCategory;
}

model DeleteExpenseCategoryRequest {
  /** id is the unique identifier for the category. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;
}

model DeleteExpenseCategoryResponse {
  /** message is a human-readable result message. */
  message: string;
}

model ListExpenseCategoriesRequest {}

model ListExpenseCategoriesResponse {
  /** categories is the categories value for this contract. */
  categories: ExpenseCategory[];
}

model ListExpensesRequest {
  @query("page")
  page: int64;

  @query("page_size")
  page_size: int64;
}

model ListExpensesResponse {
  /** expenses is the expenses value for this contract. */
  expenses: Expense[];

  page: int64;
  page_size: int64;
  has_more: boolean;
}

model UpdateExpenseBody {
  /** category_id is the category id value for this contract. */
  @encode(string)
  category_id?: int64;

  /** amount is the amount value for this contract. */
  @minValue(0)
  amount: float64;

  /** currency is the currency value for this contract. */
  @Onkt.required
  currency: string;

  /** description is the description value for this contract. */
  @Onkt.required
  description: string;

  /** vendor is who was paid, empty when not recorded. */
  vendor: string;

  /** receipt_url links the uploaded receipt, empty when none is attached. */
  receipt_url: string;

  /** payment_method is the canonical payment method for accounting. */
  payment_method: string;

  /** is_tax_deductible indicates whether the expense is tax deductible. */
  is_tax_deductible: boolean;

  /** tax_amount is the tax amount recorded for the expense. */
  tax_amount: float64;

  /** notes contains accounting notes for the expense. */
  notes: string;

  /** expense_date is the expense date value for this contract. */
  @Onkt.required
  expense_date: string;
}

model UpdateExpenseRequest {
  /** id is the unique identifier for the expense. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries updatable expense fields bound from the HTTP payload. */
  @body
  body: UpdateExpenseBody;
}

model UpdateExpenseResponse {
  /** message is a human-readable result message. */
  message: string;
}

model UpdateExpenseStatusBody {
  /** status is the current lifecycle state - "pending", "approved", or "rejected". */
  @Onkt.required
  status: string;
}

model UpdateExpenseStatusRequest {
  /** id is the unique identifier for the resource. */
  @minValueExclusive(0)
  @encode(string)
  id: int64;

  /** body carries status update fields bound from the HTTP payload. */
  @body
  body: UpdateExpenseStatusBody;
}

model UpdateExpenseStatusResponse {
  /** message is a human-readable result message. */
  message: string;
}
