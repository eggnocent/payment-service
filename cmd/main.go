package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"payment-service/clients"
	midtransClient "payment-service/clients/midtrans"
	"payment-service/common/gcs"
	"payment-service/common/response"
	"payment-service/config"
	"payment-service/constants"
	controllers "payment-service/controllers/http"
	kafkaClient "payment-service/controllers/kafka"
	"payment-service/domain/models"
	"payment-service/middlewares"
	"payment-service/repositories"
	"payment-service/routes"
	"payment-service/services"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
	Run: func(c *cobra.Command, args []string) {
		_ = godotenv.Load()
		config.Init()

		// Setup log
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
		logrus.SetLevel(logrus.DebugLevel)

		db, err := config.InitDatabase()
		if err != nil {
			panic(err)
		}

		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			panic(err)
		}
		time.Local = loc

		err = db.AutoMigrate(
			&models.Payment{},
			&models.PaymentHistory{},
		)
		if err != nil {
			panic(err)
		}

		gcs := InitGCS()
		kafka := kafkaClient.NewKafkaRegistry(config.Config.Kafka.Brokers)
		midtrans := midtransClient.NewMidtransClient(config.Config.Midtrans.ServerKey, config.Config.Midtrans.IsProduction)
		client := clients.NewClientRegistry()
		repository := repositories.NewRepositoryRegistry(db)
		service := services.NewServiceRegistry(repository, gcs, kafka, midtrans)
		controller := controllers.NewControllerRegistry(service)

		// ✅ Ganti gin.Default() → gin.New() agar HandlePanic() aktif
		router := gin.New()
		router.Use(middlewares.HandlePanic())
		router.Use(gin.Logger())

		router.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, response.Response{
				Status:  constants.Error,
				Message: fmt.Sprintf("Path %s", http.StatusText(http.StatusNotFound)),
			})
		})

		router.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, response.Response{
				Status:  constants.Success,
				Message: "Welcome to Payment Service",
			})
		})

		// CORS
		router.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name, x-api-key, x-request-at")
			c.Next()
		})

		// Optional rate limiter (komentar masih kamu simpan)
		// lmt := tollbooth.NewLimiter(
		// 	config.Config.RateLimiterMaxRequests,
		// 	&limiter.ExpirableOptions{
		// 		DefaultExpirationTTL: payment.Duration(config.Config.RateLimiterTimeSeconds) * payment.Second,
		// 	})
		// router.Use(middlewares.RateLimiter(lmt))

		group := router.Group("/api/v1")
		route := routes.NewRouteRegistry(controller, group, client)
		route.Serve()

		port := fmt.Sprintf(":%d", config.Config.Port)
		router.Run(port)
	},
}

func Run() {
	err := command.Execute()
	if err != nil {
		panic(err)
	}
}

func InitGCS() gcs.IGCSlient {
	decoded, err := base64.StdEncoding.DecodeString(config.Config.GCSCredentialsEncoded)
	if err != nil {
		panic(err)
	}

	var sa gcs.ServiceAccountKeyJSON
	if err := json.Unmarshal(decoded, &sa); err != nil {
		panic(fmt.Errorf("failed to parse service account JSON: %w", err))
	}

	return gcs.NewGCSClient(sa, config.Config.GCSBucketName)
}
