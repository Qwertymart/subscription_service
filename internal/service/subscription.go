package service

import (
	"context"
	"log/slog"
	"time"

	"aggregator_db/internal/domain"
	"aggregator_db/internal/repository/postgres"
	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo   postgres.SubscriptionRepository
	logger *slog.Logger
}

func NewSubscriptionService(repo postgres.SubscriptionRepository, logger *slog.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, req domain.CreateSubscriptionRequest) (*domain.Subscription, error) {
	sub := &domain.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		s.logger.ErrorContext(ctx, "failed to create subscription",
			slog.String("user_id", req.UserID.String()),
			slog.String("service", req.ServiceName),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.InfoContext(ctx, "subscription created",
		slog.String("id", sub.ID.String()),
		slog.String("user_id", sub.UserID.String()),
		slog.String("service", sub.ServiceName),
	)

	return sub, nil
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to get subscription",
			slog.String("id", id.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateSubscriptionRequest) (*domain.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ServiceName != nil {
		sub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		sub.Price = *req.Price
	}
	if req.StartDate != nil {
		sub.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		sub.EndDate = req.EndDate
	}

	sub.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, sub); err != nil {
		s.logger.ErrorContext(ctx, "failed to update subscription",
			slog.String("id", id.String()),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.InfoContext(ctx, "subscription updated",
		slog.String("id", sub.ID.String()),
	)

	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete subscription",
			slog.String("id", id.String()),
			slog.String("error", err.Error()),
		)
		return err
	}

	s.logger.InfoContext(ctx, "subscription deleted",
		slog.String("id", id.String()),
	)

	return nil
}

func (s *SubscriptionService) List(ctx context.Context, query domain.ListSubscriptionsQuery) ([]*domain.Subscription, error) {
	subscriptions, err := s.repo.List(ctx, query)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to list subscriptions",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return subscriptions, nil
}

func (s *SubscriptionService) CalculateTotal(ctx context.Context, req domain.CalculateTotalRequest) (*domain.CalculateTotalResponse, error) {
	total, err := s.repo.CalculateTotal(ctx, req)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to calculate total",
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	s.logger.InfoContext(ctx, "total calculated",
		slog.Int("total", total),
	)

	return &domain.CalculateTotalResponse{TotalCost: total}, nil
}
