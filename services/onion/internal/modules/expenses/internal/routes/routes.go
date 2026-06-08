package routes

import (
	"context"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/expenses/internal/application"
	"github.com/voxire/onion/internal/modules/expenses/internal/routes/handlers"
	"github.com/voxire/onion/internal/platform/server/http"
)

func RegisterRoutes(ctx context.Context, router *chi.Mux, svc *application.Service) {
	http.RegisterGET[handlers.ListCategoriesRequest, handlers.ListCategoriesResponse](
		ctx, router, handlers.ExpenseCategoriesRoutes(),
		&handlers.ListCategoriesHandler{Service: svc},
	)

	http.RegisterGET[handlers.ListExpensesRequest, handlers.ListExpensesResponse](
		ctx, router, handlers.ExpenseListRoutes(),
		&handlers.ListExpensesHandler{Service: svc},
	)

	http.RegisterGET[handlers.GetExpenseRequest, handlers.GetExpenseResponse](
		ctx, router, handlers.ExpenseGetRoutes(),
		&handlers.GetExpenseHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.CreateExpenseRequest, handlers.CreateExpenseResponse](
		ctx, router, handlers.ExpenseCreateRoutes(),
		&handlers.CreateExpenseHandler{Service: svc},
	)

	http.RegisterPUTWithValidation[handlers.UpdateExpenseRequest, handlers.UpdateExpenseResponse](
		ctx, router, handlers.ExpenseUpdateRoutes(),
		&handlers.UpdateExpenseHandler{Service: svc},
	)

	http.RegisterDELETEWithValidation[handlers.DeleteExpenseRequest, handlers.DeleteExpenseResponse](
		ctx, router, handlers.ExpenseDeleteRoutes(),
		&handlers.DeleteExpenseHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.ApproveExpenseRequest, handlers.ApproveExpenseResponse](
		ctx, router, handlers.ExpenseApproveRoutes(),
		&handlers.ApproveExpenseHandler{Service: svc},
	)

	http.RegisterPOSTWithValidation[handlers.RejectExpenseRequest, handlers.RejectExpenseResponse](
		ctx, router, handlers.ExpenseRejectRoutes(),
		&handlers.RejectExpenseHandler{Service: svc},
	)
}
