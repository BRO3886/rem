package commands

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	rem "github.com/BRO3886/rem"
	"github.com/BRO3886/rem/internal/skills"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var skillsAgentFlag string

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage AI agent skills for rem",
	Long: `Install, uninstall, and check the status of the rem agent skill.

The rem skill teaches AI coding agents (Claude Code, Codex CLI, etc.)
how to use rem effectively. It includes command references, date parsing
docs, and usage examples.`,
}

func init() {
	rootCmd.AddCommand(skillsCmd)
}

// --- skills install ---

var skillsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install rem skill for AI agents",
	Long: `Installs the rem agent skill to the selected AI agent's skill directory.

Supported agents:
  claude    → ~/.claude/skills/rem-cli/    (Claude Code, Copilot, Cursor, OpenCode, Augment)
  codex     → ~/.agents/skills/rem-cli/    (Codex CLI, Copilot, Windsurf, OpenCode, Augment)
  openclaw  → ~/.openclaw/skills/rem-cli/  (OpenClaw)

Without --agent, shows an interactive picker to select which agents to install for.`,
	RunE: runSkillsInstall,
}

func init() {
	skillsInstallCmd.Flags().StringVarP(&skillsAgentFlag, "agent", "a", "", "Agent target: claude, codex, openclaw, or all")
	skillsCmd.AddCommand(skillsInstallCmd)
}

func runSkillsInstall(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	allTargets := skills.DefaultTargets(homeDir)
	targets, err := resolveTargets(allTargets, skillsAgentFlag, homeDir, "install")
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil // user cancelled
	}

	return installToTargets(rem.EmbeddedSkills, targets, homeDir)
}

func installToTargets(embeddedFS fs.FS, targets []skills.AgentTarget, homeDir string) error {
	green := color.New(color.FgGreen, color.Bold)

	for _, t := range targets {
		written, err := skills.Install(embeddedFS, t, Version)
		if err != nil {
			return fmt.Errorf("failed to install for %s: %w", t.Name, err)
		}

		green.Print("✓ ")
		fmt.Printf("Installed rem-cli skill to %s\n", skills.DisplayPath(skills.SkillDir(t), homeDir))
		fmt.Printf("  Files: %s\n", strings.Join(written, ", "))
	}

	fmt.Println("\nThe skill will be available in your next session.")
	return nil
}

// --- skills uninstall ---

var skillsUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall rem skill from AI agents",
	Long:  `Removes the rem agent skill from the selected AI agent's skill directory.`,
	RunE:  runSkillsUninstall,
}

func init() {
	skillsUninstallCmd.Flags().StringVarP(&skillsAgentFlag, "agent", "a", "", "Agent target: claude, codex, openclaw, or all")
	skillsCmd.AddCommand(skillsUninstallCmd)
}

func runSkillsUninstall(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	allTargets := skills.DefaultTargets(homeDir)
	targets, err := resolveTargets(allTargets, skillsAgentFlag, homeDir, "uninstall")
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil // user cancelled
	}

	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow)

	for _, t := range targets {
		removed, err := skills.Uninstall(t)
		if err != nil {
			return fmt.Errorf("failed to uninstall from %s: %w", t.Name, err)
		}
		if removed {
			green.Print("✓ ")
			fmt.Printf("Removed rem-cli skill from %s\n", skills.DisplayPath(skills.SkillDir(t), homeDir))
		} else {
			yellow.Print("- ")
			fmt.Printf("Not installed at %s\n", skills.DisplayPath(skills.SkillDir(t), homeDir))
		}
	}

	return nil
}

// --- skills status ---

var skillsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show skill installation status",
	RunE:  runSkillsStatus,
}

func init() {
	skillsCmd.AddCommand(skillsStatusCmd)
}

func runSkillsStatus(cmd *cobra.Command, args []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	allTargets := skills.DefaultTargets(homeDir)
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)

	fmt.Printf("rem-cli skill (binary %s):\n", Version)
	for _, t := range allTargets {
		displayDir := skills.DisplayPath(skills.SkillDir(t), homeDir)
		if !skills.IsInstalled(t) {
			red.Printf("  ✗ ")
			fmt.Printf("%-12s %s (not installed)\n", t.Name, displayDir)
			continue
		}
		installed := skills.InstalledVersion(t)
		if installed == "" {
			yellow.Printf("  ? ")
			fmt.Printf("%-12s %s (installed, unknown version)\n", t.Name, displayDir)
		} else if installed != Version {
			yellow.Printf("  ⚠ ")
			fmt.Printf("%-12s %s (installed %s, outdated)\n", t.Name, displayDir, installed)
		} else {
			green.Printf("  ✓ ")
			fmt.Printf("%-12s %s (installed %s)\n", t.Name, displayDir, installed)
		}
	}

	return nil
}

// --- shared helpers ---

// resolveTargets determines which agent targets to operate on.
// If --agent is specified, uses that directly.
// Otherwise, shows an interactive huh.MultiSelect picker.
func resolveTargets(allTargets []skills.AgentTarget, agentFlag, homeDir, action string) ([]skills.AgentTarget, error) {
	// If --agent flag provided, resolve directly
	if agentFlag != "" {
		return resolveAgentFlag(allTargets, agentFlag)
	}

	// Check if stdin is a TTY — if not, auto-detect
	if !isTTY() {
		detected := skills.DetectAgents(allTargets)
		if len(detected) == 0 {
			// Default to claude
			return allTargets[:1], nil
		}
		return detected, nil
	}

	// Interactive multi-select
	return runAgentPicker(allTargets, homeDir, action)
}

func resolveAgentFlag(allTargets []skills.AgentTarget, flag string) ([]skills.AgentTarget, error) {
	flag = strings.ToLower(strings.TrimSpace(flag))
	if flag == "all" {
		return allTargets, nil
	}
	for _, t := range allTargets {
		if t.Key == flag {
			return []skills.AgentTarget{t}, nil
		}
	}
	return nil, fmt.Errorf("unknown agent %q (valid: claude, codex, openclaw, all)", flag)
}

func runAgentPicker(allTargets []skills.AgentTarget, homeDir, action string) ([]skills.AgentTarget, error) {
	detected := skills.DetectAgents(allTargets)
	detectedKeys := make(map[string]bool)
	for _, d := range detected {
		detectedKeys[d.Key] = true
	}

	// Build options
	options := make([]huh.Option[string], len(allTargets))
	for i, t := range allTargets {
		label := fmt.Sprintf("%-12s (%s)", t.Name, skills.DisplayPath(t.BaseDir, homeDir))
		if action == "uninstall" && skills.IsInstalled(t) {
			v := skills.InstalledVersion(t)
			if v != "" {
				label += fmt.Sprintf(" — installed, %s", v)
			} else {
				label += " — installed"
			}
		}
		options[i] = huh.NewOption(label, t.Key)
	}

	// Pre-select based on context
	var preselected []string
	if action == "uninstall" {
		// Pre-check installed agents
		for _, t := range allTargets {
			if skills.IsInstalled(t) {
				preselected = append(preselected, t.Key)
			}
		}
	} else {
		// Pre-check detected agents
		for _, t := range allTargets {
			if detectedKeys[t.Key] {
				preselected = append(preselected, t.Key)
			}
		}
		// If none detected, pre-check all
		if len(preselected) == 0 {
			for _, t := range allTargets {
				preselected = append(preselected, t.Key)
			}
		}
	}

	var selected []string
	title := fmt.Sprintf("%s rem skill for which AI agents?",
		strings.ToUpper(action[:1])+action[1:])

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(title).
				Options(options...).
				Value(&selected).
				Validate(func(s []string) error {
					if len(s) == 0 {
						return fmt.Errorf("select at least one agent")
					}
					return nil
				}),
		),
	).WithTheme(huh.ThemeCatppuccin())

	// Set preselected values
	selected = preselected

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println("Cancelled.")
			return nil, nil
		}
		return nil, fmt.Errorf("selection error: %w", err)
	}

	// Map selected keys back to targets
	selectedMap := make(map[string]bool)
	for _, s := range selected {
		selectedMap[s] = true
	}

	var targets []skills.AgentTarget
	for _, t := range allTargets {
		if selectedMap[t.Key] {
			targets = append(targets, t)
		}
	}

	return targets, nil
}
