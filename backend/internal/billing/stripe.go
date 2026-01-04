package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/models"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/invoice"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/usagerecord"
)

// Service handles billing operations
type Service struct {
	stripePricePro        string
	stripePriceEnterprise string
	log                   *logger.Logger
}

// NewService creates a new billing service
func NewService(stripeKey, pricePro, priceEnterprise string, log *logger.Logger) *Service {
	stripe.Key = stripeKey
	return &Service{
		stripePricePro:        pricePro,
		stripePriceEnterprise: priceEnterprise,
		log:                   log,
	}
}

// =============================================================================
// Customer Management
// =============================================================================

// CreateCustomer creates a Stripe customer for a tenant
func (s *Service) CreateCustomer(ctx context.Context, tenant *models.Tenant, email string) (string, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(tenant.Name),
		Metadata: map[string]string{
			"tenant_id": tenant.ID.String(),
		},
	}

	cust, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	s.log.Infow("Stripe customer created", "customer_id", cust.ID, "tenant_id", tenant.ID)
	return cust.ID, nil
}

// GetCustomer retrieves a Stripe customer
func (s *Service) GetCustomer(ctx context.Context, customerID string) (*stripe.Customer, error) {
	cust, err := customer.Get(customerID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Stripe customer: %w", err)
	}
	return cust, nil
}

// UpdateCustomer updates a Stripe customer
func (s *Service) UpdateCustomer(ctx context.Context, customerID string, updates *stripe.CustomerParams) (*stripe.Customer, error) {
	cust, err := customer.Update(customerID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update Stripe customer: %w", err)
	}
	return cust, nil
}

// =============================================================================
// Subscription Management
// =============================================================================

// CreateSubscription creates a subscription for a tenant
func (s *Service) CreateSubscription(ctx context.Context, customerID string, plan models.TenantPlan) (*stripe.Subscription, error) {
	var priceID string
	switch plan {
	case models.PlanPro:
		priceID = s.stripePricePro
	case models.PlanEnterprise:
		priceID = s.stripePriceEnterprise
	default:
		return nil, fmt.Errorf("no subscription needed for plan: %s", plan)
	}

	if priceID == "" {
		return nil, fmt.Errorf("price ID not configured for plan: %s", plan)
	}

	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
	}

	sub, err := subscription.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	s.log.Infow("subscription created", "subscription_id", sub.ID, "customer_id", customerID, "plan", plan)
	return sub, nil
}

// CancelSubscription cancels a subscription
func (s *Service) CancelSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	}

	sub, err := subscription.Update(subscriptionID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	s.log.Infow("subscription cancelled", "subscription_id", subscriptionID)
	return sub, nil
}

// GetSubscription retrieves a subscription
func (s *Service) GetSubscription(ctx context.Context, subscriptionID string) (*stripe.Subscription, error) {
	sub, err := subscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return sub, nil
}

// =============================================================================
// Usage Metering
// =============================================================================

// UsageType represents a type of usage to meter
type UsageType string

const (
	UsageTypeTokens     UsageType = "tokens"
	UsageTypeExecutions UsageType = "executions"
	UsageTypeStorage    UsageType = "storage"
)

// RecordUsage records usage for a subscription item
func (s *Service) RecordUsage(ctx context.Context, subscriptionItemID string, quantity int64, usageType UsageType, timestamp time.Time) error {
	if subscriptionItemID == "" {
		return nil // No subscription to meter
	}

	params := &stripe.UsageRecordParams{
		SubscriptionItem: stripe.String(subscriptionItemID),
		Quantity:         stripe.Int64(quantity),
		Timestamp:        stripe.Int64(timestamp.Unix()),
		Action:           stripe.String(string(stripe.UsageRecordActionIncrement)),
	}

	_, err := usagerecord.New(params)
	if err != nil {
		return fmt.Errorf("failed to record usage: %w", err)
	}

	s.log.Debugw("usage recorded", 
		"subscription_item_id", subscriptionItemID, 
		"quantity", quantity, 
		"type", usageType,
	)
	return nil
}

// =============================================================================
// Invoices
// =============================================================================

// ListInvoices lists invoices for a customer
func (s *Service) ListInvoices(ctx context.Context, customerID string, limit int) ([]*stripe.Invoice, error) {
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(customerID),
	}
	params.Limit = stripe.Int64(int64(limit))

	var invoices []*stripe.Invoice
	iter := invoice.List(params)
	for iter.Next() {
		invoices = append(invoices, iter.Invoice())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}

	return invoices, nil
}

// =============================================================================
// Webhook Handling
// =============================================================================

// WebhookEvent represents a Stripe webhook event
type WebhookEvent struct {
	Type string
	Data map[string]interface{}
}

// HandleWebhook processes a Stripe webhook
func (s *Service) HandleWebhook(ctx context.Context, event stripe.Event) error {
	s.log.Infow("processing Stripe webhook", "event_type", event.Type)

	switch event.Type {
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(ctx, event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	case "invoice.paid":
		return s.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(ctx, event)
	default:
		s.log.Debugw("ignoring unhandled Stripe event", "event_type", event.Type)
	}

	return nil
}

func (s *Service) handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	s.log.Infow("subscription created", "event_id", event.ID)
	// Update tenant plan in database
	return nil
}

func (s *Service) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	s.log.Infow("subscription updated", "event_id", event.ID)
	// Update tenant plan in database
	return nil
}

func (s *Service) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	s.log.Infow("subscription deleted", "event_id", event.ID)
	// Downgrade tenant to free plan
	return nil
}

func (s *Service) handleInvoicePaid(ctx context.Context, event stripe.Event) error {
	s.log.Infow("invoice paid", "event_id", event.ID)
	// Record payment in audit log
	return nil
}

func (s *Service) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	s.log.Warnw("invoice payment failed", "event_id", event.ID)
	// Notify tenant and potentially restrict access
	return nil
}

// =============================================================================
// Pricing Plans
// =============================================================================

// PricingPlan represents a pricing plan
type PricingPlan struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Currency    string
	Interval    string // monthly, yearly
	Features    []string
	Limits      PlanLimits
}

// PlanLimits represents usage limits for a plan
type PlanLimits struct {
	MaxAgents        int
	MaxExecutionsDay int
	MaxTokensMonth   int64
	MaxStorageGB     int
	MaxRepositories  int
	MaxKnowledgeBases int
}

// GetPricingPlans returns available pricing plans
func (s *Service) GetPricingPlans() []PricingPlan {
	return []PricingPlan{
		{
			ID:          "free",
			Name:        "Free",
			Description: "For individuals and small projects",
			Price:       0,
			Currency:    "USD",
			Interval:    "monthly",
			Features: []string{
				"3 agents",
				"100 executions/day",
				"50K tokens/month",
				"1 repository",
				"1 knowledge base",
				"Community support",
			},
			Limits: PlanLimits{
				MaxAgents:         3,
				MaxExecutionsDay:  100,
				MaxTokensMonth:    50000,
				MaxStorageGB:      1,
				MaxRepositories:   1,
				MaxKnowledgeBases: 1,
			},
		},
		{
			ID:          "pro",
			Name:        "Pro",
			Description: "For professionals and growing teams",
			Price:       49,
			Currency:    "USD",
			Interval:    "monthly",
			Features: []string{
				"Unlimited agents",
				"1000 executions/day",
				"1M tokens/month",
				"10 repositories",
				"10 knowledge bases",
				"Priority support",
				"Advanced analytics",
			},
			Limits: PlanLimits{
				MaxAgents:         -1, // Unlimited
				MaxExecutionsDay:  1000,
				MaxTokensMonth:    1000000,
				MaxStorageGB:      50,
				MaxRepositories:   10,
				MaxKnowledgeBases: 10,
			},
		},
		{
			ID:          "enterprise",
			Name:        "Enterprise",
			Description: "For large organizations with custom needs",
			Price:       0, // Custom pricing
			Currency:    "USD",
			Interval:    "monthly",
			Features: []string{
				"Everything in Pro",
				"Unlimited executions",
				"Unlimited tokens",
				"Unlimited repositories",
				"Unlimited knowledge bases",
				"Dedicated support",
				"Custom integrations",
				"SLA guarantee",
				"On-premise option",
			},
			Limits: PlanLimits{
				MaxAgents:         -1,
				MaxExecutionsDay:  -1,
				MaxTokensMonth:    -1,
				MaxStorageGB:      -1,
				MaxRepositories:   -1,
				MaxKnowledgeBases: -1,
			},
		},
	}
}

// CheckLimits checks if a tenant is within their plan limits
func (s *Service) CheckLimits(ctx context.Context, tenantID uuid.UUID, plan models.TenantPlan, usage PlanUsage) error {
	plans := s.GetPricingPlans()
	
	var limits PlanLimits
	for _, p := range plans {
		if p.ID == string(plan) {
			limits = p.Limits
			break
		}
	}

	// Check each limit
	if limits.MaxAgents > 0 && usage.Agents > limits.MaxAgents {
		return fmt.Errorf("agent limit exceeded: %d/%d", usage.Agents, limits.MaxAgents)
	}
	if limits.MaxExecutionsDay > 0 && usage.ExecutionsToday > limits.MaxExecutionsDay {
		return fmt.Errorf("daily execution limit exceeded: %d/%d", usage.ExecutionsToday, limits.MaxExecutionsDay)
	}
	if limits.MaxTokensMonth > 0 && usage.TokensThisMonth > limits.MaxTokensMonth {
		return fmt.Errorf("monthly token limit exceeded: %d/%d", usage.TokensThisMonth, limits.MaxTokensMonth)
	}
	if limits.MaxRepositories > 0 && usage.Repositories > limits.MaxRepositories {
		return fmt.Errorf("repository limit exceeded: %d/%d", usage.Repositories, limits.MaxRepositories)
	}
	if limits.MaxKnowledgeBases > 0 && usage.KnowledgeBases > limits.MaxKnowledgeBases {
		return fmt.Errorf("knowledge base limit exceeded: %d/%d", usage.KnowledgeBases, limits.MaxKnowledgeBases)
	}

	return nil
}

// PlanUsage represents current usage for a tenant
type PlanUsage struct {
	Agents          int
	ExecutionsToday int
	TokensThisMonth int64
	StorageUsedGB   float64
	Repositories    int
	KnowledgeBases  int
}

