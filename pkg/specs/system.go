package specs

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type SysSpecs struct {
	osx string
}

func New() *SysSpecs {
	return &SysSpecs{osx: runtime.GOOS}
}

func (s SysSpecs) GetGPUInfo() (string, error) {
	var output []byte
	var err error

	switch s.osx {
	case "windows":
		output, err = exec.Command("wmic", "path", "win32_VideoController", "get", "name").Output()
		if err != nil {
			return "", fmt.Errorf("error retrieving GPU information on Windows: %v", err)
		}
	case "darwin":
		output, err = exec.Command("system_profiler", "SPDisplaysDataType").Output()
		if err != nil {
			return "", fmt.Errorf("error retrieving GPU information on macOS: %v", err)
		}
	case "linux":
		output, err = exec.Command("lspci", "-vnn").Output()
		if err != nil {
			return "", fmt.Errorf("error retrieving GPU information on Linux: %v", err)
		}
	default:
		return "", fmt.Errorf("error: GPU information retrieval not implemented for %s", runtime.GOOS)
	}

	outputStr := strings.TrimSpace(string(output))

	if s.osx == "windows" {
		lines := strings.Split(outputStr, "\r\n")[1:]
		gpuName := strings.TrimSpace(strings.Join(lines, " "))
		return gpuName, nil
	}

	if s.osx == "darwin" {
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Chipset Model:") {
				fields := strings.Split(line, ":")
				if len(fields) >= 2 {
					gpuName := strings.TrimSpace(fields[1])
					return gpuName, nil
				}
			}
		}
		return "", fmt.Errorf("error parsing GPU information on macOS")
	}

	lines := strings.Split(outputStr, "\n")

	for _, line := range lines {
		if strings.Contains(line, "VGA compatible controller") {
			fields := strings.Fields(line)
			if len(fields) > 2 {
				gpuName := strings.Join(fields[2:], " ")
				return gpuName, nil
			}
		}
	}
	return "", fmt.Errorf("error parsing GPU information on Linux")
}
