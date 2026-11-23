package main

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/epistemer/example/db"
	"github.com/epistemer/example/internal/confab"
	"github.com/epistemer/example/internal/settings"
)

// migrate creates the schema in an empty database.
func migrate(ctx context.Context, s *settings.Migrate, log *logrus.Logger) error {
	store, err := confab.Open(ctx, s.DBURL().Reveal())
	if err != nil {
		return err
	}
	defer store.Close()
	if err := store.Exec(ctx, db.Schema); err != nil {
		return err
	}
	log.Info("schema created")
	return nil
}
