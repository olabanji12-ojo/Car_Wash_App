package routes

import (
	"github.com/gorilla/mux"
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"
)

type TransactionRouter struct {
	tc controllers.TransactionController
}

func NewTransactionRouter(tc controllers.TransactionController) *TransactionRouter {
	return &TransactionRouter{tc: tc}
}

func (tr *TransactionRouter) TransactionRoutes(router *mux.Router) {
	payment := router.PathPrefix("/api/payments").Subrouter()
	payment.Use(middleware.AuthMiddleware)

	// Dashboard Endpoints
	payment.HandleFunc("/summary/{carwash_id}", tr.tc.GetEarningsSummaryHandler).Methods("GET")
	payment.HandleFunc("/history/{carwash_id}", tr.tc.GetTransactionHistoryHandler).Methods("GET")

	// Legacy/Other Endpoints
	payment.HandleFunc("/user", tr.tc.GetPaymentsByUserHandler).Methods("GET")
	payment.HandleFunc("/carwash/{id}", tr.tc.GetPaymentsByCarwashHandler).Methods("GET")
}


