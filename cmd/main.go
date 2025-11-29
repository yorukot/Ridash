package main

import (
	"net/http"
	swaggerDocs "ridash/docs"
	customMiddleware "ridash/middleware"

	scalar "github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"

	"ridash/db"
	"ridash/router"
	"ridash/utils/config"
	"ridash/utils/id"
	"ridash/utils/logger"
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

// @host localhost:8000
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load env
	// Initialize logger
	logger.InitLogger()
	defer zap.L().Sync()

	env, err := config.InitConfig()
	if err != nil {
		zap.L().Fatal("Failed to load config:", zap.Error(err))
	}

	e := echo.New()
	e.Use(customMiddleware.ZapLogger(zap.L()))
	e.Use(middleware.Recover())

	// Initialize sonyflake
	err = id.Init()
	if err != nil {
		zap.L().Fatal("Failed to initialize ID generator:", zap.Error(err))
	}

	db, err := db.InitDatabase()
	if err != nil {
		zap.L().Fatal("Failed to initialize database:", zap.Error(err))
	}

	// Setup routes
	routes(e, db)
	zap.L().Fatal("Api server crash", zap.Error(e.Start(":"+env.AppPort)))
}

func routes(e *echo.Echo, db *pgxpool.Pool) {
	if config.Env().AppEnv == config.AppEnvDev {
		// Swagger documentation route
		e.GET("/swagger/*", echoSwagger.WrapHandler)
		e.GET("/reference", scalarDocsHandler())
	}

	// User routes
	api := e.Group("/api")
	router.AuthRouter(api, db)
	router.TeamRouter(api, db)
	router.FolderRouter(api, db)
	router.DocumentRouter(api, db)
}

func scalarDocsHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		html, err := scalar.ApiReferenceHTML(&scalar.Options{
			// Use generated swagger spec as inline content to avoid filesystem lookups
			SpecContent: swaggerDocs.SwaggerInfo.ReadDoc(),
			CustomOptions: scalar.CustomOptions{
				PageTitle: "Ridash API Reference",
			},
			DarkMode: true,
		})
		if err != nil {
			zap.L().Error("failed to generate Scalar docs", zap.Error(err))
			return c.String(http.StatusInternalServerError, "could not render API reference")
		}

		return c.HTML(http.StatusOK, html)
	}
}
