package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-prod-project/internal/domain/listing"
)

// ListingRepository is a secondary (driven) adapter that implements
// the domain port listing.Repository using PostgreSQL.
//
// In Domain-Driven Design (DDD):
// 1. The infrastructure layer handles all database specifics (SQL, queries, connections).
// 2. It translates low-level DB records into domain aggregates using domain constructors (Reconstitute).
// 3. It translates persistence-specific errors (like sql.ErrNoRows) into domain errors (listing.ErrNotFound).
type ListingRepository struct {
	db *sql.DB
}

// NewListingRepository constructs a new PostgreSQL listing repository.
func NewListingRepository(db *sql.DB) *ListingRepository {
	return &ListingRepository{
		db: db,
	}
}

// Create inserts a new listing entity into PostgreSQL and populates generated ID and timestamp.
func (r *ListingRepository) Create(ctx context.Context, item *listing.Listing) error {
	query := `
		INSERT INTO listings (
			title,
			description,
			price,
			city
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	var (
		id        string
		createdAt time.Time
	)

	err := r.db.QueryRowContext(
		ctx,
		query,
		item.Title(),
		item.Description(),
		item.Price().Amount(),
		item.City(),
	).Scan(&id, &createdAt)

	if err != nil {
		return fmt.Errorf("create listing: %w", err)
	}

	item.SetID(id)
	item.SetCreatedAt(createdAt)

	return nil
}

// GetByID retrieves a listing by its ID from PostgreSQL and reconstitutes the domain aggregate.
func (r *ListingRepository) GetByID(ctx context.Context, id string) (*listing.Listing, error) {
	query := `
		SELECT
			id,
			title,
			description,
			price,
			city,
			created_at
		FROM listings
		WHERE id = $1
	`

	var (
		resID       string
		title       string
		description string
		priceAmount int64
		city        string
		createdAt   time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&resID,
		&title,
		&description,
		&priceAmount,
		&city,
		&createdAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, listing.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get listing: %w", err)
	}

	price, err := listing.NewPrice(priceAmount)
	if err != nil {
		return nil, fmt.Errorf("reconstitute price: %w", err)
	}

	return listing.Reconstitute(resID, title, description, price, city, createdAt)
}

// List queries the most recent listings from PostgreSQL and reconstitutes them into domain aggregates.
func (r *ListingRepository) List(ctx context.Context) ([]*listing.Listing, error) {
	query := `
		SELECT
			id,
			title,
			description,
			price,
			city,
			created_at
		FROM listings
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list listings: %w", err)
	}
	defer rows.Close()

	items := make([]*listing.Listing, 0)

	for rows.Next() {
		var (
			id          string
			title       string
			description string
			priceAmount int64
			city        string
			createdAt   time.Time
		)

		if err := rows.Scan(
			&id,
			&title,
			&description,
			&priceAmount,
			&city,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan listing: %w", err)
		}

		price, err := listing.NewPrice(priceAmount)
		if err != nil {
			return nil, fmt.Errorf("reconstitute price: %w", err)
		}

		item, err := listing.Reconstitute(id, title, description, price, city, createdAt)
		if err != nil {
			return nil, fmt.Errorf("reconstitute listing: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate listings: %w", err)
	}

	return items, nil
}

// Delete removes a listing by ID.
func (r *ListingRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM listings
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete listing: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if affected == 0 {
		return listing.ErrNotFound
	}

	return nil
}

