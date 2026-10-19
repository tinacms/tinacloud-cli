package logout

import (
	"errors"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/command"
	"tina.io/tinacloud-cli/internal/config"
)

func Command(cfg config.Config) *command.Command {
	return &command.Command{
		Name:        "logout",
		Description: "Log out of a TinaCloud account",
		Run: func(args []string) error {
			return runLogout(cfg)
		},
	}
}

func runLogout(cfg config.Config) error {
	// Best-effort upstream revocation: a failure should never block logout.
	tokens, err := auth.PeekTokens()
	if err == nil {
		if err := auth.RevokeToken(cfg, tokens.RefreshToken); err != nil {
			fmt.Fprintf(os.Stderr, "failed to revoke token: %v\n", err)
		}
	} else if !errors.Is(err, auth.ErrUnauthenticated) {
		fmt.Fprintf(os.Stderr, "failed to read token cache: %v\n", err)
	}

	if err := auth.ClearTokens(); err != nil {
		return fmt.Errorf("failed to clear auth tokens: %w", err)
	}

	fmt.Println("✅️ Successfully signed out of your TinaCloud account")
	return nil
}
