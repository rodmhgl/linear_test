package cmd

import (
	"github.com/spf13/cobra"
)

func newBundlesCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bundles",
		Short: "Manage bundles",
		Long:  "Manage LinkDing bundles (tag collections).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newBundlesListCmd(flags),
		newBundlesGetCmd(flags),
		newBundlesCreateCmd(flags),
		newBundlesUpdateCmd(flags),
		newBundlesDeleteCmd(flags),
	)

	return cmd
}

func newBundlesListCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List bundles",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBundlesGetCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a bundle by ID",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBundlesCreateCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a bundle",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBundlesUpdateCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "update <id>",
		Short: "Update a bundle",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBundlesDeleteCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a bundle",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}
