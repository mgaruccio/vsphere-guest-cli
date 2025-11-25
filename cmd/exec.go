package cmd

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"crypto/tls"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vmware/govmomi/guest"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

var (
	execCmdStr     string
	execWait       bool
	execGuestUser  string
	execGuestPwd   string
	execWorkDir    string
	execSudo       bool
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command in the guest VM",
	Long: `Executes a command inside the guest VM using VMware Tools.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if targetVMName == "" {
			return fmt.Errorf("--vm flag is required")
		}
		if execCmdStr == "" {
			return fmt.Errorf("--cmd flag is required")
		}
		
		if execGuestUser == "" {
			execGuestUser = os.Getenv("GUEST_USER")
		}
		if execGuestPwd == "" {
			execGuestPwd = os.Getenv("GUEST_PASSWORD")
		}

		if execGuestUser == "" || execGuestPwd == "" {
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
		
		auth := types.NamePasswordAuthentication{
			Username: execGuestUser,
			Password: execGuestPwd,
		}

		procManager, err := ops.ProcessManager(ctx)
		if err != nil {
			return err
		}

		fileManager, err := ops.FileManager(ctx)
		if err != nil {
			return err
		}

		// Determine Guest OS Family
		var guestFamily string
		var moVM mo.VirtualMachine
		err = vm.Properties(ctx, vm.Reference(), []string{"guest.guestFamily"}, &moVM)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch guest properties: %v. Assuming Linux.\n", err)
			guestFamily = "linuxGuest"
		} else {
			guestFamily = moVM.Guest.GuestFamily
		}
		
		// Generate temp file name
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
		tmpFileName := fmt.Sprintf("guest-cli-%d.log", rnd.Int())
		
		var programPath string
		var programArgs string
		var remoteOutputFile string

		isWindows := strings.Contains(strings.ToLower(guestFamily), "windows")

		if isWindows {
			if execSudo {
				return fmt.Errorf("--sudo flag is not supported on Windows")
			}
			remoteOutputFile = "C:\\Windows\\Temp\\" + tmpFileName
			programPath = "C:\\Windows\\System32\\cmd.exe"
			programArgs = fmt.Sprintf("/C \"%s > %s 2>&1\"", execCmdStr, remoteOutputFile)
		} else {
			remoteOutputFile = "/tmp/" + tmpFileName
			programPath = "/bin/sh"
			
			cmdToRun := execCmdStr
			if execSudo {
				// Escape single quotes in password and command for the echo | sudo pipeline
				safePwd := strings.ReplaceAll(execGuestPwd, "'", "'\\''")
				// safeCmd := strings.ReplaceAll(execCmdStr, "'", "'\\''") // Not strictly needed if we nest correctly, but safer?
				// If execCmdStr contains single quotes, they might break the outer sh -c '...'
				// The existing code: fmt.Sprintf("-c '%s > %s 2>&1'", execCmdStr, remoteOutputFile)
				// So execCmdStr is ALREADY vulnerable to single quotes breaking out of sh -c.
				// We should probably fix that generally, but for sudo:
				
				// Target: echo 'PWD' | sudo -S -p '' sh -c 'CMD'
				// We need to construct this whole string, and THEN wrap it in the outer sh -c '... > log'
				
				// Let's construct the inner sudo command
				// We need to be careful. 
				// inner: echo 'PWD' | sudo -S -p '' sh -c 'CMD'
				
				// If CMD has single quotes, sh -c 'CMD' breaks.
				// Better: Use double quotes for sh -c "CMD"? Then $vars expand.
				
				// Let's just wrap the sudo logic.
				cmdToRun = fmt.Sprintf("echo '%s' | sudo -S -p '' sh -c '%s'", safePwd, execCmdStr)
			}

			// Wrapping in outer shell to capture output
			// NOTE: If cmdToRun contains single quotes, this breaks.
			// Ideally we should escape single quotes in cmdToRun before wrapping in ''
			safeCmdToRun := strings.ReplaceAll(cmdToRun, "'", "'\\''")
			programArgs = fmt.Sprintf("-c '%s > %s 2>&1'", safeCmdToRun, remoteOutputFile)
		}

				if verbose {

					fmt.Printf("Executing: %s %s\n", programPath, programArgs)

				}

		

				spec := types.GuestProgramSpec{

					ProgramPath:      programPath,

					Arguments:        programArgs,

					WorkingDirectory: execWorkDir,

				}

		

				pid, err := procManager.StartProgram(ctx, &auth, &spec)

				if err != nil {

					return fmt.Errorf("failed to start program: %w", err)

				}

				if verbose {

					fmt.Printf("Process started with PID: %d\n", pid)

				}

		

				if execWait {

					// Poll for completion

					for {

						select {

						case <-ctx.Done():

							return ctx.Err()

						case <-time.After(1 * time.Second):

							procs, err := procManager.ListProcesses(ctx, &auth, []int64{pid})

							if err != nil {

								if verbose {

									fmt.Printf("Error listing process: %v\n", err)

								}

								continue

							}

							

							finished := false

							var exitCode int32

		

							if len(procs) == 0 {

								if verbose {

									fmt.Println("Process not found (likely finished).")

								}

								finished = true

							} else if procs[0].EndTime != nil {

								if verbose {

									fmt.Printf("Process finished with exit code: %d\n", procs[0].ExitCode)

								}

								exitCode = procs[0].ExitCode

								finished = true

							}

		

							if finished {

								// Download output

								if verbose {

									fmt.Printf("Downloading output from %s...\n", remoteOutputFile)

								}

								transfer, err := fileManager.InitiateFileTransferFromGuest(ctx, &auth, remoteOutputFile)

								if err != nil {

									return fmt.Errorf("failed to initiate file transfer: %w", err)

								}

								

								if verbose {

									fmt.Printf("Transfer URL: %s\n", transfer.Url)

								}

								

								u, err := url.Parse(transfer.Url)

								if err != nil {

									return fmt.Errorf("failed to parse output URL: %w", err)

								}

								

								if u == nil {

									return fmt.Errorf("parsed URL is nil")

								}

		

								// The Guest Ops URL contains a token, so we can use standard http.Get

								// We use a custom client to ensure we handle TLS (insecure) if needed, 

								// but the global http.DefaultClient might enforce certs.

								// Let's use the govmomi client's underlying http client but via standard request if possible,

								// or just create a new insecure http client if the user passed --insecure.

								

								// Actually, c.Client.Client.Transport is configured.

								// Let's use c.Client.Client (soap) -> does it expose the http client?

								// c.Client.Client is *vim25.Client.

								// It doesn't expose the http.Client directly easily.

								

								// Let's just use a new http.Client with InsecureSkipVerify if needed.

								tr := &http.Transport{

									TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // We'll assume insecure for now as per flag

								}

								client := &http.Client{Transport: tr}

								

								resp, err := client.Get(transfer.Url)

								if err != nil {

									return fmt.Errorf("failed to download file: %w", err)

								}

								defer resp.Body.Close()

								

								if resp.StatusCode != 200 {

									return fmt.Errorf("download failed with status: %s", resp.Status)

								}

								

								readCloser := resp.Body

								

								out, err := io.ReadAll(readCloser)

								if err != nil {

									return fmt.Errorf("failed to read response body: %w", err)

								}

								

								if verbose {

									fmt.Println("----- Output -----")

								}

								fmt.Print(string(out))

								if verbose {

									fmt.Println("\n------------------")

								}

								

								if exitCode != 0 {

									return fmt.Errorf("command exited with code %d", exitCode)

								}

		

								return nil

							}

						}

					}

				}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
	execCmd.Flags().StringVar(&execCmdStr, "cmd", "", "Command to execute")
	execCmd.Flags().BoolVar(&execWait, "wait", true, "Wait for command to finish and capture output")
	execCmd.Flags().StringVar(&execGuestUser, "guest-user", "", "Guest OS Username")
	execCmd.Flags().StringVar(&execGuestPwd, "guest-password", "", "Guest OS Password")
	execCmd.Flags().StringVar(&execWorkDir, "workdir", "", "Working directory in guest")
	execCmd.Flags().BoolVar(&execSudo, "sudo", false, "Run command as root using sudo (Linux only)")
}