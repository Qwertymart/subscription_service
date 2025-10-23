package postgres

import (
	"context"
	"errors"
	"fmt"

	"aggregator_db/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound      = errors.New("subscription not found")
	ErrAlreadyExists = errors.New("subscription already exists")
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	Update(ctx context.Context, sub *domain.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, query domain.ListSubscriptionsQuery) ([]*domain.Subscription, error)
	CalculateTotal(ctx context.Context, req domain.CalculateTotalRequest) (int, error)
}

type subscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	query := `
        INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	_, err := r.db.Exec(ctx, query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.UserID,
		sub.StartDate,
		sub.EndDate,
		sub.CreatedAt,
		sub.UpdatedAt,
	)

	return err
}

func (r *subscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE id = $1
    `

	var sub domain.Subscription
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}

	return &sub, err
}

func (r *subscriptionRepo) Update(ctx context.Context, sub *domain.Subscription) error {
	query := `
        UPDATE subscriptions
        SET service_name = $2, price = $3, start_date = $4, end_date = $5, updated_at = $6
        WHERE id = $1
    `

	result, err := r.db.Exec(ctx, query,
		sub.ID,
		sub.ServiceName,
		sub.Price,
		sub.StartDate,
		sub.EndDate,
		sub.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *subscriptionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *subscriptionRepo) List(ctx context.Context, query domain.ListSubscriptionsQuery) ([]*domain.Subscription, error) {
	sqlQuery := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE 1=1
    `
	args := []interface{}{}
	argIndex := 1

	if query.UserID != nil {
		userUUID, err := uuid.Parse(*query.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user_id format: %w", err)
		}
		sqlQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userUUID)
		argIndex++
	}

	if query.ServiceName != nil {
		sqlQuery += fmt.Sprintf(" AND service_name = $%d", argIndex)
		args = append(args, *query.ServiceName)
		argIndex++
	}

	sqlQuery += " ORDER BY created_at DESC"

	if query.Limit > 0 {
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, query.Limit)
		argIndex++
	} else {
		sqlQuery += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, 100)
		argIndex++
	}

	if query.Offset > 0 {
		sqlQuery += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, query.Offset)
	}

	rows, err := r.db.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	subscriptions := make([]*domain.Subscription, 0)
	for rows.Next() {
		var sub domain.Subscription
		err := rows.Scan(
			&sub.ID,
			&sub.ServiceName,
			&sub.Price,
			&sub.UserID,
			&sub.StartDate,
			&sub.EndDate,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, &sub)
	}

	return subscriptions, rows.Err()
}

func (r *subscriptionRepo) CalculateTotal(ctx context.Context, req domain.CalculateTotalRequest) (int, error) {
	sqlQuery := `
        WITH period_calculations AS (
            SELECT 
                price,
                GREATEST(
                    TO_DATE(start_date, 'MM-YYYY'),
                    TO_DATE($1, 'MM-YYYY')
                ) as calc_start,
                LEAST(
                    COALESCE(TO_DATE(end_date, 'MM-YYYY'), TO_DATE($2, 'MM-YYYY')),
                    TO_DATE($2, 'MM-YYYY')
                ) as calc_end
            FROM subscriptions
            WHERE 
                TO_DATE(start_date, 'MM-YYYY') <= TO_DATE($2, 'MM-YYYY')
                AND (end_date IS NULL OR TO_DATE(end_date, 'MM-YYYY') >= TO_DATE($1, 'MM-YYYY'))
    `

	args := []interface{}{req.StartPeriod, req.EndPeriod}
	argIndex := 3

	if req.UserID != nil {
		userUUID, err := uuid.Parse(*req.UserID)
		if err != nil {
			return 0, fmt.Errorf("invalid user_id format: %w", err)
		}
		sqlQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, userUUID)
		argIndex++
	}

	if req.ServiceName != nil {
		sqlQuery += fmt.Sprintf(" AND service_name = $%d", argIndex)
		args = append(args, *req.ServiceName)
		argIndex++
	}

	sqlQuery += `
        )
        SELECT COALESCE(SUM(
            price * (
                (EXTRACT(YEAR FROM calc_end)::int - EXTRACT(YEAR FROM calc_start)::int) * 12 +
                (EXTRACT(MONTH FROM calc_end)::int - EXTRACT(MONTH FROM calc_start)::int) + 1
            )
        ), 0)::int as total
        FROM period_calculations
        WHERE calc_end >= calc_start
    `

	var total int
	err := r.db.QueryRow(ctx, sqlQuery, args...).Scan(&total)
	return total, err
}
