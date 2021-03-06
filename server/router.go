package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/websentry/websentry/config"
	"github.com/websentry/websentry/controllers"
	"github.com/websentry/websentry/middlewares"
)

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/v1/worker/fetch_task"},
	}))
	r.Use(gin.Recovery())

	// CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = config.GetConfig().CROSAllowOrigins
	corsConfig.AddAllowHeaders("WS-User-Token")
	corsConfig.AddAllowHeaders("WS-Worker-Key")
	r.Use(cors.New(corsConfig))

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	v1 := r.Group("/v1")
	{

		// sensitive api
		sensitive := v1.Group("")
		sensitive.Use(middlewares.GetSensitiveLimiter())
		{
			sensitive.POST("/get_verification", controllers.UserGetSignUpVerification)
			sensitive.POST("/create_user", controllers.UserCreateWithVerification)
			sensitive.POST("/login", controllers.UserLogin)
		}

		// general api
		general := v1.Group("")
		general.Use(middlewares.UserAuthRequired)
		general.Use(middlewares.GetGeneralLimiter())
		{
			// user
			userGroup := general.Group("/user")
			{
				userGroup.POST("/info", controllers.UserInfo)
				userGroup.POST("/update", controllers.UserUpdateSettings)
			}

			// sentry
			sentryGroup := general.Group("/sentry")
			{
				sentryGroup.POST("/wait_full_screenshot", controllers.SentryWaitFullScreenshot)
				sentryGroup.POST("/create", controllers.SentryCreate)
				sentryGroup.POST("/list", controllers.SentryList)
				sentryGroup.POST("/info", controllers.SentryInfo)
				sentryGroup.POST("/remove", controllers.SentryRemove)

				screenshot := sentryGroup.Group("")
				screenshot.Use(middlewares.GetScreenshotLimiter())
				{
					screenshot.POST("/request_full_screenshot", controllers.SentryRequestFullScreenshot)
				}
			}

			// notification
			notificationGroup := general.Group("/notification")
			{
				notificationGroup.POST("/list", controllers.NotificationList)
				notificationGroup.POST("/add_serverchan", controllers.NotificationAddServerChan)
			}

		}

		// worker
		workerGroup := v1.Group("/worker")
		workerGroup.Use(middlewares.WorkerAuth)
		workerGroup.Use(middlewares.GetWorkerLimiter())
		{
			workerGroup.POST("/init", controllers.WorkerInit)
			workerGroup.POST("/fetch_task", controllers.WorkerFetchTask)
			workerGroup.POST("/submit_task", controllers.WorkerSubmitTask)
		}

		// common
		commonGroup := v1.Group("/common")
		{
			commonGroup.GET("/get_history_image", controllers.GetHistoryImage)
			commonGroup.GET("/get_full_screenshot_image", controllers.GetFullScreenshotImage)
		}

	}

	return r
}
