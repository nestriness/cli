/*
Copyright © 2024 Nestri <>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed nestri.ascii
var art string

var (
	gpu int
	hdr bool
)

// rootCmd represents the base command when called without any subcommands
// For a good reference point, start here: https://github.com/charmbracelet/taskcli/blob/main/cmds.go
var rootCmd = &cobra.Command{
	Use:   "nestri",
	Short: "A CLI tool to manage your cloud gaming service",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// this is for the "nestri neofetch" subcommand, has no arguments
var neoFetchCmd = &cobra.Command{
	Use:   "neofetch",
	Short: "Show important system information",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		lipgloss.SetColorProfile(termenv.TrueColor)

		baseStyle := lipgloss.NewStyle().
			PaddingTop(1).
			PaddingRight(4).
			PaddingBottom(1).
			PaddingLeft(4)

		var (
			b      strings.Builder
			lines  = strings.Split(art, "\n")
			colors = []string{"#F8481C", "#F74127", "#F53B30", "#F23538", "#F02E40"}
			step   = len(lines) / len(colors)
		)

		for i, l := range lines {
			n := clamp(0, len(colors)-1, i/step)
			b.WriteString(colorize(colors[n], l))
			b.WriteRune('\n')
		}

		t := table.New().
			Border(lipgloss.HiddenBorder())

		t.Row(baseStyle.Render(b.String()), baseStyle.Render("System Info goes here"))

		fmt.Print(t)

		return nil
	},
}

// this is the "nestri run" subcommand, takes no arguments for now
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a game using nestri",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS != "linux" {
			//make sure os is linux
			fmt.Println("This command is only supported on Linux.")
			return nil
		}

		var game string
		if len(args) > 0 {
			game = args[0]
			viper.Set("game", game)
			viper.WriteConfig()
		} else {
			game = viper.GetString("game")
			if filepath.Ext(game) != ".exe" {
				return fmt.Errorf("Make sure the game is a .exe")
			}
			if game == "" {
				return fmt.Errorf("no game specified and no previous game selected")
			}
		}

		fmt.Printf("Running game: %s\n\n", game)
		if gpu > 0 {
			fmt.Print("Using gpu %s\n", gpu)
		}
		if hdr {
			fmt.Println("Enabling HDR mode")
		}

		//get linux version
		versionCmd := exec.Command("grep", "VERSION", "/etc/os-release")
		versionOutput, err := versionCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error getting linux version:")
		}
		fmt.Printf("Linux version:\n%s\n", string(versionOutput))

		//Step 1: change to games dir
		fmt.Println("changing to game dir.") //this is a temp command for debug as well as leads to a hardcoded dir

		HomeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("error getting home directory %v\n", err)
		}

		err = os.Chdir(fmt.Sprintf("%s/game", HomeDir))
		if err != nil {
			return fmt.Errorf("error changing directory: %v\n", err)
		}
		//verify we are in game dir
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %v\n", err)
		}
		fmt.Printf("Current directory: %s\n\n", dir)

		//list games dir
		listDir := exec.Command("ls", "-la", ".")
		listDirOutput, err := listDir.CombinedOutput()
		if err != nil {
			fmt.Errorf("error listing games: %v\n")
		}
		fmt.Printf("List of Games: \n%s\n", listDirOutput)

		//step 2: Generate a Session ID
		//generate id
		SID := exec.Command("bash", "-c", "head /dev/urandom | LC_ALL=C tr -dc 'a-zA-Z0-9' | head -c 16")

		//save output to variable
		output, err := SID.Output()
		if err != nil {
			fmt.Errorf("Error generating Session ID: %v\n", err)
		}
		sessionID := strings.TrimSpace(string(output))
		fmt.Printf("Your Session ID is: %s\n\n", sessionID)

		//step 3: Launch netris server
		fmt.Println("Installing Netris/Launching Netris Server\n")
		checkRunning := exec.Command("sudo", "docker", "ps", "-q", "-f", "name=netris")
		containerId, err := checkRunning.Output()
		if err != nil {
			return fmt.Errorf("error checking running Docker container: %v", err)
		}

		if len(containerId) == 0 {
			checkExisting := exec.Command("sudo", "docker", "ps", "-aq", "-f", "name=netris")
			containerId, err = checkExisting.Output()
			if err != nil {
				return fmt.Errorf("error checking for existing docker container: %v", err)
			}

			if len(containerId) == 0 {
				installCmd := exec.Command(
					"sudo", "docker", "run", "-d", "--gpus", "all", "--device=/dev/dri",
					"--name", "netris", "-it", "--entrypoint", "/bin/bash",
					"-e", fmt.Sprintf("SESSION_ID=%s", sessionID),
					"-v", fmt.Sprintf("%s:/game", dir), "-p", "8080:8080/udp",
					"--cap-add=SYS_NICE", "--cap-add=SYS_ADMIN", "ghcr.io/netrisdotme/netris/server:nightly",
				)
				installCmd.Stdout = os.Stdout
				installCmd.Stderr = os.Stderr

				if err := installCmd.Run(); err != nil {
					return fmt.Errorf("error running docker command: %v", err)
				}
			} else {
				startContainer := exec.Command("sudo", "docker", "start", "netris")
				startContainer.Stdout = os.Stdout
				startContainer.Stderr = os.Stderr

				if err := startContainer.Run(); err != nil {
					return fmt.Errorf("error starting existing Docker container: %v", err)
				}
			}
		}

		//main part of step 4:
		//start netris server

		fmt.Println("starting netris server\n\n")
		checkFileCmd := exec.Command("sudo", "docker", "exec", "netris", "ls", "-la", "/tmp")
		output, err = checkFileCmd.Output()
		if err != nil {
			return fmt.Errorf("error checking /tmp dir in docker container: %v\n", err)
		}

		if !strings.Contains(string(output), ".X11-unix") {
			startupCmd := exec.Command("sudo", "docker", "exec", "netris", "/etc/startup.sh", ">", "/dev/null", "&")
			startupCmd.Stdout = os.Stdout
			startupCmd.Stderr = os.Stderr

			if err := startupCmd.Run(); err != nil {
				return fmt.Errorf("error running startup command: %v\n", err)
			}

			for {
				time.Sleep(7 * time.Minute)
				output, err := checkFileCmd.Output()
				if err != nil {
					return fmt.Errorf("error checking /tmp directory in container: %v\n", err)
				}
				if strings.Contains(string(output), ".X11-unix") {
					break
				}
			}
		}

		gameCmd := fmt.Sprintf("netris-proton -pr %s", game)
		execCmd := exec.Command("sudo", "docker", "exec", "netris", gameCmd)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		if err := execCmd.Run(); err != nil {
			return fmt.Errorf("error executing game command in docker container: %v\n", err)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(neoFetchCmd)

	rootCmd.AddCommand(runCmd)

	runCmd.Flags().IntVar(&gpu, "gpu", 0, "Specify GPU index")
	runCmd.Flags().BoolVar(&hdr, "hdr", false, "Enable HDR mode")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {

		} else {
			fmt.Println("error reading config file: %v(you should be able to ignore this)", err)
		}
	}

	//If you want to add subcommands to run for example "netri run -fsr" do it like this
	// runCmd.Flags().BoolP("fsr", "f", false, "Run the Game with FSR enabled or not")
}

func colorize(c, s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c)).Render(s)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
