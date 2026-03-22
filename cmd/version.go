package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/rodmhgl/ldctl/internal/version"
)

func newVersionCmd(_ *rootFlags) *cobra.Command {
	var short bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information for ldctl.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if short {
				fmt.Fprintln(cmd.OutOrStdout(), version.Version)
				return nil
			}

			flags := cmd.InheritedFlags()
			jsonFlag := flags.Lookup("json")
			if jsonFlag != nil && jsonFlag.Value.String() == "true" {
				info := version.Get()
				b, err := jsonMarshal(info)
				if err != nil {
					return fmt.Errorf("marshal version info: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(b))
				return nil
			}

			fmt.Fprintln(cmd.OutOrStdout(), version.String())
			return nil
		},
	}

	cmd.Flags().BoolVar(&short, "short", false, "Print only the semver version")

	return cmd
}
