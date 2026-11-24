// Code generated from sqlc's request by github.com/epistemer/models. DO NOT EDIT.

package models

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// RegisterTypes registers every enum of the schema, and its array, with
// conn's type map, so pgx can encode and scan []EventStatus. For
// pgxpool.Config.AfterConnect.
func RegisterTypes(ctx context.Context, conn *pgx.Conn) error {
	types, err := conn.LoadTypes(ctx, []string{
		"public.event_status", "public._event_status",
		"public.talk_level", "public._talk_level",
		"public.talk_format", "public._talk_format",
		"public.talk_state", "public._talk_state",
		"public.talk_topic", "public._talk_topic",
	})
	if err != nil {
		return err
	}
	conn.TypeMap().RegisterTypes(types)
	return nil
}
