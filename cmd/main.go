package main

import (
	"log"
	"ridash/db"
	"ridash/router"
	"ridash/utils/config"
	"ridash/utils/id"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "ridash/docs" // Import generated docs
)

// @title Ridash API
// @version 1.0
// @description This is the Ridash API server for user authentication and management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load env
	env, err := config.InitConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Initialize sonyflake
	err = id.Init()
	if err != nil {
		e.Logger.Fatal("Failed to initialize ID generator:", err)
	}

	db, err := db.InitDatabase(e.Logger)
	if err != nil {
		e.Logger.Fatal("Failed to connect to database:", err)
	}

	// Setup routes
	routes(e, db)
	e.Logger.Infof("Starting server on port %s in %s mode", env.AppPort, env.AppEnv)
	e.Logger.Fatal(e.Start(":" + env.AppPort))
}

func routes(e *echo.Echo, db *pgxpool.Pool) {
	if config.Env().AppEnv == config.AppEnvDev {
		// Swagger documentation route
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}

	// User routes
	api := e.Group("/api")
	router.AuthRouter(api, db)
}
