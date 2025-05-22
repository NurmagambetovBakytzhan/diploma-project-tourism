// Package postgres implements postgres connection.
package postgres

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"io/ioutil"
	"log"
	"time"
	"tourism-backend/config"
	"tourism-backend/internal/entity"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres -.
type Postgres struct {
	Conn *gorm.DB
}

// New -.
func New(url string, opts ...Option) (*Postgres, error) {

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	pg := &Postgres{
		Conn: db,
	}

	return pg, nil
}

func (p *Postgres) Connect(cfg *config.Config) error {
	conn, err := gorm.Open(postgres.Open(cfg.URL),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   cfg.PG.TablePrefix,
				SingularTable: false,
			}})
	if err != nil {
		return err
	}
	err = p.Conn.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", cfg.PG.TablePrefix[:len(cfg.PG.TablePrefix)-1])).Error

	p.Conn = conn
	return nil
}

func (p *Postgres) Migrate() error {
	err := p.Conn.AutoMigrate(
		&entity.Tour{},
		&entity.Image{},
		&entity.Video{},
		&entity.Panorama{},
		&entity.User{},
		&entity.TourEvent{},
		&entity.Purchase{},
		&entity.TourCategory{},
		&entity.TourLocation{},
		&entity.Category{},
		&entity.UserFavorites{},
		&entity.UserActivity{},
	)
	if err != nil {
		log.Println("Migrating entities to Postgres - err: %w", err)
		return err
	}
	query, err := ioutil.ReadFile("pkg/postgres/create_categories.sql")
	if err != nil {
		log.Println("Error reding Create_categories.sql")
		panic(err)
	}
	if err := p.Conn.Exec(string(query)).Error; err != nil {
		log.Println("Error Executing create_categories.sql")
		panic(err)
	}
	if err != nil {
		log.Println("Failed to create schema:", err)
	}

	queryTourEmbeddings, err := ioutil.ReadFile("pkg/postgres/create_tour_embeddings.sql")
	if err != nil {
		log.Println("Error reading create_tour_embeddings.sql")
		panic(err)
	}
	if err := p.Conn.Exec(string(queryTourEmbeddings)).Error; err != nil {
		log.Println("Error executing tourembeddings")
		panic(err)
	}
	return nil
}
