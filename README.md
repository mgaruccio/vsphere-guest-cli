# guest-cli

A CLI tool for interacting with vSphere guests via VMware Tools, designed for AI agents and automation.

## Features

*   **Command Execution:** Run commands inside the guest OS (Windows/Linux) and capture output.
*   **File Transfer:** Upload and download files to/from the guest.
*   **Console Interaction:** Send keystrokes (HID scan codes) to the VM console.
*   **VM Listing:** List available VMs and their details.
*   **Zero-Network:** Works via VMware Tools (VMX channel), requiring no direct network connectivity to the guest IP.

## Installation

1.  **Build from source:**
    ```bash
    go build -o guest-cli main.go
    ```

2.  **Environment Setup:**
    Set the following environment variables to avoid passing them as flags:
    ```bash
    export VSPHERE_HOST="https://vcenter.example.com/sdk"
    export VSPHERE_USER="administrator@vsphere.local"
    export VSPHERE_PASSWORD="password"
    export VSPHERE_INSECURE=true  # If using self-signed certs
    export VSPHERE_DATACENTER="DatacenterName" # Optional, but recommended
    ```

## Usage

### List VMs
```bash
./guest-cli list
```

### Execute Command
Run `uname -a` on a Linux VM:
```bash
./guest-cli exec --vm "my-linux-vm" --guest-user "root" --guest-password "password" --cmd "uname -a"
```

Run `ipconfig` on a Windows VM:
```bash
./guest-cli exec --vm "my-windows-vm" --guest-user "Administrator" --guest-password "password" --cmd "ipconfig"
```

### File Transfer
**Upload:**
```bash
./guest-cli upload local-file.txt /tmp/remote-file.txt --vm "my-vm" --guest-user "root" --guest-password "password"
```

**Download:**
```bash
./guest-cli download /var/log/syslog ./syslog.txt --vm "my-vm" --guest-user "root" --guest-password "password"
```

### Read File (Cat)
Directly print file content to stdout:
```bash
./guest-cli cat /etc/hosts --vm "my-vm" --guest-user "root" --guest-password "password"
```

### Send Keystrokes
Type text into the VM console (useful for login screens):
```bash
./guest-cli type --vm "my-vm" "password123" --enter
```

## Requirements

*   **VMware Tools:** Must be installed and running on the guest VM.
*   **Guest Credentials:** You must have valid username/password for the Guest OS.
*   **Permissions:** vSphere user requires `VirtualMachine.GuestOperations` privileges.
