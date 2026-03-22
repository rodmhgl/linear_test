package cmd

import (
	"github.com/spf13/cobra"
)

func newBookmarksCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bookmarks",
		Aliases: []string{"bm"},
		Short:   "Manage bookmarks",
		Long:    "Manage LinkDing bookmarks. Use 'bookmarks --help' for subcommands.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newBookmarksListCmd(flags),
		newBookmarksGetCmd(flags),
		newBookmarksAddCmd(flags),
		newBookmarksCheckCmd(flags),
		newBookmarksUpdateCmd(flags),
		newBookmarksArchiveCmd(flags),
		newBookmarksUnarchiveCmd(flags),
		newBookmarksDeleteCmd(flags),
	)

	return cmd
}

func newBookmarksListCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List bookmarks",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBookmarksGetCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get a bookmark by ID",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBookmarksAddCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Add a bookmark",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBookmarksCheckCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "check <url>",
		Short: "Check if a URL is already bookmarked",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBookmarksUpdateCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "update <id>",
		Short: "Update a bookmark",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBookmarksArchiveCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "archive <id>",
		Short: "Archive a bookmark",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBookmarksUnarchiveCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "unarchive <id>",
		Short: "Unarchive a bookmark",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newBookmarksDeleteCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a bookmark",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}
