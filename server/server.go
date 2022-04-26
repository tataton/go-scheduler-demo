package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// sla defines the response time limit for handler methods.
const sla = 500 * time.Millisecond

type Config struct {
	Addr string
	Repo storer
}

// New returns an http.Server configured with gin routes.
func New(c Config) *http.Server {
	h := handlerGroup{
		repo: c.Repo,
		sla:  sla,
	}
	router := gin.Default()
	router.Use(gin.Logger())
	router.GET("/availability", h.getAvailability)
	router.POST("/availability", h.postAvailability)
	router.DELETE("/availability", h.deleteAvailability)
	return &http.Server{
		Addr:    c.Addr,
		Handler: router,
	}
}
