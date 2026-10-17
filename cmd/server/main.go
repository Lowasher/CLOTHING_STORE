package main

import (
    "log/slog"
    "net/http"
    "os"

    "clothing-store/internal/config"
    "clothing-store/internal/database"
    "clothing-store/internal/handlers"
)

func main() {
    // Инициализация структурированного логирования в формате JSON
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(logger)

    cfg := config.Load()

    db, err := database.InitDB(cfg.DBPath)
    if err != nil {
        slog.Error("Failed to initialize database", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    mux := http.NewServeMux()

    // Регистрация эндпоинтов
    productHandler := &handlers.ProductHandler{DB: db}
    mux.Handle("/product", productHandler)

    slog.Info("Starting HTTP server", "port", cfg.ServerPort)
    if err := http.ListenAndServe(":"+cfg.ServerPort, mux); err != nil {
        slog.Error("Server failed", "error", err)
        os.Exit(1)
    }
}