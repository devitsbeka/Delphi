package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/delphi-platform/delphi/backend/internal/billing"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/stripe/stripe-go/v76/webhook"
)

// BillingHandler handles billing-related HTTP requests
type BillingHandler struct {
	billingService *billing.Service
	webhookSecret  string
	log            *logger.Logger
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(bs *billing.Service, webhookSecret string, log *logger.Logger) *BillingHandler {
	return &BillingHandler{
		billingService: bs,
		webhookSecret:  webhookSecret,
		log:            log,
	}
}

// GetPricingPlans returns available pricing plans
func (h *BillingHandler) GetPricingPlans(w http.ResponseWriter, r *http.Request) {
	plans := h.billingService.GetPricingPlans()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plans": plans,
	})
}

// GetCurrentPlan returns the current plan for a tenant
func (h *BillingHandler) GetCurrentPlan(w http.ResponseWriter, r *http.Request) {
	// Extract tenant from context (set by auth middleware)
	tenantID := r.Context().Value("tenant_id")
	if tenantID == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// TODO: Fetch tenant and return plan info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plan":  "pro",
		"usage": map[string]interface{}{
			"executions":   847,
			"tokens":       654000,
			"storage_gb":   23.5,
		},
		"limits": map[string]interface{}{
			"executions": 1000,
			"tokens":     1000000,
			"storage_gb": 50,
		},
	})
}

// CreateSubscription creates a new subscription
func (h *BillingHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Plan string `json:"plan"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Create subscription via Stripe
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"subscription_id": "sub_xxx",
		"plan":            req.Plan,
		"status":          "active",
	})
}

// CancelSubscription cancels a subscription
func (h *BillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	// TODO: Cancel subscription via Stripe
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "cancelled",
		"message": "Subscription will remain active until end of billing period",
	})
}

// GetInvoices returns invoices for a tenant
func (h *BillingHandler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	// TODO: Fetch invoices from Stripe
	invoices := []map[string]interface{}{
		{"date": "2026-01-01", "amount": 49.00, "status": "paid"},
		{"date": "2025-12-01", "amount": 49.00, "status": "paid"},
		{"date": "2025-11-01", "amount": 49.00, "status": "paid"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invoices": invoices,
	})
}

// HandleStripeWebhook handles Stripe webhook events
func (h *BillingHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Errorw("failed to read webhook body", "error", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), h.webhookSecret)
	if err != nil {
		h.log.Errorw("failed to verify webhook signature", "error", err)
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	if err := h.billingService.HandleWebhook(r.Context(), event); err != nil {
		h.log.Errorw("failed to handle webhook", "error", err, "event_type", event.Type)
		http.Error(w, "webhook handler failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetUsageMetrics returns usage metrics for a tenant
func (h *BillingHandler) GetUsageMetrics(w http.ResponseWriter, r *http.Request) {
	// TODO: Fetch actual usage metrics
	metrics := map[string]interface{}{
		"period": map[string]interface{}{
			"start": "2026-01-01T00:00:00Z",
			"end":   "2026-01-31T23:59:59Z",
		},
		"executions": map[string]interface{}{
			"total":    847,
			"by_agent": map[string]int{
				"code-review": 342,
				"bug-fixer":   287,
				"content":     218,
			},
		},
		"tokens": map[string]interface{}{
			"input":  456000,
			"output": 198000,
			"total":  654000,
		},
		"cost": map[string]interface{}{
			"openai":    245.80,
			"anthropic": 189.50,
			"google":    67.20,
			"total":     502.50,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// SetBudget sets spending budget for a tenant or business
func (h *BillingHandler) SetBudget(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Budget      float64 `json:"budget"`
		AlertAt     float64 `json:"alert_at"`
		HardLimit   bool    `json:"hard_limit"`
		BusinessID  string  `json:"business_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Save budget settings
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"budget":     req.Budget,
		"alert_at":   req.AlertAt,
		"hard_limit": req.HardLimit,
		"updated":    true,
	})
}

