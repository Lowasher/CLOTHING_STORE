package database

import (
    "database/sql"
    "fmt"
    "os"

    _ "github.com/glebarez/sqlite"

)

func InitDB(dbPath string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", dbPath+"?_foreign_keys=on")
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    // Чтение и выполнение скрипта миграций
    script, err := os.ReadFile("./migrations/init.sql")
    if err != nil {
        return nil, fmt.Errorf("failed to read migration file: %w", err)
    }

    _, err = db.Exec(string(script))
    if err != nil {
        return nil, fmt.Errorf("failed to execute migration script: %w", err)
    }

    return db, nil
}