package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	authController "github.com/CPU-commits/Template_Go-EventDriven/src/auth/controller"
	"github.com/CPU-commits/Template_Go-EventDriven/src/cmd/bus/queue"
	"github.com/CPU-commits/Template_Go-EventDriven/src/cmd/http/docs"
	"github.com/CPU-commits/Template_Go-EventDriven/src/package/logger"
	"github.com/CPU-commits/Template_Go-EventDriven/src/settings"
	"github.com/CPU-commits/Template_Go-EventDriven/src/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/secure"
	"github.com/markbates/goth/gothic"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"

	// swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

var settingsData = settings.GetSettings()

func Init(zapLogger *zap.Logger, logger logger.Logger) {
	router := gin.New()
	// Proxies
	router.SetTrustedProxies([]string{"localhost"})
	// Logger
	router.Use(ginzap.GinzapWithConfig(zapLogger, &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		SkipPathRegexps: []*regexp.Regexp{
			regexp.MustCompile(`(?i)(swagger|healthz)`),
		},
	}))
	router.Use(ginzap.RecoveryWithZap(zapLogger, true))

	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Server Internal Error: %s", err))
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"title": "Server internal error",
		})
	}))
	// Docs
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Version = "v1"
	docs.SwaggerInfo.Host = "localhost:8080"
	// CORS
	if settingsData.GO_ENV == "prod" {
		router.Use(cors.New(cors.Config{
			AllowOrigins:     strings.Split(settingsData.CORS_DOMAINS, ","),
			AllowMethods:     []string{"GET", "OPTIONS", "PUT", "DELETE", "POST"},
			AllowCredentials: true,
			AllowHeaders:     []string{"*"},
			AllowWebSockets:  false,
			MaxAge:           12 * time.Hour,
		}))
	} else {
		router.Use(cors.Default())
	}
	// Secure
	sslUrl := "ssl." + settingsData.CLIENT_DOMAIN
	secureConfig := secure.Config{
		SSLHost:              sslUrl,
		STSSeconds:           315360000,
		STSIncludeSubdomains: true,
		FrameDeny:            true,
		ContentTypeNosniff:   true,
		BrowserXssFilter:     true,
		IENoOpen:             true,
		ReferrerPolicy:       "strict-origin-when-cross-origin",
		SSLProxyHeaders: map[string]string{
			"X-Fowarded-Proto": "https",
		},
	}
	router.Use(secure.New(secureConfig))
	// I18n
	router.Use(func(ctx *gin.Context) {
		lang := ctx.DefaultQuery("lang", "es")
		ctx.Set("localizer", utils.GetLocalizer(lang))
	})
	// Init bus
	bus := queue.New(logger)
	// Route docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// Auth
	initAuthProviders()
	authProviders := router.Group("auth")
	auth := router.Group("api/auth")
	{
		// Controllers
		authControlle := authController.NewAuthHttpController(bus)
		// Define routes
		auth.POST("/login", authControlle.Login)
		auth.POST("/refresh", authControlle.Refresh)
		// auth.POST("/register", authControlle.Register)
		authProviders.GET("/:provider/callback", func(ctx *gin.Context) {
			provider := ctx.Param("provider")
			ctx.Request = ctx.Request.WithContext(
				context.WithValue(
					context.Background(),
					"provider",
					provider,
				),
			)

			user, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
			if err != nil {
				redirectToSavePage(ctx)
				return
			}

			err = authControlle.HandleAuthUser(ctx, user)
			if err != nil {
				redirectToSavePage(ctx)
				return
			}

			redirectToClient(ctx)
		})
		authProviders.GET("/:provider", func(ctx *gin.Context) {
			provider := ctx.Param("provider")
			ctx.Request = ctx.Request.WithContext(
				context.WithValue(
					context.Background(),
					"provider",
					provider,
				),
			)

			if gothUser, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request); err != nil && gothUser.AccessToken != "" {
				redirectToClient(ctx)
			} else {
				gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
			}
		})
	}
	// Route healthz
	router.GET("/healthz", func(ctx *gin.Context) {
		ctx.String(200, "OK")
	})
	// No route
	router.NoRoute(func(ctx *gin.Context) {
		ctx.String(404, "Not found")
	})
	// Init server
	if err := router.Run(); err != nil {
		log.Fatalf("Error init server")
	}
}
