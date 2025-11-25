package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/vmware/govmomi/guest"
	"github.com/vmware/govmomi/vim25/types"
)

var (
	cpGuestUser string
	cpGuestPwd  string
)

var uploadCmd = &cobra.Command{
	Use:   "upload <local-path> <remote-path>",
	Short: "Upload a file to the guest VM",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		localPath := args[0]
		remotePath := args[1]
		return runTransfer(cmd.Context(), localPath, remotePath, true)
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download <remote-path> <local-path>",
	Short: "Download a file from the guest VM",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := args[0]
		localPath := args[1]
		return runTransfer(cmd.Context(), localPath, remotePath, false)
	},
}

func runTransfer(ctx context.Context, localPath, remotePath string, upload bool) error {
	if targetVMName == "" {
		return fmt.Errorf("--vm flag is required")
	}
	if cpGuestUser == "" || cpGuestPwd == "" {
		return fmt.Errorf("--guest-user and --guest-password are required")
	}

	c, err := GetClient()
	if err != nil {
		return err
	}
	defer c.Logout(ctx)

	vm, err := c.FindVM(ctx, targetVMName)
	if err != nil {
		return fmt.Errorf("failed to find VM %s: %w", targetVMName, err)
	}

	ops := guest.NewOperationsManager(c.Client.Client, vm.Reference())
	fileManager, err := ops.FileManager(ctx)
	if err != nil {
		return err
	}

	auth := types.NamePasswordAuthentication{
		Username: cpGuestUser,
		Password: cpGuestPwd,
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	httpClient := &http.Client{Transport: tr}

	if upload {
		// Upload
		f, err := os.Open(localPath)
		if err != nil {
			return fmt.Errorf("failed to open local file: %w", err)
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			return err
		}

		// Overwrite?
		attr := types.GuestFileAttributes{}
		urlStr, err := fileManager.InitiateFileTransferToGuest(ctx, &auth, remotePath, &attr, stat.Size(), true)
		if err != nil {
			return fmt.Errorf("failed to initiate upload: %w", err)
		}

		req, err := http.NewRequest("PUT", urlStr, f)
		if err != nil {
			return fmt.Errorf("failed to create upload request: %w", err)
		}
		req.ContentLength = stat.Size()
		
		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != 200 {
			return fmt.Errorf("upload failed with status: %s", resp.Status)
		}
		
		if verbose {
			fmt.Printf("Successfully uploaded %s to %s\n", localPath, remotePath)
		}

	} else {
		// Download
		transfer, err := fileManager.InitiateFileTransferFromGuest(ctx, &auth, remotePath)
		if err != nil {
			return fmt.Errorf("failed to initiate download: %w", err)
		}

		resp, err := httpClient.Get(transfer.Url)
		if err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("download failed with status: %s", resp.Status)
		}

		out, err := os.Create(localPath)
		if err != nil {
			return fmt.Errorf("failed to create local file: %w", err)
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to write local file: %w", err)
		}
		if verbose {
			fmt.Printf("Successfully downloaded %s to %s\n", remotePath, localPath)
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(downloadCmd)

	// Add flags to both
	for _, cmd := range []*cobra.Command{uploadCmd, downloadCmd} {
		cmd.Flags().StringVar(&cpGuestUser, "guest-user", "", "Guest OS Username")
		cmd.Flags().StringVar(&cpGuestPwd, "guest-password", "", "Guest OS Password")
	}
}
