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

func runConfigShow(_ *cobra.Command, _ *rootFlags, _ []string) error {
	// stub: full implementation in a future issue
	return nil
}

func newConfigInitCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize ldctl configuration",
		RunE: func(_ *cobra.Command, _ []string) error {
			// stub
			return nil
		},
	}
}

func newConfigShowCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow(cmd, flags, args)
		},
	}
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
