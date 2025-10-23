package http

import (
	"errors"
	"net/http"

	"aggregator_db/internal/domain"
	"aggregator_db/internal/repository/postgres"
	"aggregator_db/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	service *service.SubscriptionService
}

func NewSubscriptionHandler(service *service.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

// CreateSubscription godoc
// @Summary      Создать новую подписку
// @Description  Создает новую запись о подписке пользователя
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription body domain.CreateSubscriptionRequest true "Данные подписки"
// @Success      201 {object} domain.Subscription
// @Failure      400 {object} domain.ErrorResponse
// @Failure      500 {object} domain.ErrorResponse
// @Router       /subscriptions [post]
func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req domain.CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	subscription, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// GetSubscription godoc
// @Summary      Получить подписку по ID
// @Description  Возвращает информацию о подписке по её идентификатору
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id path string true "ID подписки" Format(uuid)
// @Success      200 {object} domain.Subscription
// @Failure      400 {object} domain.ErrorResponse
// @Failure      404 {object} domain.ErrorResponse
// @Router       /subscriptions/{id} [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "invalid subscription id"})
		return
	}

	subscription, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: "subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// UpdateSubscription godoc
// @Summary      Обновить подписку
// @Description  Обновляет существующую подписку
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id path string true "ID подписки" Format(uuid)
// @Param        subscription body domain.UpdateSubscriptionRequest true "Обновляемые данные"
// @Success      200 {object} domain.Subscription
// @Failure      400 {object} domain.ErrorResponse
// @Failure      404 {object} domain.ErrorResponse
// @Failure      500 {object} domain.ErrorResponse
// @Router       /subscriptions/{id} [put]
func (h *SubscriptionHandler) UpdateSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "invalid subscription id"})
		return
	}

	var req domain.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	subscription, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: "subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscription)
}

// DeleteSubscription godoc
// @Summary      Удалить подписку
// @Description  Удаляет подписку по её идентификатору
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id path string true "ID подписки" Format(uuid)
// @Success      200 {object} domain.SuccessResponse
// @Failure      400 {object} domain.ErrorResponse
// @Failure      404 {object} domain.ErrorResponse
// @Router       /subscriptions/{id} [delete]
func (h *SubscriptionHandler) DeleteSubscription(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "invalid subscription id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: "subscription not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "subscription deleted"})
}

// ListSubscriptions godoc
// @Summary      Получить список подписок
// @Description  Возвращает список подписок с возможностью фильтрации
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        user_id query string false "ID пользователя" Format(uuid)
// @Param        service_name query string false "Название сервиса"
// @Param        limit query int false "Лимит записей" default(100)
// @Param        offset query int false "Смещение" default(0)
// @Success      200 {array} domain.Subscription
// @Failure      400 {object} domain.ErrorResponse
// @Router       /subscriptions [get]
func (h *SubscriptionHandler) ListSubscriptions(c *gin.Context) {
	var query domain.ListSubscriptionsQuery

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	subscriptions, err := h.service.List(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// CalculateTotal godoc
// @Summary      Рассчитать суммарную стоимость
// @Description  Рассчитывает суммарную стоимость подписок за период с фильтрацией
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        user_id query string false "ID пользователя" Format(uuid)
// @Param        service_name query string false "Название сервиса"
// @Param        start_period query string true "Начало периода" Format(MM-YYYY)
// @Param        end_period query string true "Конец периода" Format(MM-YYYY)
// @Success      200 {object} domain.CalculateTotalResponse
// @Failure      400 {object} domain.ErrorResponse
// @Failure      500 {object} domain.ErrorResponse
// @Router       /subscriptions/calculate [get]
func (h *SubscriptionHandler) CalculateTotal(c *gin.Context) {
	var req domain.CalculateTotalRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.service.CalculateTotal(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
