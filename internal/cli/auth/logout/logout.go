package logout

import (
	"flag"
	"fmt"
	"os"

	"tina.io/tinacloud-cli/internal/auth"
	"tina.io/tinacloud-cli/internal/config"
)

func Logout(args []string, cfg config.Config) error {
	flagSet := flag.NewFlagSet("logout", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: tinacloud-cli auth logout")
		fmt.Fprintln(os.Stderr, "\nLog out of a TinaCloud account")
	}
	_ = flagSet.Parse(args)

	err := auth.ClearTokens()
	if err != nil {
		return fmt.Errorf("failed to clear auth tokens: %w", err)
	}

	fmt.Println("Successfully signed out of your TinaCloud account")
	return nil
}
