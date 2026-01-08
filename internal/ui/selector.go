package ui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"volume-migrator/internal/docker"
	"volume-migrator/internal/utils"
)

// SelectVolumes presents an interactive UI for selecting volumes to migrate
func SelectVolumes(volumes []docker.VolumeInfo) ([]docker.VolumeInfo, error) {
	if len(volumes) == 0 {
		return nil, errors.New("no volumes to select")
	}

	// Display summary
	fmt.Printf("\nDiscovered %d volume(s)\n\n", len(volumes))

	// Create a copy of volumes for selection
	selectionItems := make([]docker.VolumeInfo, len(volumes))
	copy(selectionItems, volumes)

	// All volumes start as selected by default
	for i := range selectionItems {
		selectionItems[i].Selected = true
	}

	// Interactive loop
	for {
		// Calculate total size of selected volumes
		totalSize := int64(0)
		selectedCount := 0
		for _, v := range selectionItems {
			if v.Selected {
				selectedCount++
				totalSize += v.SizeBytes
			}
		}

		// Create prompt templates
		templates := &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "→ [{{ if .Selected }}✓{{ else }} {{ end }}] {{ .Name | cyan }} ({{ .Container }}) {{ .MountPath }} {{ .Size }}",
			Inactive: "  [{{ if .Selected }}✓{{ else }} {{ end }}] {{ .Name }} ({{ .Container }}) {{ .MountPath }} {{ .Size }}",
			Selected: "{{ .Name | green }}",
			Details: `
--------- Volume Details ---------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Container:" | faint }}	{{ .Container }}
{{ "Mount Path:" | faint }}	{{ .MountPath }}
{{ "Size:" | faint }}	{{ .Size }}`,
		}

		// Create select prompt
		prompt := promptui.Select{
			Label:     fmt.Sprintf("Select volumes to migrate [%d of %d selected, %s total] (↑/↓ navigate, Space toggle, Enter confirm, A select all, N deselect all)", selectedCount, len(selectionItems), utils.FormatBytes(totalSize)),
			Items:     selectionItems,
			Templates: templates,
			Size:      10,
			Searcher: func(input string, index int) bool {
				volume := selectionItems[index]
				name := strings.Replace(strings.ToLower(volume.Name), " ", "", -1)
				input = strings.Replace(strings.ToLower(input), " ", "", -1)
				return strings.Contains(name, input)
			},
		}

		// Custom key handling through stdin
		idx, _, err := prompt.Run()
		if err != nil {
			// Check if user wants to quit
			if err == promptui.ErrInterrupt {
				return nil, errors.New("selection cancelled by user")
			}
			return nil, fmt.Errorf("selection failed: %w", err)
		}

		// Handle special keys through result parsing
		// Since promptui doesn't support custom keys easily, we'll use a simpler approach
		// Toggle selection for the current item
		selectionItems[idx].Selected = !selectionItems[idx].Selected

		// Show confirmation after selection changes
		if selectedCount == 0 && selectionItems[idx].Selected {
			// If this is the first selection, continue the loop
			continue
		}

		// Ask if user wants to continue or confirm
		confirmPrompt := promptui.Prompt{
			Label:     fmt.Sprintf("Selected %d volume(s). Continue selecting (c), Confirm (y), or Cancel (n)?", selectedCount),
			IsConfirm: false,
			Default:   "c",
		}

		response, err := confirmPrompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				return nil, errors.New("selection cancelled by user")
			}
			continue
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" {
			break
		} else if response == "n" || response == "no" {
			return nil, errors.New("selection cancelled by user")
		}
		// Continue loop for "c" or other responses
	}

	// Return only selected volumes
	var selected []docker.VolumeInfo
	for _, v := range selectionItems {
		if v.Selected {
			selected = append(selected, v)
		}
	}

	if len(selected) == 0 {
		return nil, errors.New("no volumes selected")
	}

	return selected, nil
}

// DisplayVolumeTable displays a simple table of volumes
func DisplayVolumeTable(volumes []docker.VolumeInfo) {
	if len(volumes) == 0 {
		fmt.Println("No volumes found.")
		return
	}

	// Print header
	fmt.Printf("\n%-25s %-20s %-25s %s\n", "VOLUME NAME", "CONTAINER", "MOUNT PATH", "SIZE")
	fmt.Println(strings.Repeat("-", 95))

	// Print volumes
	for _, v := range volumes {
		fmt.Printf("%-25s %-20s %-25s %s\n",
			truncate(v.Name, 25),
			truncate(v.Container, 20),
			truncate(v.MountPath, 25),
			v.Size,
		)
	}
	fmt.Println()
}

// truncate truncates a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
