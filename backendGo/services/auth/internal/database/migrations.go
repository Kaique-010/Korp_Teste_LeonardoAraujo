package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {
	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		return err
	}

	sort.Strings(files)

	log.Printf("migrations encontradas: %v", files)

	for _, file := range files {
		log.Printf("executando migration: %s", file)

		sql, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf(
				"erro ao ler migration %s: %w",
				file,
				err,
			)
		}

		if err := db.Exec(string(sql)).Error; err != nil {
			return fmt.Errorf(
				"erro na migration %s: %w",
				file,
				err,
			)
		}

		log.Printf("migration executada: %s", file)
	}

	return nil
}
