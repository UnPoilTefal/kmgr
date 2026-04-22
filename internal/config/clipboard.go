package config

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ReadClipboard returns the current clipboard content as a string.
// It uses OS-specific commands; no external Go library is required.
func ReadClipboard() (string, error) {
	cmd, err := clipboardCmd()
	if err != nil {
		return "", err
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("lecture du clipboard impossible (%s) : %w", cmd.Path, err)
	}
	content := strings.TrimRight(string(out), "\r\n")
	if content == "" {
		return "", fmt.Errorf("le clipboard est vide")
	}
	return content, nil
}

// clipboardCmd returns the OS-appropriate command to read the clipboard.
func clipboardCmd() (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("pbpaste"), nil

	case "windows":
		return exec.Command("powershell.exe", "-NoProfile", "-Command", "Get-Clipboard"), nil

	default: // Linux / WSL
		if isWSL() {
			return exec.Command("powershell.exe", "-NoProfile", "-Command", "Get-Clipboard"), nil
		}
		// Wayland takes priority over X11.
		if tool, ok := findCmd("wl-paste", "xclip", "xsel"); ok {
			switch tool {
			case "wl-paste":
				return exec.Command("wl-paste", "--no-newline"), nil
			case "xclip":
				return exec.Command("xclip", "-selection", "clipboard", "-o"), nil
			case "xsel":
				return exec.Command("xsel", "--clipboard", "--output"), nil
			}
		}
		return nil, fmt.Errorf(
			"aucun outil clipboard trouvé — installe wl-paste (Wayland), xclip ou xsel (X11)",
		)
	}
}

// isWSL detects whether we are running inside Windows Subsystem for Linux.
func isWSL() bool {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	lower := strings.ToLower(string(data))
	return strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl")
}

// findCmd returns the first tool name from candidates that is found in PATH.
func findCmd(candidates ...string) (string, bool) {
	for _, c := range candidates {
		if _, err := exec.LookPath(c); err == nil {
			return c, true
		}
	}
	return "", false
}
