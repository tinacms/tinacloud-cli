package logout

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
	"tina.io/tinacloud-cli/internal/posthog"
)

func Logout(args []string, cfg config.Config, track posthog.Tracker) error {
	flagSet := flag.NewFlagSet("logout", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud auth logout")
		fmt.Fprintln(os.Stderr, "\nLog out of a TinaCloud account")
	}
	_ = flagSet.Parse(args)

	// Best-effort upstream revocation: a failure should never block logout.
	tokens, err := auth.PeekTokens()
	if err == nil {
		if err := auth.RevokeToken(cfg, tokens.RefreshToken); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to revoke token upstream, it will expire on its own: %v\n", err)
		}
	} else if !errors.Is(err, auth.ErrUnauthenticated) {
		fmt.Fprintf(os.Stderr, "warning: failed to read token cache for revocation: %v\n", err)
	}

	if err := auth.ClearTokens(); err != nil {
		return fmt.Errorf("failed to clear auth tokens: %w", err)
	}

	track(posthog.UserLoggedOut())

	fmt.Println("✅️ Successfully signed out of your TinaCloud account")
	return nil
}
