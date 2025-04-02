package meal_collection

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func ReadMealCollectionFromDB(postgresURL string, recipeCreatedCutoff int64) (MealCollection, error) {
	// Temporary types just for DB scans and JSON unmarshaling.
	type DBIngredient struct {
		Item     string  `json:"item"`
		Quantity float64 `json:"quantity"`
		Unit     string  `json:"unit"`
		Aisle    string  `json:"aisle"`
	}

	type DBRecipe struct {
		ID           int            `json:"id"`
		Category     string         `json:"category"`
		Name         string         `json:"name"`
		URL          string         `json:"url"`
		Ingredients  []DBIngredient `json:"ingredients"`
		DateCreated  time.Time      `json:"date_created"`
		DateModified time.Time      `json:"date_modified"`
		Enabled      bool           `json:"enabled"`
	}

	if postgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is not set")
	}

	conn, err := pgx.Connect(context.Background(), postgresURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer func() {
		if err := conn.Close(context.Background()); err != nil {
			fmt.Printf("error closing connection: %v\n", err)
		}
	}()

	rows, err := conn.Query(context.Background(), `
		SELECT id, category, name, url, ingredients, date_created, date_modified, enabled
		FROM recipes
		WHERE date_created < to_timestamp($1)
	`, recipeCreatedCutoff)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var recipes []DBRecipe

	for rows.Next() {
		var (
			r              DBRecipe
			rawIngredients []byte
		)
		if err := rows.Scan(
			&r.ID,
			&r.Category,
			&r.Name,
			&r.URL,
			&rawIngredients,
			&r.DateCreated,
			&r.DateModified,
			&r.Enabled,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		if err := json.Unmarshal(rawIngredients, &r.Ingredients); err != nil {
			return nil, fmt.Errorf("unmarshal failed: %v", err)
		}
		recipes = append(recipes, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	var mealCollection MealCollection
	for _, recipe := range recipes {
		// Convert DBIngredients -> Ingredients
		var ingredients []Ingredient
		for _, dbIng := range recipe.Ingredients {
			ingredients = append(ingredients, Ingredient{
				Name:     dbIng.Item,
				Quantity: dbIng.Quantity,
				Unit:     Unit(dbIng.Unit),
				Aisle:    Aisle(dbIng.Aisle),
			})
		}

		meal := Meal{
			Name:        recipe.Name,
			URL:         &recipe.URL,
			Ingredients: ingredients,
			Disabled:    !recipe.Enabled,
			Category:    &recipe.Category,
		}

		mealCollection = append(mealCollection, meal)
	}

	// Sort meals by name
	sort.Slice(mealCollection, func(i, j int) bool {
		return strings.ToLower(mealCollection[i].Name) < strings.ToLower(mealCollection[j].Name)
	})

	return mealCollection, nil
}

type MealUpdate struct {
	Name     string `json:"name"`
	Disabled bool   `json:"disabled"`
}

func UpdateMealsInDB(postgresURL string, updates []MealUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	if postgresURL == "" {
		return fmt.Errorf("POSTGRES_URL is not set")
	}

	conn, err := pgx.Connect(context.Background(), postgresURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer func() {
		if err := conn.Close(context.Background()); err != nil {
			fmt.Printf("error closing connection: %v\n", err)
		}
	}()

	// Build slices for names and the desired enabled state.
	// Our DB stores an "enabled" boolean, so we set enabled = !Disabled.
	names := make([]string, len(updates))
	enabledStates := make([]bool, len(updates))
	for i, update := range updates {
		names[i] = update.Name
		enabledStates[i] = !update.Disabled
	}

	// Update the recipes table using unnest to update multiple rows in one query.
	_, err = conn.Exec(context.Background(), `
		UPDATE recipes
		SET enabled = t.enabled
		FROM (
			SELECT unnest($1::text[]) AS name, unnest($2::boolean[]) AS enabled
		) AS t
		WHERE recipes.name = t.name
	`, names, enabledStates)
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}

	return nil
}

func ReadExtraItemsFromDB(postgresURL string) ([]ExtraItem, error) {
	type DBItem struct {
		ID           int       `json:"id"`
		Aisle        string    `json:"aisle"`
		Name         string    `json:"name"`
		DateCreated  time.Time `json:"date_created"`
		DateModified time.Time `json:"date_modified"`
	}

	if postgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL is not set")
	}

	conn, err := pgx.Connect(context.Background(), postgresURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	defer func() {
		if err := conn.Close(context.Background()); err != nil {
			fmt.Printf("error closing connection: %v\n", err)
		}
	}()

	rows, err := conn.Query(context.Background(), `
        SELECT
            id,
            aisle,
            name,
            date_created,
            date_modified
        FROM item
	`)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var items []DBItem

	for rows.Next() {
		var i DBItem

		err := rows.Scan(
			&i.ID,
			&i.Aisle,
			&i.Name,
			&i.DateCreated,
			&i.DateModified,
		)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}

		items = append(items, i)
	}

	// Sort items by name
	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	var itemsOut []ExtraItem
	for _, item := range items {
		itemsOut = append(itemsOut, ExtraItem{
			Name:  item.Name,
			Aisle: Aisle(item.Aisle),
			ID:    item.ID,
		})
	}

	return itemsOut, nil
}

type Action string

const (
	Add    Action = "Add"
	Update Action = "Update"
	Delete Action = "Delete"
)

type FEItem struct {
	Name  string `json:"Name"`
	Aisle Aisle  `json:"Aisle"`
}
type FEExtraItem struct {
	Action Action `json:"Action"`
	Old    FEItem `json:"Old"`
	New    FEItem `json:"New"`
}

func UpdateExtraItemsInDB(postgresURL string, updates []FEExtraItem) error {
	if len(updates) == 0 {
		return nil
	}

	if postgresURL == "" {
		return fmt.Errorf("POSTGRES_URL is not set")
	}

	conn, err := pgx.Connect(context.Background(), postgresURL)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer func() {
		if err := conn.Close(context.Background()); err != nil {
			fmt.Printf("error closing connection: %v\n", err)
		}
	}()

	for _, update := range updates {
		switch update.Action {
		case Add:
			_, err = conn.Exec(context.Background(), `
				INSERT INTO item (aisle, name)
				VALUES ($1, $2)
			`, update.New.Aisle, update.New.Name)
			if err != nil {
				return fmt.Errorf("query failed: %v", err)
			}
		case Update:
			_, err = conn.Exec(context.Background(), `
				UPDATE item
				SET aisle = $1, name = $2
				WHERE aisle = $3 AND name = $4
			`, update.New.Aisle, update.New.Name, update.Old.Aisle, update.Old.Name)
			if err != nil {
				return fmt.Errorf("query failed: %v", err)
			}
		case Delete:
			_, err = conn.Exec(context.Background(), `
				DELETE FROM item
				WHERE aisle = $1 AND name = $2
			`, update.Old.Aisle, update.Old.Name)
			if err != nil {
				return fmt.Errorf("query failed: %v", err)
			}
		default:
			return fmt.Errorf("unknown action: %s", update.Action)
		}
	}

	return nil
}
