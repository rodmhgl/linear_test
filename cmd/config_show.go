package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	ldcerr "github.com/rodmhgl/ldctl/internal/errors"
	"github.com/rodmhgl/ldctl/internal/config"
)

// maskToken returns a masked representation of an API token.
// If the token is 6 characters or fewer, the entire value is replaced with "***".
// Otherwise the first 3 and last 3 characters are shown with "..." in between.
func maskToken(token string) string {
	if len(token) <= 6 {
		return "***"
	}
	return token[:3] + "..." + token[len(token)-3:]
}

// configShowJSON is the JSON output structure for config show.
type configShowJSON struct {
	URL   configShowField `json:"url"`
	Token configShowField `json:"token"`
}

type configShowField struct {
	Value  string `json:"value"`
	Source string `json:"source"`
}

// newConfigShowCmd returns the cobra command for "ldctl config show".
func newConfigShowCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current ldctl configuration, including URL and token source.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow(cmd, flags, args)
		},
	}
}

// runConfigShow implements the config show logic. It is also invoked by the
// bare "ldctl config" command (see cmd/config.go).
func runConfigShow(cmd *cobra.Command, flags *rootFlags, _ []string) error {
	cfgPath, pathErr := config.ConfigPath()

	// Check file permissions before loading (non-fatal warning).
	if pathErr == nil {
		ok, permErr := config.FilePermissionsOK(cfgPath)
		if permErr == nil && !ok {
			fmt.Fprintf(cmd.ErrOrStderr(),
				"Warning: config file has overly permissive permissions.\n"+
					"Run: chmod 600 %s\n", cfgPath)
		}
	}

	result, err := config.Load()
	if err != nil {
		var ldcErr *ldcerr.Error
		// Surface "no configuration found" with a friendly message + exit 1.
		if isErr(err, ldcerr.ConfigError) {
			ldcErr, _ = err.(*ldcerr.Error)
			// Check if it's specifically "no configuration found".
			if ldcErr != nil && strings.Contains(ldcErr.Message, "no configuration found") {
				fmt.Fprintln(cmd.ErrOrStderr(),
					"No configuration found. Run 'ldctl config init' to get started.")
				return ldcerr.New(ldcerr.ConfigError, "no configuration found")
			}
			return ldcErr
		}
		return err
	}

	masked := maskToken(result.Config.Token)

	if flags.json {
		out := configShowJSON{
			URL: configShowField{
				Value:  result.Config.URL,
				Source: result.Source.URL,
			},
			Token: configShowField{
				Value:  masked,
				Source: result.Source.Token,
			},
		}
		b, marshalErr := json.MarshalIndent(out, "", "  ")
		if marshalErr != nil {
			return ldcerr.Newf(ldcerr.IOError, "failed to marshal JSON: %v", marshalErr)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", b)
		return nil
	}

	// Human-readable output with aligned columns.
	fmt.Fprintf(cmd.OutOrStdout(),
		"URL:   %-40s (%s)\n", result.Config.URL, result.Source.URL)
	fmt.Fprintf(cmd.OutOrStdout(),
		"Token: %-40s (%s)\n", masked, result.Source.Token)

	return nil
}

// isErr reports whether err is an *ldcerr.Error of the given type.
func isErr(err error, t ldcerr.Type) bool {
	if e, ok := err.(*ldcerr.Error); ok {
		return e.Type == t
	}
	return false
}
