// Command confab is the neozp example app: a conference organiser's back
// office. Its settings are declared in
// internal/settings/schema; run go generate after changing them.
package main

//go:generate go tool genx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/epistemer/cfgx"

	"github.com/epistemer/example/internal/logging"
	"github.com/epistemer/example/internal/settings"
)

func main() {
	// Interrupt or SIGTERM cancels the context, so commands can stop
	// cleanly; a second one stops the program at once.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctx.Done()
		stop()
	}()
	err := newRoot().ExecuteContext(ctx)
	stop()
	if err != nil {
		os.Exit(1)
	}
}

func newRoot() *cobra.Command {
	log := logging.New()
	root := &cobra.Command{Use: "confab", Short: "Confab, a conference organiser's back office", SilenceUsage: true}
	// cfgx's warnings, such as unknown keys in the config file, go to the
	// same logger as the rest.
	// auth, handlers and i18n log through log/slog's default logger: send
	// it to logrus too, not only cfgx's warnings (findings F14).
	slog.SetDefault(logging.Slog(log))
	settings.BindRoot(root, cfgx.WithLogger(logging.Slog(log)))
	root.AddCommand(
		settings.ServeCommand(&cobra.Command{Use: "serve", Short: "Run the HTTP server"},
			func(ctx context.Context, s *settings.Serve) error {
				if err := logging.Configure(log, s); err != nil {
					return err
				}
				return serve(ctx, s, log)
			}),
		settings.MigrateCommand(&cobra.Command{Use: "migrate", Short: "Create the database schema"},
			func(ctx context.Context, s *settings.Migrate) error { return migrate(ctx, s, log) }),
		settings.SeedCommand(&cobra.Command{Use: "seed", Short: "Add the organiser and demo data"},
			func(ctx context.Context, s *settings.Seed) error { return seed(ctx, s, log) }),
		settings.ConfigCommand(),
		&cobra.Command{Use: "version", Short: "Print the version", Args: cobra.NoArgs,
			Run: func(cmd *cobra.Command, _ []string) { fmt.Fprintln(cmd.OutOrStdout(), version()) }},
	)
	return root
}

// version is the module's version and VCS revision, from the build info.
func version() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	v := bi.Main.Version
	for _, s := range bi.Settings {
		if s.Key == "vcs.revision" && len(s.Value) >= 12 {
			v += " (" + s.Value[:12] + ")"
		}
	}
	return v
}
