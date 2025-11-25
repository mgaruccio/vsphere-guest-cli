package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available Virtual Machines",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		c, err := GetClient()
		if err != nil {
			return err
		}
		defer c.Logout(ctx)

		m := view.NewManager(c.Client.Client)

		v, err := m.CreateContainerView(ctx, c.Client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
		if err != nil {
			return fmt.Errorf("failed to create container view: %w", err)
		}
		defer v.Destroy(ctx)

		var vms []mo.VirtualMachine
		err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"name", "summary", "guest.ipAddress", "guest.guestFamily"}, &vms)
		if err != nil {
			return fmt.Errorf("failed to retrieve VMs: %w", err)
		}

		if verbose {
			fmt.Printf("Found %d VMs.\n", len(vms))
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATUS\tIP ADDRESS\tOS FAMILY")
		for _, vm := range vms {
			ip := "<unknown>"
			if vm.Guest != nil && vm.Guest.IpAddress != "" {
				ip = vm.Guest.IpAddress
			}
			
			family := "<unknown>"
			if vm.Guest != nil && vm.Guest.GuestFamily != "" {
				family = vm.Guest.GuestFamily
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", vm.Name, vm.Summary.OverallStatus, ip, family)
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
