import "@typespec/http";
import "@onkt/typespec";

using TypeSpec.Http;

namespace finance.expenses.v1;

/** ExpensesService defines expense tracking endpoints. */
@route("/finance/expenses/v1")
interface ExpensesService {
  @get
  @route("/expenses")
  listExpenses(...ListExpensesRequest): ListExpensesResponse;

  @post
  @route("/expenses")
  createExpense(...CreateExpenseRequest): CreateExpenseResponse;

  @put
  @route("/expenses/{id}")
  updateExpense(...UpdateExpenseRequest): UpdateExpenseResponse;

  @delete
  @route("/expenses/{id}")
  deleteExpense(...DeleteExpenseRequest): DeleteExpenseResponse;

  @get
  @route("/expenses/{id}")
  getExpense(...GetExpenseRequest): GetExpenseResponse;

  @patch
  @route("/expenses/{id}/status")
  updateExpenseStatus(
    ...UpdateExpenseStatusRequest,
  ): UpdateExpenseStatusResponse;

  @get
  @route("/categories")
  listExpenseCategories(
    ...ListExpenseCategoriesRequest,
  ): ListExpenseCategoriesResponse;

  @post
  @route("/categories")
  createExpenseCategory(
    ...CreateExpenseCategoryRequest,
  ): CreateExpenseCategoryResponse;

  @put
  @route("/categories/{id}")
  updateExpenseCategory(
    ...UpdateExpenseCategoryRequest,
  ): UpdateExpenseCategoryResponse;

  @delete
  @route("/categories/{id}")
  deleteExpenseCategory(
    ...DeleteExpenseCategoryRequest,
  ): DeleteExpenseCategoryResponse;
}
