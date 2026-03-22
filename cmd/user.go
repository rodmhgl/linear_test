package cmd

import (
	"github.com/spf13/cobra"
)

func newUserCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Show user profile",
		Long:  "Show the LinkDing user profile. Running 'user' alone shows the profile.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserProfile(cmd, flags, args)
		},
	}

	cmd.AddCommand(newUserProfileCmd(flags))

	return cmd
}

// runUserProfile is called by both 'user' bare and 'user profile'.
func runUserProfile(_ *cobra.Command, _ *rootFlags, _ []string) error {
	// stub: full implementation in a future issue
	return nil
}

func newUserProfileCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "profile",
		Short: "Show user profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserProfile(cmd, flags, args)
		},
	}
}
