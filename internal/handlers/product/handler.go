package product

import (
    "net/http"
    "strconv"

    "store/internal/models"
    "store/internal/services/product"

    "github.com/gin-gonic/gin"
)

type ProductHandler struct {
    service product.ProductService
}

func NewProductHandler(service product.ProductService) *ProductHandler {
    return &ProductHandler{service: service}
}


func (h *ProductHandler) CreateProduct(c *gin.Context) {
    var req models.Product
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    created, err := h.service.CreateProduct(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"ok": true, "book": created})
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
    idStr := c.Param("bookId")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book ID"})
        return
    }

    if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"ok": true, "deleted_book_id": id})
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
    idStr := c.Param("bookId")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book ID"})
        return
    }

    var req models.Product
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    updated, err := h.service.UpdateProduct(c.Request.Context(), id, req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"ok": true, "book": updated})
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
    idStr := c.Param("bookId")
    id, err := strconv.ParseInt(idStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid book ID"})
        return
    }

    book, err := h.service.GetProductByID(c.Request.Context(), id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"ok": true, "book": book})
}

func (h *ProductHandler) SearchProducts(c *gin.Context) {
    title := c.Query("title")

    results, err := h.service.SearchProducts(c.Request.Context(), title)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"ok": true, "results": results})
}
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
    // Default values
    limit := 10
    offset := 0

    // Parse query params if provided
    if l := c.Query("limit"); l != "" {
        var err error
        limit, err = strconv.Atoi(l)
        if err != nil || limit < 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
            return
        }
    }

    if o := c.Query("offset"); o != "" {
        var err error
        offset, err = strconv.Atoi(o)
        if err != nil || offset < 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
            return
        }
    }

    // Call service layer
    rv, err := h.service.GetAllProducts(c.Request.Context(), limit, offset)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "ok":      true,
        "results": rv,
        "limit":   limit,
        "offset":  offset,
    })
}


func (h *ProductHandler) SetupProductRoutes(rg *gin.RouterGroup) {
    rg.POST("", h.CreateProduct)
    rg.DELETE("/:bookId", h.DeleteProduct)
    rg.PATCH("/:bookId", h.UpdateProduct)
    rg.GET("/:bookId", h.GetProduct)
    rg.GET("/", h.GetAllProducts)
    rg.GET("/search", h.SearchProducts)
}


