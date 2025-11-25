package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vmware/govmomi/vim25/types"
	"vsphere-guest-cli/pkg/input"
)

var (
	typeEnter bool
)

var typeCmd = &cobra.Command{

	Use:   "type <text>",

	Short: "Send keystrokes to the guest VM console",

	Args:  cobra.ExactArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {

		text := args[0]



		if targetVMName == "" {

			return fmt.Errorf("--vm flag is required")

		}



		if typeEnter {

			text += "\n"

		}



		codes, err := input.StringToUsbScanCodes(text)

		if err != nil {

			return err

		}



		if len(codes) == 0 {

			fmt.Println("No valid characters to send.")

			return nil

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



		// PutUsbScanCodes is on the VirtualMachine object

		_, err = vm.PutUsbScanCodes(ctx, types.UsbScanCodeSpec{

			KeyEvents: codes,

		})

		if err != nil {

			return fmt.Errorf("failed to send keystrokes: %w", err)

		}



		if verbose {

			fmt.Printf("Sent %d keystrokes to %s\n", len(codes), targetVMName)

		}



		return nil

	},

}

func init() {
	rootCmd.AddCommand(typeCmd)
	typeCmd.Flags().BoolVar(&typeEnter, "enter", false, "Append Enter key after text")
}
