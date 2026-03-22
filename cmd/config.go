package cmd

import (
	"github.com/spf13/cobra"
)

func newConfigCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage ldctl configuration",
		Long:  "Manage ldctl configuration. Running 'config' alone shows the current config.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow(cmd, flags, args)
		},
	}

	cmd.AddCommand(
		newConfigInitCmd(flags),
		newConfigShowCmd(flags),
		newConfigTestCmd(flags),
	)

	return cmd
}



func newConfigTestCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Test connection to LinkDing",
		RunE: func(_ *cobra.Command, _ []string) error {
			// stub
			return nil
		},
	}
}
