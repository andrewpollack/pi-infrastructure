package meal_db_sync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"meals/meal_collection"
	"time"

	"github.com/jackc/pgx/v5"
)

type Config struct {
	BucketName  string
	BucketKey   string
	PostgresURL string
	CleanTable  bool
	LongLive    bool
}

func (c Config) SyncMeals() error {
	ctx := context.Background()

	// Fetch meal data from S3
	mealData, err := meal_collection.OpenFromS3(c.BucketName, c.BucketKey)
	if err != nil {
		return fmt.Errorf("error fetching meal data from S3: %w", err)
	}
	defer mealData.Close()

	jsonFile, err := io.ReadAll(mealData)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	var mealCollection meal_collection.MealCollection
	if err := json.Unmarshal(jsonFile, &mealCollection); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	if c.PostgresURL == "" {
		return fmt.Errorf("POSTGRES_URL is not set")
	}

	conn, err := pgx.Connect(ctx, c.PostgresURL)
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

	for _, item := range mealCollection {
		ingJSON, err := json.Marshal(item.Ingredients)
		if err != nil {
			return fmt.Errorf("error marshaling ingredients: %w", err)
		}

		if item.Category == nil {
			item.Category = new(string)
		}

		if item.URL == nil {
			item.URL = new(string)
		}

		res, err := conn.Exec(
			context.Background(),
			upsertQuery,
			item.Name,
			item.Category,
			item.URL,
			ingJSON,
		)
		if err != nil {
			return fmt.Errorf("upsert failed for recipe '%s': %v", item.Name, err)
		}

		rowsAffected := res.RowsAffected()
		switch rowsAffected {
		case 1:
			log.Printf("Upserted recipe: [%s]\n", item.Name)
		case 0:
			log.Printf("No changes needed for recipe '%s'\n", item.Name)
		default:
			log.Printf("%d rows affected for recipe '%s' (unexpected)\n", rowsAffected, item.Name)
		}
	}

	if c.CleanTable {
		var names []string
		for _, item := range mealCollection {
			names = append(names, item.Name)
		}

		// Delete rows that do not match any of the meal names
		deleteQuery := "DELETE FROM recipes WHERE NOT (name = ANY($1))"
		deleteRes, err := conn.Exec(ctx, deleteQuery, names)
		if err != nil {
			return fmt.Errorf("error deleting recipes not in sync: %w", err)
		}
		log.Printf("Deleted %d recipes that are not in the current meal collection\n", deleteRes.RowsAffected())
	}

	return nil
}

func (c Config) SyncMealsWrapper() error {
	if c.LongLive {
		for {
			if err := c.SyncMeals(); err != nil {
				log.Printf("Failed to sync meals: %v\n", err)
			}
			time.Sleep(20 * time.Minute)
		}
	} else {
		if err := c.SyncMeals(); err != nil {
			return fmt.Errorf("failed to sync meals: %v", err)
		}
	}

	return nil
}
