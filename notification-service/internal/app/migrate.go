//go:build migrate

package app

import (
	"time"

	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	_defaultAttempts = 50
	_defaultTimeout  = time.Second
)

//
//func init() {
//	databaseURL, ok := os.LookupEnv("PG_URL")
//	if !ok || len(databaseURL) == 0 {
//		log.Fatalf("migrate: environment variable not declared: PG_URL")
//	}
//
//	databaseURL += "?sslmode=disable"
//
//	var (
//		//attempts = _defaultAttempts
//		err error
//		m   *migrate.Migrate
//	)
//
//	//for attempts > 0 {
//	//	m, err = migrate.New("file://migrations", databaseURL)
//	//	if err == nil {
//	//		break
//	//	}
//	//	log.Printf("migrate: unable to connect to postgres database: %s", databaseURL)
//	//	log.Printf("Migrate: postgres is trying to connect, attempts left: %d", attempts)
//	//	time.Sleep(_defaultTimeout)
//	//	attempts--
//	//}
//
//	if err != nil {
//		log.Fatalf("Migrate: postgres connect error: %s", err)
//	}
//
//	err = m.Up()
//	defer m.Close()
//	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
//		log.Fatalf("Migrate: up error: %s", err)
//	}
//
//	if errors.Is(err, migrate.ErrNoChange) {
//		log.Printf("Migrate: no change")
//		return
//	}
//
//	log.Printf("Migrate: up success")
//}
