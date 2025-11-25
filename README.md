# guest-cli

A CLI tool for interacting with vSphere guests via VMware Tools. Designed for AI agents and automation, it allows you to execute commands, transfer files, and type into the console without requiring network connectivity to the guest VM.

## Quick Start

**1. Configure Connection:**
Set these environment variables to avoid repeating them for every command:
```bash
# vSphere Connection
export VSPHERE_HOST="https://vcenter.example.com/sdk"
export VSPHERE_USER="administrator@vsphere.local"
export VSPHERE_PASSWORD="password"
export VSPHERE_INSECURE=true         # Required if using self-signed certs
export VSPHERE_DATACENTER="MyDC"     # Optional, but recommended

# Guest Credentials (Optional, avoids flags)
export GUEST_USER="ubuntu"
export GUEST_PASSWORD="password"
```

**2. Run Commands:**

**List VMs:**
```bash
./guest-cli list
```

**Execute a Command (Linux):**
```bash
./guest-cli exec --vm "ubuntu-vm" --cmd "uname -a"
# If env vars not set:
# ./guest-cli exec --vm "ubuntu-vm" --guest-user "ubuntu" --guest-password "pass" --cmd "uname -a"
```

**Execute with Sudo (Linux):**
```bash
./guest-cli exec --vm "ubuntu-vm" --cmd "apt update" --sudo
```

**Read a File:**
```bash
./guest-cli cat /etc/hostname --vm "ubuntu-vm"
```

## Detailed Usage

### Global Flags
*   `--verbose`, `-v`: Enable detailed debug logging (hidden by default).

### `exec` - Run Commands
Executes a process inside the guest and streams the output back to your terminal.
*   `--sudo`: (Linux only) Elevates the command using `sudo`. Assumes `guest-password` (or `GUEST_PASSWORD`) is the sudo password.
*   `--wait`: Wait for the command to finish (default `true`).
*   `--workdir`: Set working directory.

**Windows Example:**
```bash
./guest-cli exec --vm "win-vm" --guest-user "Administrator" --guest-password "pass" --cmd "ipconfig"
```

### `cp` - File Transfer
**Upload:**
```bash
./guest-cli upload local-file.txt /tmp/remote-file.txt --vm "my-vm"
```

**Download:**
```bash
./guest-cli download /var/log/syslog ./syslog.txt --vm "my-vm"
```

### `type` - Console Input
Send keystrokes directly to the VM console (HID events). Useful for typing passwords at login screens or interacting with non-networked VMs.
```bash
./guest-cli type --vm "my-vm" "mypassword" --enter
```

## Installation

### Download Binary
Download the latest release for your platform from the [GitHub Releases](https://github.com/mgaruccio/vsphere-guest-cli/releases) page.

### Build from Source
Requires Go 1.23+:
```bash
git clone https://github.com/mgaruccio/vsphere-guest-cli.git
cd vsphere-guest-cli
go build -o guest-cli main.go
```

## Features
*   **Zero-Network:** Works entirely via the VMware Tools (VMX) channel. No SSH/RDP or IP connectivity required.
*   **Cross-Platform:** Supports Linux and Windows guests.
*   **Secure:** Uses standard vSphere API authentication and Guest Operations privileges. Credential flags or env vars supported.
*   **Automation Ready:** Returns proper exit codes and structured output (optional verbose mode).

## Requirements
*   **VMware Tools:** Must be installed and running on the target guest VM.
*   **Guest Credentials:** Valid username/password for the guest OS.
*   **vSphere Permissions:** User requires `VirtualMachine.GuestOperations` privileges.
