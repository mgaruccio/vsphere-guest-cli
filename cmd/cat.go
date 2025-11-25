package cmd

import (
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
	catGuestUser string
	catGuestPwd  string
)

var catCmd = &cobra.Command{
	Use:   "cat <remote-file>",
	Short: "Read a file from the guest VM and print to stdout",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		remotePath := args[0]
		
		if targetVMName == "" {
			return fmt.Errorf("--vm flag is required")
		}

		if catGuestUser == "" {
			catGuestUser = os.Getenv("GUEST_USER")
		}
		if catGuestPwd == "" {
			catGuestPwd = os.Getenv("GUEST_PASSWORD")
		}

		if catGuestUser == "" || catGuestPwd == "" {
			return fmt.Errorf("--guest-user and --guest-password (or GUEST_USER/GUEST_PASSWORD env vars) are required")
		}

		ctx := cmd.Context()
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
			Username: catGuestUser,
			Password: catGuestPwd,
		}

		if verbose {
			fmt.Printf("Initiating transfer for %s...\n", remotePath)
		}

		transfer, err := fileManager.InitiateFileTransferFromGuest(ctx, &auth, remotePath)
		if err != nil {
			return fmt.Errorf("failed to initiate file transfer: %w", err)
		}

		if verbose {
			fmt.Printf("Transfer URL: %s\n", transfer.Url)
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		}
		httpClient := &http.Client{Transport: tr}

		resp, err := httpClient.Get(transfer.Url)
		if err != nil {
			return fmt.Errorf("failed to download file: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("download failed with status: %s", resp.Status)
		}

		_, err = io.Copy(os.Stdout, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read file content: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(catCmd)
	catCmd.Flags().StringVar(&catGuestUser, "guest-user", "", "Guest OS Username")
	catCmd.Flags().StringVar(&catGuestPwd, "guest-password", "", "Guest OS Password")
}