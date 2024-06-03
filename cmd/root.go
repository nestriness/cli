/*
Copyright © 2024 Nestri <>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

//go:embed nestri.ascii
var art string

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
	Args:  cobra.NoArgs,
	//For now just show the "help"
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
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
