// Package cmd implements the ldctl command-line interface.
package cmd

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	ldcerr "github.com/rodmhgl/ldctl/internal/errors"
	"github.com/rodmhgl/ldctl/internal/version"
)

// rootFlags holds global flag values shared across all subcommands.
type rootFlags struct {
	json    bool
	quiet   bool
	verbose bool
	version bool
}

// errVersionRequested is a sentinel error returned by PersistentPreRunE when
// --version is passed. It lets [Execute] distinguish a "success" early exit
// from a real error without calling [os.Exit] inside a command handler.
var errVersionRequested = stderrors.New("version requested")

func newRootCmd(v string) (*cobra.Command, *rootFlags) {
	flags := &rootFlags{}

	cmd := &cobra.Command{
		Use:   "ldctl",
		Short: "ldctl - LinkDing CLI client (version " + v + ")",
		Long: `ldctl is a command-line client for the LinkDing self-hosted
bookmark manager. It speaks to the LinkDing REST API using a
token you configure once with 'ldctl config init'.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if flags.version {
				info := version.Get()
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"ldctl version %s (commit %s, built %s, %s)\n",
					v, info.Commit, info.BuildDate, info.GoVersion,
				)
				return errVersionRequested
			}
			if flags.quiet && flags.verbose {
				return stderrors.New("cannot use --quiet and --verbose together")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.CompletionOptions.DisableDefaultCmd = true

	// Global persistent flags available to every subcommand.
	pf := cmd.PersistentFlags()
	pf.BoolVar(&flags.json, "json", false, "Output as JSON")
	pf.BoolVarP(&flags.quiet, "quiet", "q", false, "Suppress non-essential output")
	pf.BoolVarP(&flags.verbose, "verbose", "v", false, "Enable verbose output")
	pf.BoolVar(&flags.version, "version", false, "Print version and exit")

	// Override the help function only on the root command. Subcommands keep
	// cobra's default help template so their flags/usage render properly.
	cmd.SetHelpFunc(func(c *cobra.Command, _ []string) {
		if c == cmd {
			fmt.Fprint(c.OutOrStdout(), helpTemplate)
			return
		}
		// Fallback to cobra's built-in usage for subcommands.
		// Error is intentionally discarded — help output has no actionable path.
		_ = c.Usage()
	})

	// Register all command groups.
	cmd.AddCommand(
		newConfigCmd(flags),
		newBookmarksCmd(flags),
		newTagsCmd(flags),
		newBundlesCmd(flags),
		newAssetsCmd(flags),
		newUserCmd(flags),
		newVersionCmd(flags),
	)

	return cmd, flags
}

// NewRootCmd is the public constructor — used by tests to get a fresh tree.
func NewRootCmd(v string) *cobra.Command {
	cmd, _ := newRootCmd(v)
	return cmd
}

// HandleError writes the error to w in human or JSON format and returns the
// appropriate exit code. It is exported for testing; production code calls it
// via Execute().
func HandleError(err error, jsonMode bool, w io.Writer) int {
	var ldctlErr *ldcerr.Error
	if stderrors.As(err, &ldctlErr) {
		if jsonMode {
			_ = ldcerr.PrintJSON(w, ldctlErr)
		} else {
			ldcerr.PrintHuman(w, ldctlErr)
		}
		return ldctlErr.ExitCode()
	}
	fmt.Fprintf(w, "Error: %v\n", err)
	return 1
}

// Execute builds the command tree and runs it, returning an exit code.
func Execute() int {
	root, flags := newRootCmd(version.Version)
	if err := root.Execute(); err != nil {
		if stderrors.Is(err, errVersionRequested) {
			return 0
		}
		return HandleError(err, flags.json, os.Stderr)
	}
	return 0
}

// helpTemplate is used by [newRootCmd]'s SetHelpFunc for the root command.
// It is a const (not a var) to satisfy the gochecknoglobals linter.
// Each line is kept under 80 chars.
const helpTemplate = `ldctl - LinkDing CLI client

Usage:
  ldctl [command] [flags]

Available Commands:
  config      Manage ldctl configuration (init, show, test)
  bookmarks   Manage bookmarks (alias: bm)
  tags        Manage tags (list, get, create)
  bundles     Manage bundles (list, get, create, update, delete)
  assets      Manage bookmark assets (list, download, upload, delete)
  user        Show user profile
  version     Print version information

Global Flags:
  --json         Output as JSON
  -q, --quiet    Suppress non-essential output
  -v, --verbose  Enable verbose output
  --version      Print version and exit
  -h, --help     Show help

Examples:
  ldctl config init             # Configure API token
  ldctl bookmarks list          # List all bookmarks
  ldctl bm add https://go.dev   # Add a bookmark (alias)
  ldctl tags list               # List all tags
  ldctl version                 # Show version info

More: https://github.com/rodmhgl/ldctl
`

// jsonMarshal is a thin wrapper so callers get a formatted JSON string.
func jsonMarshal(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
