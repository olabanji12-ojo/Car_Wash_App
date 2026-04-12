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
	// Root namespaces (support both plural and singular for robustness)
	payments := router.PathPrefix("/api/payments").Subrouter()
	payment := router.PathPrefix("/api/payment").Subrouter()

	payments.Use(middleware.AuthMiddleware)
	payment.Use(middleware.AuthMiddleware)

	// Dashboard Endpoints
	payments.HandleFunc("/summary/{carwash_id}", tr.tc.GetEarningsSummaryHandler).Methods("GET")
	payments.HandleFunc("/history/{carwash_id}", tr.tc.GetTransactionHistoryHandler).Methods("GET")
	
	// Alias support
	payment.HandleFunc("/summary/{carwash_id}", tr.tc.GetEarningsSummaryHandler).Methods("GET")
	payment.HandleFunc("/history/{carwash_id}", tr.tc.GetTransactionHistoryHandler).Methods("GET")

	// Legacy/Other Endpoints
	payments.HandleFunc("/user", tr.tc.GetPaymentsByUserHandler).Methods("GET")
	payments.HandleFunc("/carwash/{id}", tr.tc.GetPaymentsByCarwashHandler).Methods("GET")
}


