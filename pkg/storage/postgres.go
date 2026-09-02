package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/corlin/AIMeter/pkg/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClient struct {
	pool *pgxpool.Pool
}

func NewPostgresClient(dsn string) (*PostgresClient, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres dsn: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 2
	config.MaxConnLifetime = 1 * time.Hour

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres pool: %w", err)
	}

	client := &PostgresClient{pool: pool}
	// Seed default tenants if needed
	_ = client.SeedDefaultTenants(context.Background())
	return client, nil
}

func (p *PostgresClient) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *PostgresClient) Close() {
	p.pool.Close()
}

func (p *PostgresClient) SeedDefaultTenants(ctx context.Context) error {
	defaultTenants := []domain.Tenant{
		{ID: "org-enterprise-1", Name: "Enterprise Corp", DefaultCurrency: "USD", GlobalDiscount: 0.15},
		{ID: "org-fintech-2", Name: "Fintech Global", DefaultCurrency: "USD", GlobalDiscount: 0.0},
		{ID: "default", Name: "Default Organization", DefaultCurrency: "USD", GlobalDiscount: 0.0},
	}
	for _, t := range defaultTenants {
		_ = p.UpsertTenant(ctx, t)
	}
	return nil
}

// SeedRates inserts rates if not existing
func (p *PostgresClient) SeedRates(ctx context.Context, rates []domain.RateEntry) error {
	query := `
		INSERT INTO rate_catalogs (
			id, provider, model, meter_name, region, service_tier, 
			pricing_type, unit_price, currency, unit, effective_start_at, effective_end_at,
			tenant_id, discount_rate, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW()
		)
		ON CONFLICT (id) DO UPDATE SET
			unit_price = EXCLUDED.unit_price,
			discount_rate = EXCLUDED.discount_rate,
			updated_at = NOW()
	`

	for _, r := range rates {
		id := r.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		if r.EffectiveStartAt.IsZero() {
			r.EffectiveStartAt = time.Now().UTC()
		}
		if r.Currency == "" {
			r.Currency = "USD"
		}
		if r.Unit == "" {
			r.Unit = "Count"
		}
		if r.PricingType == "" {
			r.PricingType = "flat"
		}
		if r.Region == "" {
			r.Region = "global"
		}
		if r.ServiceTier == "" {
			r.ServiceTier = "default"
		}

		_, err := p.pool.Exec(ctx, query,
			id, r.Provider, r.Model, r.MeterName, r.Region, r.ServiceTier,
			r.PricingType, r.UnitPrice, r.Currency, r.Unit, r.EffectiveStartAt, r.EffectiveEndAt,
			r.TenantID, r.DiscountRate,
		)
		if err != nil {
			return fmt.Errorf("failed to seed rate (%s-%s-%s): %w", r.Provider, r.Model, r.MeterName, err)
		}
	}

	return nil
}

// GetRates loads all rates from database
func (p *PostgresClient) GetRates(ctx context.Context) ([]domain.RateEntry, error) {
	query := `
		SELECT 
			id, provider, model, meter_name, region, service_tier, 
			pricing_type, unit_price, currency, unit, effective_start_at, effective_end_at,
			tenant_id, discount_rate
		FROM rate_catalogs
		ORDER BY provider, model, meter_name
	`

	rates := make([]domain.RateEntry, 0)
	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return rates, fmt.Errorf("failed to query rates: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r domain.RateEntry
		if err := rows.Scan(
			&r.ID, &r.Provider, &r.Model, &r.MeterName, &r.Region, &r.ServiceTier,
			&r.PricingType, &r.UnitPrice, &r.Currency, &r.Unit, &r.EffectiveStartAt, &r.EffectiveEndAt,
			&r.TenantID, &r.DiscountRate,
		); err == nil {
			rates = append(rates, r)
		}
	}

	return rates, nil
}

// UpsertTenant creates or updates a tenant
func (p *PostgresClient) UpsertTenant(ctx context.Context, tenant domain.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, default_currency, global_discount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			default_currency = EXCLUDED.default_currency,
			global_discount = EXCLUDED.global_discount,
			updated_at = NOW()
	`

	_, err := p.pool.Exec(ctx, query, tenant.ID, tenant.Name, tenant.DefaultCurrency, tenant.GlobalDiscount)
	return err
}

// GetTenants returns all registered tenants
func (p *PostgresClient) GetTenants(ctx context.Context) ([]domain.Tenant, error) {
	query := `SELECT id, name, default_currency, global_discount, created_at, updated_at FROM tenants ORDER BY name`
	tenants := make([]domain.Tenant, 0)
	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return tenants, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t domain.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.DefaultCurrency, &t.GlobalDiscount, &t.CreatedAt, &t.UpdatedAt); err == nil {
			tenants = append(tenants, t)
		}
	}
	return tenants, nil
}
