package handler

import (
	"net/http"

	_ "viopal-service/docs"
	"viopal-service/internal/middleware"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handlers struct {
	Auth    *Auth
	Balance *Balance
	Invoice *Invoice
	Payment *Payment
	Refund  *Refund
}

func RegisterRoutes(h Handlers) http.Handler {
	r := chi.NewRouter()

	// Swagger
	r.Get("/swagger/*", httpSwagger.WrapHandler.ServeHTTP)

	// API v1
	r.Route("/api/v1", func(v1 chi.Router) {

		// Authentication
		v1.Route("/auth", func(auth chi.Router) {
			auth.Post("/register", h.Auth.Register)
			auth.Post("/login", h.Auth.Login)
		})

		// Protected Routes for Merchant
		v1.Group(func(merchant chi.Router) {
			merchant.Use(middleware.AuthMiddleware)
			merchant.Use(middleware.RoleMiddleware("merchant", "user"))

			// Me
			merchant.Get("/me", h.Auth.Me)

			// Wallet
			merchant.Get("/wallet", h.Balance.MerchantBalance)

			// Balance Requests
			merchant.Route("/balance-requests", func(api chi.Router) {
				api.Post("/", h.Balance.RequestTopUp)
				api.Get("/", h.Balance.ListRequestTopUp)
			})

			// Merchant Invoices
			merchant.Route("/invoices", func(api chi.Router) {
				api.Get("/", h.Invoice.ListInvoice)
				api.Post("/", h.Invoice.RequestInvoices)
				api.Get("/{code}", h.Invoice.InvoiceByCode)
			})

			// Balance Requests
			merchant.Route("/refunds", func(api chi.Router) {
				api.Post("/", h.Refund.RequestRefund)
				api.Get("/", h.Refund.ListRefund)
			})
		})

		// Public Payment
		v1.Route("/public", func(public chi.Router) {
			public.Get("/pay/{token}", h.Payment.Payment)
			public.Post("/pay/{token}/intents", h.Payment.PaymentIntent)
			public.Get("/intents/{code}", h.Payment.PaymentIntentByCode)
		})

		// Protected Routes for Admin
		v1.Route("/admin", func(admin chi.Router) {
			admin.Use(middleware.AuthMiddleware)
			admin.Use(middleware.RoleMiddleware("admin", "user"))

			// Payment Intent
			admin.Route("/payment-intents", func(api chi.Router) {
				api.Get("/", h.Payment.ListPaymentIntent)
				api.Patch("/{payment_intent_number}/status", h.Payment.UpdatePaymentIntent)
			})

			// Balance Request
			admin.Route("/balance-requests", func(api chi.Router) {
				api.Get("/", h.Balance.ListRequestTopUpAdmin)
				api.Patch("/{ref_number}/status", h.Balance.UpdateTopUp)
			})

			// Refund
			admin.Route("/refunds", func(api chi.Router) {
				api.Get("/", h.Refund.ListRefundAdmin)
				api.Patch("/{refund_number}/decision", h.Refund.DecideRefund) // approve or reject
				api.Patch("/{refund_number}/process", h.Refund.ProcessRefund) // success or failed
			})

			// Statistic
			admin.Get("/stats", h.Invoice.InvoiceStatistic)
		})
	})

	return r
}
