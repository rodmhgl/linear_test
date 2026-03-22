package cmd

import (
	"github.com/spf13/cobra"
)

func newTagsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Manage tags",
		Long:  "Manage LinkDing tags.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newTagsListCmd(flags),
		newTagsGetCmd(flags),
		newTagsCreateCmd(flags),
	)

	return cmd
}

func newTagsListCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tags",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newTagsGetCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a tag by ID",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newTagsCreateCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "create <name>",
		Short: "Create a tag",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}
