package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/alekssaul/template/internal/config"
	"github.com/alekssaul/template/internal/model"
	"github.com/alekssaul/template/internal/store"
)

func main() {
	cfg := config.Load()

	db, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Seed API Keys
	fmt.Println("--- Seeding API Keys ---")
	writeKey, err := db.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{
		Name:       "Developer Write Key",
		Permission: model.PermissionWrite,
	})
	if err != nil {
		log.Fatalf("Failed to create write key: %v", err)
	}
	fmt.Printf("✓ Created Write Key: %s\n", writeKey.Key)

	readKey, err := db.CreateAPIKey(ctx, &model.CreateAPIKeyRequest{
		Name:       "Developer Read Key",
		Permission: model.PermissionRead,
	})
	if err != nil {
		log.Fatalf("Failed to create read key: %v", err)
	}
	fmt.Printf("✓ Created Read Key: %s\n", readKey.Key)

	// Seed Items
	fmt.Println("\n--- Seeding Items ---")
	for i := 1; i <= 10; i++ {
		item, err := db.CreateItem(ctx, &model.CreateItemRequest{
			Name:        fmt.Sprintf("Dummy Item %d", i),
			Description: fmt.Sprintf("This is the description for dummy item #%d. Seeded at %s", i, time.Now().Format(time.TimeOnly)),
		})
		if err != nil {
			log.Fatalf("Failed to create item %d: %v", i, err)
		}
		fmt.Printf("✓ Created Item ID: %d (%s)\n", item.ID, item.Name)
	}

	fmt.Println("\n✅ Seeding complete!")
}
