package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"vsphere-guest-cli/pkg/vsphere"
)

var (
	cfgFile      string
	host         string
	user         string
	password     string
	insecure     bool
	datacenter   string
	targetVMName string
	verbose      bool
)

var rootCmd = &cobra.Command{
	Use:   "guest-cli",
	Short: "A CLI for interacting with vSphere guests via VM Tools",
	Long: `guest-cli allows you to run commands, transfer files, and interact with 
the console of Virtual Machines running on vSphere, primarily designed for 
AI agents and automation.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	envCfg := vsphere.GetEnvConfig()

	rootCmd.PersistentFlags().StringVar(&host, "host", envCfg.Host, "vSphere Host URL (e.g., https://vcsa.example.com/sdk) [Env: VSPHERE_HOST]")
	rootCmd.PersistentFlags().StringVar(&user, "user", envCfg.User, "vSphere Username [Env: VSPHERE_USER]")
	rootCmd.PersistentFlags().StringVar(&password, "password", envCfg.Password, "vSphere Password [Env: VSPHERE_PASSWORD]")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", envCfg.Insecure, "Skip SSL verification [Env: VSPHERE_INSECURE]")
	rootCmd.PersistentFlags().StringVar(&datacenter, "datacenter", envCfg.Datacenter, "vSphere Datacenter [Env: VSPHERE_DATACENTER]")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	
	// We will add the --vm flag to individual subcommands or here if it applies to all. 
	// Since "help" or "version" might not need it, we'll add it as a PersistentFlag but not mark it mandatory globally yet.
	rootCmd.PersistentFlags().StringVar(&targetVMName, "vm", "", "Target Virtual Machine Name")
}

func GetClient() (*vsphere.Client, error) {
	if host == "" || user == "" || password == "" {
		return nil, fmt.Errorf("host, user, and password are required (via flags or environment variables)")
	}

	return vsphere.NewClient(rootCmd.Context(), vsphere.ConnectionConfig{
		Host:       host,
		User:       user,
		Password:   password,
		Insecure:   insecure,
		Datacenter: datacenter,
	})
}
