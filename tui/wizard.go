package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/AegirAexx/mdam/internal/config"
	"github.com/AegirAexx/mdam/internal/setup"
)

// wizardStep enumerates setup wizard screens.
type wizardStep int

const (
	stepBaseDir   wizardStep = iota // text input
	stepEditor                      // text input
	stepAuthor                      // text input
	stepTheme                       // selection
	stepNerdFonts                   // toggle
	stepExportDir                   // text input
	stepConfirm                     // preview + confirm
	stepCount                       // sentinel
)

var themeNames = []string{"tokyonight", "nord", "gruvbox", "catppuccin", "dracula"}

// WizardModel is the BubbleTea model for the first-run setup wizard.
type WizardModel struct {
	cfgPath string
	step    wizardStep
	theme   Theme
	width   int
	height  int

	// Collected values.
	baseDir   string
	editor    string
	author    string
	themeName string
	nerdFonts bool
	exportDir string

	// Input state.
	textInput textinput.Model
	cursor    int // selection cursor for theme/nerdFonts steps

	// Completion.
	done   bool
	config config.Config
	err    error
}

// NewWizard creates a wizard model with sensible defaults.
func NewWizard(cfgPath string) WizardModel {
	home, _ := os.UserHomeDir()

	ti := textinput.New()
	ti.CharLimit = 256
	ti.Focus()
	ti.Placeholder = filepath.Join(home, "notes")

	return WizardModel{
		cfgPath:   cfgPath,
		step:      stepBaseDir,
		theme:     NewTheme("tokyonight"),
		themeName: "tokyonight",
		editor:    os.Getenv("EDITOR"),
		exportDir: filepath.Join(home, "Downloads"),
		baseDir:   "",
		textInput: ti,
		width:     80,
		height:    24,
	}
}

// Init satisfies tea.Model.
func (m WizardModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update satisfies tea.Model.
func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	var cmd tea.Cmd
	if m.isTextStep() {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

// handleKey dispatches key presses based on the current step.
func (m WizardModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()

	// Global: quit.
	if k == "ctrl+c" {
		return m, tea.Quit
	}

	// Back navigation.
	if k == "esc" && m.step > stepBaseDir {
		m.step--
		m = m.syncInputToStep()
		return m, nil
	}

	switch m.step {
	case stepTheme:
		return m.updateThemeStep(k)
	case stepNerdFonts:
		return m.updateNerdFontsStep(k)
	case stepConfirm:
		return m.updateConfirmStep(k)
	default:
		return m.updateTextStep(msg)
	}
}

// updateTextStep handles text input steps (baseDir, editor, author, exportDir).
func (m WizardModel) updateTextStep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		m = m.saveCurrentValue()
		m.step++
		m = m.syncInputToStep()
		return m, nil
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// updateThemeStep handles j/k navigation and enter to select a theme.
func (m WizardModel) updateThemeStep(k string) (tea.Model, tea.Cmd) {
	switch k {
	case "j", "down":
		if m.cursor < len(themeNames)-1 {
			m.cursor++
		}
		m.themeName = themeNames[m.cursor]
		m.theme = NewTheme(m.themeName)
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
		m.themeName = themeNames[m.cursor]
		m.theme = NewTheme(m.themeName)
	case "enter":
		m.themeName = themeNames[m.cursor]
		m.theme = NewTheme(m.themeName)
		m.step++
		m.cursor = 0
		m = m.syncInputToStep()
	}
	return m, nil
}

// updateNerdFontsStep handles j/k toggle and enter to confirm.
func (m WizardModel) updateNerdFontsStep(k string) (tea.Model, tea.Cmd) {
	switch k {
	case "j", "k", "down", "up", " ":
		m.nerdFonts = !m.nerdFonts
		if m.nerdFonts {
			m.cursor = 1
		} else {
			m.cursor = 0
		}
	case "enter":
		m.step++
		m = m.syncInputToStep()
	}
	return m, nil
}

// updateConfirmStep handles the final confirmation screen.
func (m WizardModel) updateConfirmStep(k string) (tea.Model, tea.Cmd) {
	if k == "enter" {
		if err := m.writeAndScaffold(); err != nil {
			m.err = err
			return m, tea.Quit
		}
		cfg, err := config.LoadFrom(m.cfgPath)
		if err != nil {
			m.err = fmt.Errorf("reloading config: %w", err)
			return m, tea.Quit
		}
		m.config = cfg
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

// View satisfies tea.Model.
func (m WizardModel) View() string {
	var b strings.Builder

	// Header.
	title := m.theme.Accent.Render(" mdam — First-Run Setup")
	stepInfo := m.theme.Subtle.Render(fmt.Sprintf("Step %d of %d", m.step+1, int(stepCount)))
	b.WriteString(title + "  " + stepInfo + "\n\n")

	// Step content.
	switch m.step {
	case stepBaseDir:
		b.WriteString(m.renderTextStep("Base Directory", "Where should mdam store your documents?", m.textInput))
	case stepEditor:
		b.WriteString(m.renderTextStep("Editor", "Which editor should mdam use? (defaults to $EDITOR)", m.textInput))
	case stepAuthor:
		b.WriteString(m.renderTextStep("Author", "Your name, used in document metadata.", m.textInput))
	case stepTheme:
		b.WriteString(m.renderThemeStep())
	case stepNerdFonts:
		b.WriteString(m.renderNerdFontsStep())
	case stepExportDir:
		b.WriteString(m.renderTextStep("Export Directory", "Where should exported documents be saved?", m.textInput))
	case stepConfirm:
		b.WriteString(m.renderConfirmStep())
	}

	// Navigation hints.
	b.WriteString("\n")
	hint := "Enter: next"
	if m.step > stepBaseDir {
		hint += "  Esc: back"
	}
	b.WriteString(m.theme.Muted.Render(" "+hint))

	// Status bar.
	b.WriteString("\n")
	modeStyle := m.theme.StatusNewDoc.Render(" SETUP ")
	b.WriteString(modeStyle)

	return b.String()
}

// renderTextStep renders a labeled text input field.
func (m WizardModel) renderTextStep(label, desc string, ti textinput.Model) string {
	var b strings.Builder
	b.WriteString(m.theme.Accent.Render(" "+label) + "\n")
	b.WriteString(m.theme.Subtle.Render(" "+desc) + "\n\n")
	b.WriteString(" " + ti.View() + "\n")
	return b.String()
}

// renderThemeStep renders the theme selection list.
func (m WizardModel) renderThemeStep() string {
	var b strings.Builder
	b.WriteString(m.theme.Accent.Render(" Theme") + "\n")
	b.WriteString(m.theme.Subtle.Render(" Select a color theme for the TUI.") + "\n\n")

	maxWidth := m.width - 4
	if maxWidth < 20 {
		maxWidth = 20
	}
	for i, name := range themeNames {
		text := " " + name
		if i == m.cursor {
			b.WriteString(lipgloss.NewStyle().Reverse(true).Width(maxWidth).Render(text))
		} else {
			b.WriteString(m.theme.FileNormal.Render(text))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// renderNerdFontsStep renders the nerd fonts toggle.
func (m WizardModel) renderNerdFontsStep() string {
	var b strings.Builder
	b.WriteString(m.theme.Accent.Render(" Nerd Fonts") + "\n")
	b.WriteString(m.theme.Subtle.Render(" Does your terminal use a Nerd Font?") + "\n\n")

	maxWidth := m.width - 4
	if maxWidth < 20 {
		maxWidth = 20
	}
	options := []string{"No", "Yes"}
	for i, opt := range options {
		text := " " + opt
		if i == m.cursor {
			b.WriteString(lipgloss.NewStyle().Reverse(true).Width(maxWidth).Render(text))
		} else {
			b.WriteString(m.theme.FileNormal.Render(text))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// renderConfirmStep renders the final config preview.
func (m WizardModel) renderConfirmStep() string {
	var b strings.Builder
	b.WriteString(m.theme.Accent.Render(" Configuration Preview") + "\n")
	b.WriteString(m.theme.Subtle.Render(fmt.Sprintf(" Will be saved to: %s", m.cfgPath)) + "\n")
	b.WriteString(m.theme.Muted.Render(" You can edit this file at any time.") + "\n\n")

	yaml := wizardConfigYAML(m)
	for _, line := range strings.Split(yaml, "\n") {
		b.WriteString(" " + m.theme.FileNormal.Render(line) + "\n")
	}
	return b.String()
}

// isTextStep returns true if the current step uses a text input.
func (m WizardModel) isTextStep() bool {
	switch m.step {
	case stepBaseDir, stepEditor, stepAuthor, stepExportDir:
		return true
	}
	return false
}

// saveCurrentValue stores the text input value for the current step.
func (m WizardModel) saveCurrentValue() WizardModel {
	val := strings.TrimSpace(m.textInput.Value())
	switch m.step {
	case stepBaseDir:
		if val == "" {
			home, _ := os.UserHomeDir()
			val = filepath.Join(home, "notes")
		}
		m.baseDir = expandHome(val)
	case stepEditor:
		m.editor = val
	case stepAuthor:
		m.author = val
	case stepExportDir:
		if val == "" {
			home, _ := os.UserHomeDir()
			val = filepath.Join(home, "Downloads")
		}
		m.exportDir = expandHome(val)
	}
	return m
}

// syncInputToStep resets the text input for the new step.
func (m WizardModel) syncInputToStep() WizardModel {
	home, _ := os.UserHomeDir()
	m.textInput.Reset()
	switch m.step {
	case stepBaseDir:
		m.textInput.Placeholder = filepath.Join(home, "notes")
		if m.baseDir != "" {
			m.textInput.SetValue(m.baseDir)
		}
	case stepEditor:
		m.textInput.Placeholder = os.Getenv("EDITOR")
		if m.editor != "" {
			m.textInput.SetValue(m.editor)
		}
	case stepAuthor:
		m.textInput.Placeholder = "Your Name"
		if m.author != "" {
			m.textInput.SetValue(m.author)
		}
	case stepExportDir:
		m.textInput.Placeholder = filepath.Join(home, "Downloads")
		if m.exportDir != "" {
			m.textInput.SetValue(m.exportDir)
		}
	case stepTheme:
		for i, name := range themeNames {
			if name == m.themeName {
				m.cursor = i
				break
			}
		}
	case stepNerdFonts:
		if m.nerdFonts {
			m.cursor = 1
		} else {
			m.cursor = 0
		}
	}
	m.textInput.Focus()
	return m
}

// expandHome expands a leading ~/ to the user's home directory.
func expandHome(path string) string {
	if len(path) < 2 || path[:2] != "~/" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[2:])
}

// wizardConfigYAML generates the YAML config string from collected values.
func wizardConfigYAML(m WizardModel) string {
	var b strings.Builder
	b.WriteString("# mdam configuration\n\n")
	b.WriteString(fmt.Sprintf("editor: %q\n", m.editor))
	b.WriteString(fmt.Sprintf("author: %q\n", m.author))
	b.WriteString(fmt.Sprintf("base_dir: %s\n", m.baseDir))
	b.WriteString(fmt.Sprintf("export_dir: %s\n", m.exportDir))
	b.WriteString(fmt.Sprintf("theme: %s\n", m.themeName))
	b.WriteString(fmt.Sprintf("nerd_fonts: %v\n", m.nerdFonts))
	b.WriteString("\njournal:\n")
	b.WriteString("  auto_create: true\n")
	return b.String()
}

// writeAndScaffold writes the config file, creates dirs, seeds templates, and
// ensures scratch and todo files exist.
func (m WizardModel) writeAndScaffold() error {
	yaml := wizardConfigYAML(m)
	if err := os.MkdirAll(filepath.Dir(m.cfgPath), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	if err := os.WriteFile(m.cfgPath, []byte(yaml), 0o644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}
	if err := setup.ScaffoldDirs(m.baseDir); err != nil {
		return fmt.Errorf("scaffolding dirs: %w", err)
	}
	templatesDir := filepath.Join(m.baseDir, ".templates")
	if err := setup.SeedTemplates(templatesDir); err != nil {
		return fmt.Errorf("seeding templates: %w", err)
	}
	if err := setup.EnsureScratch(filepath.Join(m.baseDir, "scratch.md")); err != nil {
		return fmt.Errorf("ensuring scratch: %w", err)
	}
	if err := setup.EnsureTodo(filepath.Join(m.baseDir, "todo.md")); err != nil {
		return fmt.Errorf("ensuring todo: %w", err)
	}
	return nil
}

// RunWizard launches the setup wizard TUI, returning the loaded config on success.
func RunWizard(cfgPath string) (config.Config, error) {
	m := NewWizard(cfgPath)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return config.Config{}, fmt.Errorf("wizard: %w", err)
	}
	wm := result.(WizardModel)
	if wm.err != nil {
		return config.Config{}, wm.err
	}
	if !wm.done {
		return config.Config{}, fmt.Errorf("setup cancelled")
	}
	return wm.config, nil
}
