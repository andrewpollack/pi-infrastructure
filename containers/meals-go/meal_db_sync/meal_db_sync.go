package meal_db_sync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"meals/meal_collection"
	"os"

	"github.com/jackc/pgx/v5"
)

func SyncMeals() error {
	ctx := context.Background()

	// Fetch meal data from S3
	mealData, err := meal_collection.OpenFromS3()
	if err != nil {
		return fmt.Errorf("error fetching meal data from S3: %w", err)
	}
	defer mealData.Close()

	jsonFile, err := io.ReadAll(mealData)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var categories []meal_collection.Category
	if err := json.Unmarshal(jsonFile, &categories); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		return fmt.Errorf("POSTGRES_URL is not set")
	}
	conn, err := pgx.Connect(ctx, postgresURL)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	defer conn.Close(ctx)

	upsertQuery := `
        INSERT INTO recipes (name, category, url, ingredients)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (name) DO UPDATE
          SET category      = EXCLUDED.category,
              url           = EXCLUDED.url,
              ingredients   = EXCLUDED.ingredients,
              date_modified = now()
          WHERE (
            recipes.category    	IS DISTINCT FROM EXCLUDED.category
            OR recipes.url      	IS DISTINCT FROM EXCLUDED.url
            OR recipes.ingredients 	IS DISTINCT FROM EXCLUDED.ingredients
          )
    `

	for _, cat := range categories {
		for _, item := range cat.Items {
			ingJSON, err := json.Marshal(item.Ingredients)
			if err != nil {
				return fmt.Errorf("error marshaling ingredients: %w", err)
			}

			if item.URL == nil {
				item.URL = new(string)
			}

			res, err := conn.Exec(
				context.Background(),
				upsertQuery,
				item.Name,
				cat.Category,
				item.URL,
				ingJSON,
			)
			if err != nil {
				return fmt.Errorf("upsert failed for recipe '%s': %v", item.Name, err)
			}

			rowsAffected := res.RowsAffected()
			switch rowsAffected {
			case 1:
				fmt.Printf("Upserted recipe: [%s]\n", item.Name)
			case 0:
				fmt.Printf("No changes needed for recipe '%s'\n", item.Name)
			default:
				fmt.Printf("%d rows affected for recipe '%s' (unexpected)\n", rowsAffected, item.Name)
			}
		}
	}

	return nil
}
