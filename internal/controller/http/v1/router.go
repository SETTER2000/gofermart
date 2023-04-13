// Package v1 реализует пути маршрутизации. Каждая служба в своем файле.
package v1

import (
	"github.com/SETTER2000/gofermart/config"
	"github.com/SETTER2000/gofermart/internal/client"
	"github.com/SETTER2000/gofermart/internal/usecase"
	"github.com/SETTER2000/gofermart/internal/usecase/encryp"
	"github.com/SETTER2000/gofermart/pkg/compress/gzip"
	"github.com/SETTER2000/gofermart/pkg/log/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

// NewRouter -.
// Swagger spec:
// @title       Gofermart
// @description URL shortener server
// @version     0.1.0
// @host        localhost:8080
// @BasePath    /
func NewRouter(handler *chi.Mux, l logger.Interface, s usecase.Gofermart, cfg *config.Config, c *client.AClient) {
	headerTypes := []string{
		"application/javascript",
		"application/x-gzip",
		"application/gzip",
		"application/json",
		"text/css",
		"text/html",
		"text/plain",
		"text/xml",
	}
	// AllowContentType применяет белый список запросов Content-Types,
	// в противном случае отвечает статусом 415 Unsupported Media Type.
	handler.Use(middleware.AllowContentType(headerTypes...))
	handler.Use(middleware.Compress(5, headerTypes...))
	handler.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	handler.Use(middleware.RequestID)
	handler.Use(middleware.Logger)
	handler.Use(middleware.Recoverer)
	handler.Use(render.SetContentType(render.ContentTypePlainText))
	handler.Use(encryp.EncryptionKeyCookie)
	handler.Use(gzip.DeCompressGzip)

	// Routers
	h := handler.Route("/api", func(r chi.Router) {
		r.Routes()
	})
	{
		newGofermartRoutes(h, s, l, cfg, c)
	}
}
