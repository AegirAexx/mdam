package cli

import (
	"fmt"

	"github.com/AegirAexx/mdam/internal/template"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage document templates",
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		templates, err := template.Discover(cfg.TemplatesDir())
		if err != nil {
			return fmt.Errorf("discovering templates: %w", err)
		}

		// Always include built-ins in the listing.
		builtinNames := make(map[string]bool)
		for name := range template.BuiltinTemplates() {
			builtinNames[name] = true
		}
		userNames := make(map[string]bool)
		for _, t := range templates {
			userNames[t.Name] = true
		}

		for _, t := range templates {
			cmd.Println(t.Name)
		}
		for name := range builtinNames {
			if !userNames[name] {
				cmd.Printf("%s (built-in)\n", name)
			}
		}
		return nil
	},
}

var templateShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Display a template's content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Check user templates first.
		tmpl, err := template.Find(cfg.TemplatesDir(), name)
		if err != nil {
			// Fall back to built-in.
			builtins := template.BuiltinTemplates()
			content, ok := builtins[name]
			if !ok {
				return fmt.Errorf("template %q not found", name)
			}
			tmpl = template.Template{Name: name, Content: content}
		}

		cmd.Print(tmpl.Content)
		return nil
	},
}

func init() {
	templateCmd.AddCommand(templateListCmd, templateShowCmd)
	rootCmd.AddCommand(templateCmd)
}
