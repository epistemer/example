package confab

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/epistemer/auth"
	"github.com/epistemer/errors"
	"github.com/jackc/pgx/v5"

	"github.com/epistemer/example/internal/models"
)

// Directory is auth's view of the staff table: the app owns its users.
type Directory struct{ S *Store }

func (d Directory) user(s models.Staff) auth.User {
	return auth.User{ID: auth.UserID(s.ID), Name: s.Email, Disabled: s.Disabled}
}

func (d Directory) Find(ctx context.Context, login string) (auth.User, error) {
	s, err := d.S.GetUserByEmail(ctx, login)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.User{}, auth.ErrNoUser
	}
	return d.user(s), err
}

func (d Directory) Get(ctx context.Context, id auth.UserID) (auth.User, error) {
	s, err := d.S.GetUser(ctx, string(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.User{}, auth.ErrNoUser
	}
	return d.user(s), err
}

func (d Directory) Address(ctx context.Context, id auth.UserID, ch auth.Channel) (string, error) {
	if ch != auth.ChannelEmail {
		return "", nil
	}
	s, err := d.S.GetUser(ctx, string(id))
	if err != nil {
		return "", err
	}
	return s.Email, nil
}

// Outbox delivers auth's messages as files, one per message, for
// development: codes never reach the log (auth's rule).
type Outbox struct{ Dir string }

func (o Outbox) Send(_ context.Context, m auth.Message) error {
	if err := os.MkdirAll(o.Dir, 0o700); err != nil {
		return err
	}
	body := fmt.Sprintf("To: %s\nKind: %d\nCode: %s\nToken: %s\nExpires: %s\n",
		m.To, m.Kind, m.Code, m.Token, m.Expires.Format(time.RFC3339))
	name := filepath.Join(o.Dir, fmt.Sprintf("%s-%d.txt", time.Now().Format("20060102-150405.000000"), m.Kind))
	return os.WriteFile(name, []byte(body), 0o600)
}
