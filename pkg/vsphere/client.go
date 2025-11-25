package vsphere

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/soap"
)

type Client struct {
	Client *govmomi.Client
	Finder *find.Finder
}

// ConnectionConfig holds the parameters for connecting to vSphere
type ConnectionConfig struct {
	Host       string
	User       string
	Password   string
	Insecure   bool
	Datacenter string
}

// NewClient creates a new authenticated vSphere client
func NewClient(ctx context.Context, config ConnectionConfig) (*Client, error) {
	u, err := soap.ParseURL(config.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host URL: %w", err)
	}

	u.User = url.UserPassword(config.User, config.Password)

	// Handle insecure flag
	insecure := config.Insecure

	c, err := govmomi.NewClient(ctx, u, insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to create govmomi client: %w", err)
	}

	finder := find.NewFinder(c.Client, true)
	
	if config.Datacenter != "" {
		dc, err := finder.Datacenter(ctx, config.Datacenter)
		if err != nil {
			return nil, fmt.Errorf("failed to find datacenter %s: %w", config.Datacenter, err)
		}
		finder.SetDatacenter(dc)
	}

	return &Client{
		Client: c,
		Finder: finder,
	}, nil
}

// FindVM finds a Virtual Machine by name (path or name)
func (c *Client) FindVM(ctx context.Context, name string) (*object.VirtualMachine, error) {
	// Try finding by inventory path first
	vm, err := c.Finder.VirtualMachine(ctx, name)
	if err == nil {
		return vm, nil
	}

	// If path lookup fails, try finding by name across all datacenters
	// Note: This might be slower and ambiguous if names are not unique
	// but convenient for the user.
	// For now, let's stick to the Finder which searches recursively if a datacenter is set,
	// but initially, we might just want to set a default datacenter or search everywhere.
	
	// Let's try to find the default datacenter if we can, to narrow scope,
	// otherwise Finder usually requires a fairly specific path or a default DC to be set.
	
	// Optimization: If the user provides a simple name "my-vm", standard finder might fail
	// if it expects a path. However, Finder.VirtualMachine generally handles names well
	// if they are unique within the scope.
	
	// Let's rely on the Finder's default behavior for now.
	return vm, err
}

// Logout logs out of the vSphere session
func (c *Client) Logout(ctx context.Context) {
	if c.Client != nil {
		c.Client.Logout(ctx)
	}
}

// GetEnvConfig reads connection details from environment variables
func GetEnvConfig() ConnectionConfig {
	return ConnectionConfig{
		Host:       os.Getenv("VSPHERE_HOST"),
		User:       os.Getenv("VSPHERE_USER"),
		Password:   os.Getenv("VSPHERE_PASSWORD"),
		Insecure:   os.Getenv("VSPHERE_INSECURE") == "true",
		Datacenter: os.Getenv("VSPHERE_DATACENTER"),
	}
}
