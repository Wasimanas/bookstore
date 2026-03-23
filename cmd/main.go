package main


import (
    "store/internal/config"
    "store/internal/db"
    "store/internal/mq"
    AuthRepo "store/internal/repositories/auth"
    AuthService "store/internal/services/auth"
    AuthHandler "store/internal/handlers/auth"

    ProductRepo "store/internal/repositories/product"
    ProductService "store/internal/services/product"
    ProductHandler "store/internal/handlers/product"


    OrdersRepo "store/internal/repositories/orders"
    OrdersService "store/internal/services/orders"
    OrdersHandler "store/internal/handlers/orders"

    PaymentsRepo "store/internal/repositories/payments"
    PaymentsService "store/internal/services/payments"
    PaymentsHandler "store/internal/handlers/payments"

    "log"

    "github.com/gin-gonic/gin"
	"go.uber.org/zap"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)


func main() { 

     // Logger 
    logger, _ := zap.NewProduction()
    defer logger.Sync() 
    sugar := logger.Sugar()

    // Load Configuration
    config, err := config.InitConfig() 
    if err != nil { 
        sugar.Errorw("Error Loading Configuration")
        panic("Unable to Load Config")
    }

    // Initialize DB & Migrations
    DBConn, err := db.InitDB(config)
    defer DBConn.Close()
    if err != nil { 
        sugar.Errorw("Error Loading Database")
        panic("Unable to Load Config")

    }
    // err = db.RunDBMigrations(DBConn, sugar)
    if err != nil { 
        sugar.Errorw("Error Running DB Migrations")
        panic("DB Migration Error")

    }

    // Init MessageQueue
    MQConn, err := mq.InitMQ(config)
    if err != nil  { 
        log.Println("Error Connecting to Message Queue")
    }
    // Creating Consumer & Publisher Channels
    PubChan, err := MQConn.Channel()
    ConChan, err := MQConn.Channel()


     

    // repositories
    authRepo := AuthRepo.NewAuthRepo(DBConn)
    productRepo := ProductRepo.NewProductRepo(DBConn)
    ordersRepo := OrdersRepo.NewOrdersRepo(DBConn)
    paymentsRepo := PaymentsRepo.NewPaymentsRepo(DBConn)


    // services
    authService := AuthService.NewAuthService(authRepo, config)
    productService := ProductService.NewProductsService(productRepo)
    ordersService := OrdersService.NewOrdersService(ordersRepo, authRepo, paymentsRepo, config)
    paymentsService := PaymentsService.NewPaymentService(paymentsRepo, authRepo, ordersRepo, config)

    // handlers
    authHandler := AuthHandler.NewAuthHandler(authService)
    productHandler := ProductHandler.NewProductHandler(productService)
    ordersHandler := OrdersHandler.NewOrdersHandler(ordersService, config)
    paymentsHandler := PaymentsHandler.NewPaymentsHandler(paymentsService)

    // gin router
    Router := gin.Default()

    // Route Registrations
	v1 := Router.Group("/v1/api")
    
    // AuthRouteRegistrations
	authGroup := v1.Group("/auth")
	authHandler.SetupAuthRoutes(authGroup)

    // ProductRoutes
	productGroup := v1.Group("/product")
	productHandler.SetupProductRoutes(productGroup)


    //OrdersRoutes
	ordersGroup := v1.Group("/orders")
	ordersHandler.SetupOrderRoutes(ordersGroup)

    //PaymentsRoutes
	paymentsGroup := v1.Group("/payments")
	paymentsHandler.SetupPaymentRoutes(paymentsGroup)



    RouterErr := Router.Run(":8081")
    if RouterErr != nil { 
        sugar.Errorw("Error Starting Server")
        panic("Gin Router Error, Fix it.")
    }

}
