// @title VioPal
// @version 1.0
// @description API documentation for Project VioPal

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"log"
	"net/http"

	"viopal-service/internal/config"
	"viopal-service/internal/handler"
	"viopal-service/internal/repository"
	"viopal-service/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db := config.NewMySQLConnection()

	// repositories
	authRepo := repository.NewAuthRepository(db)
	balanceRepo := repository.NewBalanceRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	refundRepo := repository.NewRefundRepository(db)

	// services
	authService := service.NewAuthService(authRepo)
	balanceService := service.NewBalanceService(authRepo, balanceRepo)
	invoiceService := service.NewInvoiceService(authRepo, invoiceRepo)
	paymentService := service.NewPaymentService(authRepo, paymentRepo, invoiceRepo, balanceRepo)
	refundService := service.NewRefundService(authRepo, refundRepo, paymentRepo, balanceRepo)

	// handlers
	authHandler := handler.NewAuth(authService)
	balanceHandler := handler.NewBalance(balanceService)
	invoiceHandler := handler.NewInvoice(invoiceService)
	paymentHandler := handler.NewPayment(paymentService)
	refundHandler := handler.NewRefund(refundService)

	// routes
	router := handler.RegisterRoutes(handler.Handlers{
		Auth:    authHandler,
		Balance: balanceHandler,
		Invoice: invoiceHandler,
		Payment: paymentHandler,
		Refund:  refundHandler,
	})

	log.Fatal(http.ListenAndServe(":8080", router))
}
