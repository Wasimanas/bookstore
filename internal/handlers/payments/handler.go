package payments

import (
	"store/internal/services/payments"
	"net/http"
    "log"
	"github.com/gin-gonic/gin"
    "strconv"
)

type PaymentHandler struct {
	service payment.PaymentService
}

func NewPaymentsHandler(s payment.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: s}
}


func (h *PaymentHandler) Pay(c *gin.Context) {
    orderID := c.Param("orderId")
    if orderID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "orderId is required"})
        return
    }

    orderId, e := strconv.Atoi(orderID)
    if e != nil { 
        log.Println("Error Processing OrderID")
    }

    checkoutURL, err := h.service.CreateCheckoutSession(c.Request.Context(), orderId)
    if err != nil {
        log.Printf("Payment error: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "could not initiate payment"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"url": checkoutURL})
}


func (h *PaymentHandler) CreateSetupIntent(c *gin.Context) {
	userIdVar, _ := c.Get("userId")
	userId := userIdVar.(int)

	clientSecret, err := h.service.CreateSetupIntent(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"client_secret": clientSecret})
}

func (h *PaymentHandler) AddPaymentMethod(c *gin.Context) {
	var payload struct {
		PaymentMethodId string `json:"payment_method_id"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIdVar, _ := c.Get("userId")
	userId := userIdVar.(int)

	pm, err := h.service.AddPaymentMethod(c.Request.Context(), userId, payload.PaymentMethodId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pm)
}

func (h *PaymentHandler) ListPayments(c *gin.Context) {
	userIdVar, _ := c.Get("userId")
	userId := userIdVar.(int)

	methods, err := h.service.ListPaymentMethods(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, methods)
}

func (h *PaymentHandler) DeletePayment(c *gin.Context) {
	pmId := c.Param("id")
	userIdVar, _ := c.Get("userId")
	userId := userIdVar.(int)

	err := h.service.DeletePaymentMethod(c.Request.Context(), userId, pmId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *PaymentHandler) SetupPaymentRoutes(rg *gin.RouterGroup) {
	rg.POST("/setup-intent", h.CreateSetupIntent)
	rg.POST("/", h.AddPaymentMethod)
	rg.GET("/", h.ListPayments)
    rg.POST("/pay/:orderId", h.Pay)
	rg.DELETE("/:id", h.DeletePayment)
}
