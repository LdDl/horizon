package docs

import (
	"embed"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

// Embed a single file
//go:embed index.html
var f embed.FS

func PrepareStaticPage() func(*fiber.Ctx) error {
	return filesystem.New(filesystem.Config{
		Root: http.FS(f),
	})
}

// Embed a directory
//go:embed assets/*
var embedDirStatic embed.FS

func PrepareStaticAssets() func(*fiber.Ctx) error {
	return filesystem.New(filesystem.Config{
		Root:       http.FS(embedDirStatic),
		PathPrefix: "",
		Browse:     false,
	})
}
