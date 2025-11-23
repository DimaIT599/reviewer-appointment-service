package internal

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"reviewer-appointment-service/internal/handlers"
	"reviewer-appointment-service/internal/services"
	"reviewer-appointment-service/internal/storage/postgresql"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	handler    *handlers.Handler
}

func NewServer(port string, storage *postgresql.Storage) *Server {
	userStorage := postgresql.NewUserStorage(storage)
	teamStorage := postgresql.NewTeamRepo(storage)
	prStorage := postgresql.NewPRRepo(storage)

	userService := services.NewUserService(userStorage)
	teamService := services.NewTeamService(teamStorage, userStorage)
	prService := services.NewPRService(prStorage, userStorage, teamStorage)

	statsRepo := postgresql.NewStatsRepo(storage)

	handler := handlers.NewHandler(userService, teamService, prService, statsRepo)

	router := setupRouter(handler)

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + port,
			Handler: router,
		},
		handler: handler,
	}
}

func setupRouter(h *handlers.Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/health", h.HealthCheck)

	r.POST("/team/add", h.CreateTeam)
	r.GET("/team/get", h.GetTeam)

	r.POST("/users/setIsActive", h.SetIsActive)
	r.GET("/users/getReview", h.GetUserReviewPRs)

	r.POST("/pullRequest/create", h.CreatePR)
	r.POST("/pullRequest/merge", h.MergePR)
	r.POST("/pullRequest/reassign", h.ReassignReviewer)

	r.GET("/stats", h.GetStats)

	return r
}

func (s *Server) Run() error {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server is running on %s", s.httpServer.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server stopped")
	return nil
}
