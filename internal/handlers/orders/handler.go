package orders

import (
	"store/internal/models"
	"store/internal/services/orders"
	"store/internal/util"
    "store/internal/config"
	"io"
	"net/http"
	"strconv"
    
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	service orders.OrderService
    config *config.ApplicationConfig
}

func NewOrdersHandler(service orders.OrderService, config *config.ApplicationConfig) *OrderHandler {
    return &OrderHandler{service: service, config: config}
}



func (h *OrderHandler) CreateCheckoutSession(c *gin.Context) {
	orderId, _ := strconv.Atoi(c.Param("orderId"))
	var payload struct {
		SuccessUrl string `json:"success_url"`
		CancelUrl  string `json:"cancel_url"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authContext, err := util.GetUserFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId, _ := strconv.Atoi(authContext.UserId)

	url, err := h.service.CreateCheckoutSession(c.Request.Context(), orderId, userId, payload.SuccessUrl, payload.CancelUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}

func (h *OrderHandler) Webhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	err = h.service.HandleWebhook(c.Request.Context(), payload, sigHeader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *OrderHandler) RefundOrder(c *gin.Context) {
	var payload struct {
		OrderID int          `json:"order_id"`
		Items   []models.RefundItem `json:"items"`
		Reason  string       `json:"reason"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.RefundOrder(
		c.Request.Context(),
		payload.OrderID,
		payload.Items,
		payload.Reason,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var payload models.CreateOrderPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), payload.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "order": order})
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	authContext, err := util.GetUserFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	orders, err := h.service.GetUserOrders(c.Request.Context(), authContext.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "results": orders})
}

func (h *OrderHandler) CreateOrderItem(c *gin.Context) {
	var payload models.CreateOrderItemPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authContext, err := util.GetUserFromContext(c, h.config)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	orderItem, err := h.service.CreateOrderItem(c.Request.Context(), payload, authContext.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true, "result": orderItem})
}

func (h *OrderHandler) SetupOrderRoutes(rg *gin.RouterGroup) {
	rg.GET("/", h.GetUserOrders)
	rg.POST("", h.CreateOrder)
	rg.POST("/item", h.CreateOrderItem)
	rg.POST("/:orderId/checkout", h.CreateCheckoutSession)
}
func (h *OrderHandler) SetupWebhookRoutes(rg *gin.RouterGroup) {
	rg.POST("/stripe/webhook", h.Webhook)
}

func (h *OrderHandler) SetupInternalRoutes(rg *gin.RouterGroup) {
	rg.POST("/orders/refund", h.RefundOrder)
}

