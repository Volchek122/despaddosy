package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"despaddosy/pkg/config"
	"despaddosy/pkg/store"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type seedData struct {
	Admin struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	} `json:"admin"`
	Posts []struct {
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
	} `json:"posts"`
}

func main() {
	cfgPath := os.Getenv("SOCIALD_CONFIG")
	if cfgPath == "" {
		cfgPath = "configs/sociald.yaml"
	}
	cfg, err := config.LoadSocialConfig(cfgPath)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	db, err := store.NewPostgres(ctx, cfg.Database.DSN)
	if err != nil {
		panic(err)
	}
	if err := runMigrations(ctx, db, "migrations"); err != nil {
		panic(err)
	}
	if err := seed(ctx, db, "configs/seed/seed.json"); err != nil {
		panic(err)
	}
}

func runMigrations(ctx context.Context, db *pgxpool.Pool, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	files := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			return err
		}
		if _, err := db.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("migration %s failed: %w", file, err)
		}
	}
	return nil
}

func seed(ctx context.Context, db *pgxpool.Pool, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var seed seedData
	if err := json.Unmarshal(data, &seed); err != nil {
		return err
	}
	var adminID int
	err = db.QueryRow(ctx, "SELECT id FROM users WHERE email=$1", seed.Admin.Email).Scan(&adminID)
	if err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte(seed.Admin.Password), bcrypt.DefaultCost)
		err = db.QueryRow(ctx, "INSERT INTO users(email, username, password_hash, role) VALUES ($1,$2,$3,$4) RETURNING id",
			seed.Admin.Email, seed.Admin.Username, string(hash), seed.Admin.Role).Scan(&adminID)
		if err != nil {
			return err
		}
	}
	for _, post := range seed.Posts {
		_, _ = db.Exec(ctx, "INSERT INTO posts(user_id, content, image_url) VALUES ($1,$2,$3)", adminID, post.Content, post.ImageURL)
	}
	return nil
}
