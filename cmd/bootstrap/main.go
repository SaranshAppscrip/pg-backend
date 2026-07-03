package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/nivas/server/internal/config"
	"github.com/nivas/server/internal/database"
	"github.com/nivas/server/internal/domain"
	"github.com/nivas/server/internal/password"
	"github.com/nivas/server/internal/repository/postgres"
	"github.com/nivas/server/pkg/logger"
)

const defaultOrgID = "00000000-0000-0000-0000-000000000001"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	email := envOr("BOOTSTRAP_EMAIL", "owner@nivas.local")
	plainPassword := envOr("BOOTSTRAP_PASSWORD", "admin123")
	orgIDStr := envOr("BOOTSTRAP_ORG_ID", defaultOrgID)

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		log.Fatalf("invalid BOOTSTRAP_ORG_ID: %v", err)
	}

	if len(plainPassword) < 6 {
		log.Fatal("BOOTSTRAP_PASSWORD must be at least 6 characters")
	}

	ctx := context.Background()
	logr := logger.New(logger.Config{Level: "info", Format: "text"})

	pool, err := database.NewPool(ctx, cfg.Database, logr)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	repos := postgres.NewStoreBundle(pool)

	_, err = repos.Settings.Get(ctx, orgID)
	if err != nil {
		log.Fatalf("organization %s not found — run make migrate first", orgID)
	}

	profiles, err := repos.Staff.List(ctx, orgID)
	if err != nil {
		log.Fatalf("list staff: %v", err)
	}
	if len(profiles) > 0 {
		log.Printf("staff already exists for org %s — skipping bootstrap", orgID)
		return
	}

	hash, err := password.Hash(plainPassword)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	staff := &domain.Staff{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Email:          strings.TrimSpace(strings.ToLower(email)),
		PasswordHash:   hash,
		IsOwner:        true,
	}

	if err := repos.Staff.Create(ctx, staff); err != nil {
		log.Fatalf("create owner: %v", err)
	}

	log.Printf("bootstrap complete")
	log.Printf("  organization_id: %s", orgID)
	log.Printf("  email:           %s", staff.Email)
	log.Printf("  password:        (from BOOTSTRAP_PASSWORD)")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
