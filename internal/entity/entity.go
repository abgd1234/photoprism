/*
Package entity contains models for data storage based on GORM.

See http://gorm.io/docs/ for more information about GORM.

Additional information concerning data storage can be found in our Developer Guide:

https://github.com/photoprism/photoprism/wiki/Storage
*/
package entity

import (
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/photoprism/photoprism/internal/event"
)

var log = event.Log
var resetFixturesOnce sync.Once

func logError(result *gorm.DB) {
	if result.Error != nil {
		log.Error(result.Error.Error())
	}
}

type Types map[string]interface{}

// List of database entities and their table names.
var Entities = Types{
	"accounts":        &Account{},
	"files":           &File{},
	"files_share":     &FileShare{},
	"files_sync":      &FileSync{},
	"photos":          &Photo{},
	"descriptions":    &Description{},
	"places":          &Place{},
	"locations":       &Location{},
	"cameras":         &Camera{},
	"lenses":          &Lens{},
	"countries":       &Country{},
	"albums":          &Album{},
	"photos_albums":   &PhotoAlbum{},
	"labels":          &Label{},
	"categories":      &Category{},
	"photos_labels":   &PhotoLabel{},
	"keywords":        &Keyword{},
	"photos_keywords": &PhotoKeyword{},
	"links":           &Link{},
}

// WaitForMigration waits for the database migration to be successful.
func (list Types) WaitForMigration() {
	attempts := 100

	for name := range list {
		for i := 0; i <= attempts; i++ {
			if err := Db().Raw(fmt.Sprintf("DESCRIBE `%s`", name)).Scan(&struct{}{}).Error; err == nil {
				log.Debugf("entity: table %s migrated", name)
				break
			} else {
				log.Debugf("entity: %s", err.Error())
			}

			if i == attempts {
				panic("migration failed")
			}

			time.Sleep(50 * time.Millisecond)
		}
	}
}

// Drop migrates all database tables of registered entities.
func (list Types) Migrate() {
	for _, entity := range list {
		if err := Db().AutoMigrate(entity).Error; err != nil {
			panic(err)
		}
	}
}

// Drop drops all database tables of registered entities.
func (list Types) Drop() {
	for _, entity := range list {
		if err := Db().DropTableIfExists(entity).Error; err != nil {
			panic(err)
		}
	}
}

// MigrateDb creates all tables and inserts default entities as needed.
func MigrateDb() {
	Entities.Migrate()
	Entities.WaitForMigration()

	CreateUnknownPlace()
	CreateUnknownCountry()
	CreateUnknownCamera()
	CreateUnknownLens()
}

// DropTables drops database tables for all known entities.
func DropTables() {
	Entities.Drop()
}

// ResetDb drops database tables for all known entities and re-creates them with fixtures.
func ResetDb(testFixtures bool) {
	DropTables()
	MigrateDb()

	if testFixtures {
		CreateTestFixtures()
	}
}

// InitTestFixtures resets the database and test fixtures once.
func InitTestFixtures() {
	resetFixturesOnce.Do(func() {
		ResetDb(true)
	})
}

// InitTestDb connects to and completely initializes the test database incl fixtures.
func InitTestDb(dsn string) *Gorm {
	if HasDbProvider() {
		return nil
	}

	db := &Gorm{
		Driver: "mysql",
		Dsn:    dsn,
	}

	SetDbProvider(db)
	InitTestFixtures()

	return db
}
