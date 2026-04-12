package routes

import (
	"github.com/olabanji12-ojo/CarWashApp/controllers"
	"github.com/olabanji12-ojo/CarWashApp/middleware"

	"github.com/gorilla/mux"
)

type WalletRouter struct {
	walletController *controllers.WalletController
}

func NewWalletRouter(wc *controllers.WalletController) *WalletRouter {
	return &WalletRouter{walletController: wc}
}

func (wr *WalletRouter) RegisterWalletRoutes(router *mux.Router) {
	// Protected Subrouter
	wallet := router.PathPrefix("/api/wallet").Subrouter()
	wallet.Use(middleware.AuthMiddleware)

	wallet.HandleFunc("", wr.walletController.GetWallet).Methods("GET")
	wallet.HandleFunc("/fund", wr.walletController.FundWallet).Methods("POST")
	wallet.HandleFunc("/verify", wr.walletController.VerifyDeposit).Methods("POST")
	wallet.HandleFunc("/escrow", wr.walletController.EscrowHold).Methods("POST")
	wallet.HandleFunc("/escrow/release", wr.walletController.EscrowRelease).Methods("POST")
	wallet.HandleFunc("/withdraw", wr.walletController.RequestWithdrawal).Methods("POST")
}
