package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/vibast-solutions/ms-go-profile/app/controller"
	profilegrpc "github.com/vibast-solutions/ms-go-profile/app/grpc"
	"github.com/vibast-solutions/ms-go-profile/app/repository"
	"github.com/vibast-solutions/ms-go-profile/app/service"
	"github.com/vibast-solutions/ms-go-profile/app/types"
	"github.com/vibast-solutions/ms-go-profile/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP and gRPC servers",
	Long:  "Start both HTTP (Echo) and gRPC servers for the profile service.",
	Run:   runServe,
}

// init registers the serve command.
func init() {
	rootCmd.AddCommand(serveCmd)
}

// runServe wires dependencies and starts HTTP and gRPC servers.
func runServe(_ *cobra.Command, _ []string) {
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}
	if err := configureLogging(cfg); err != nil {
		logrus.WithError(err).Fatal("Failed to configure logging")
	}

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.MySQLMaxOpen)
	db.SetMaxIdleConns(cfg.MySQLMaxIdle)
	db.SetConnMaxLifetime(cfg.MySQLMaxLife)

	if err := db.Ping(); err != nil {
		logrus.WithError(err).Fatal("Failed to ping database")
	}

	profileRepo := repository.NewProfileRepository(db)
	profileService := service.NewProfileService(profileRepo)
	profileController := controller.NewProfileController(profileService)

	e := setupHTTPServer(profileController)
	grpcServer, lis := setupGRPCServer(cfg, profileService)

	go func() {
		httpAddr := net.JoinHostPort(cfg.HTTPHost, cfg.HTTPPort)
		logrus.WithField("addr", httpAddr).Info("Starting HTTP server")
		if err := e.Start(httpAddr); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("HTTP server error")
		}
	}()

	go func() {
		logrus.WithField("addr", lis.Addr().String()).Info("Starting gRPC server")
		if err := grpcServer.Serve(lis); err != nil {
			logrus.WithError(err).Fatal("gRPC server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		logrus.WithError(err).Warn("HTTP shutdown error")
	}
	grpcServer.GracefulStop()

	logrus.Info("Server stopped")
}

// setupHTTPServer configures the Echo HTTP server and routes.
func setupHTTPServer(ctrl *controller.ProfileController) *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	e.Use(echomiddleware.RequestLoggerWithConfig(echomiddleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogMethod:    true,
		LogRemoteIP:  true,
		LogLatency:   true,
		LogUserAgent: true,
		LogError:     true,
		HandleError:  true,
		LogRequestID: true,
		LogValuesFunc: func(c echo.Context, v echomiddleware.RequestLoggerValues) error {
			fields := logrus.Fields{
				"remote_ip":  v.RemoteIP,
				"host":       v.Host,
				"method":     v.Method,
				"uri":        v.URI,
				"status":     v.Status,
				"latency":    v.Latency.String(),
				"latency_ns": v.Latency.Nanoseconds(),
				"user_agent": v.UserAgent,
			}
			entry := logrus.WithFields(fields)
			if v.Error != nil {
				entry = entry.WithError(v.Error)
			}
			entry.Info("http_request")
			return nil
		},
	}))
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())
	//request ID middleware used with custom generator to differentiate the request ids coming from HTTP vs request ids
	//created by us
	e.Use(echomiddleware.RequestIDWithConfig(echomiddleware.RequestIDConfig{
		Generator: func() string {
			return fmt.Sprintf("rest-%s", uuid.New().String())
		},
	}))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	profiles := e.Group("/profiles")
	profiles.POST("", ctrl.Create)
	profiles.GET("/:id", ctrl.GetByID)
	profiles.GET("/user/:user_id", ctrl.GetByUserID)
	profiles.PUT("/:id", ctrl.Update)
	profiles.DELETE("/:id", ctrl.Delete)

	return e
}

// setupGRPCServer builds the gRPC server and listener.
func setupGRPCServer(cfg *config.Config, svc *service.ProfileService) (*grpc.Server, net.Listener) {
	grpcAddr := net.JoinHostPort(cfg.GRPCHost, cfg.GRPCPort)
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to listen on gRPC port")
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			profilegrpc.RecoveryInterceptor(),
			profilegrpc.RequestIDInterceptor(),
			profilegrpc.LoggingInterceptor(),
		),
	)
	profileServer := profilegrpc.NewProfileServer(svc)
	types.RegisterProfileServiceServer(grpcServer, profileServer)

	return grpcServer, lis
}
