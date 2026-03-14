package cli

import (
	"fmt"
	"time"

	"github.com/AegirAexx/mdam/internal/search"
	"github.com/spf13/cobra"
)

var (
	searchTag           string
	searchType          string
	searchModifiedAfter string
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search across managed documents",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := ""
		if len(args) == 1 {
			query = args[0]
		}

		filters := search.Filters{
			Tag:  searchTag,
			Type: searchType,
		}
		if searchModifiedAfter != "" {
			t, err := time.Parse("2006-01-02", searchModifiedAfter)
			if err != nil {
				return fmt.Errorf("invalid date %q: expected YYYY-MM-DD", searchModifiedAfter)
			}
			filters.ModifiedAfter = t
		}

		results, err := search.Search(cfg.BaseDir, query, filters)
		if err != nil {
			return fmt.Errorf("search: %w", err)
		}

		if len(results) == 0 {
			cmd.Println("no results found")
			return nil
		}
		for _, r := range results {
			cmd.Printf("%s\t%s\n", r.Frontmatter.Title, r.Path)
		}
		return nil
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchTag, "tag", "", "filter by tag")
	searchCmd.Flags().StringVar(&searchType, "type", "", "filter by document type")
	searchCmd.Flags().StringVar(&searchModifiedAfter, "modified-after", "", "filter documents modified after date (YYYY-MM-DD)")
	rootCmd.AddCommand(searchCmd)
}
