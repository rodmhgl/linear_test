package cmd

import (
	"github.com/spf13/cobra"
)

func newAssetsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assets",
		Short: "Manage bookmark assets",
		Long:  "Manage LinkDing bookmark assets (cached page content).",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(
		newAssetsListCmd(flags),
		newAssetsDownloadCmd(flags),
		newAssetsUploadCmd(flags),
		newAssetsDeleteCmd(flags),
	)

	return cmd
}

func newAssetsListCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list <bookmark-id>",
		Short: "List assets for a bookmark",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newAssetsDownloadCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "download <asset-id>",
		Short: "Download an asset",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newAssetsUploadCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "upload <bookmark-id> <file>",
		Short: "Upload an asset",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}

func newAssetsDeleteCmd(_ *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <asset-id>",
		Short: "Delete an asset",
		RunE:  func(_ *cobra.Command, _ []string) error { return nil },
	}
}
