/*
Copyright © 2024 Nestri <>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/netrisdotme/cli/pkg/specs"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

//go:embed nestri.ascii
var art string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nestri",
	Short: "A CLI tool to manage your cloud gaming service",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var neoFetchCmd = &cobra.Command{
	Use:   "neofetch",
	Short: "Show important system information",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		lipgloss.SetColorProfile(termenv.TrueColor)

		// baseStyle := lipgloss.NewStyle().
		// 	MarginTop(1).
		// 	MarginRight(4).
		// 	MarginBottom(1).
		// 	MarginLeft(4)

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
			Border(lipgloss.HiddenBorder()).BorderStyle(lipgloss.NewStyle().Width(3))

		info := &specs.Specs{}
		infoChan := make(chan specs.Specs, 1)
		var wg sync.WaitGroup
		wg.Add(1)
		go getSpecs(info, infoChan, &wg)
		wg.Wait()
		newInfo := <-infoChan

		t.Row(b.String(), newInfo.GPU)

		fmt.Print(t)

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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.AddCommand(neoFetchCmd)

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

func getSpecs(info *specs.Specs, infoChan chan specs.Specs, wg *sync.WaitGroup) {
	defer wg.Done()
	sys := specs.New()
	// info.Userhost = getUserHostname()
	// info.OS = getOSName()
	// info.Kernel = getKernelVersion()
	// info.Uptime = getUptime()
	// info.Shell = getShell()
	// info.CPU = getCPUName()
	// info.RAM = getMemStats()
	info.GPU, _ = sys.GetGPUInfo()
	// info.SystemArch, _ = getSystemArch()
	// info.DiskUsage, _ = getDiskUsage()
	infoChan <- *info
}
