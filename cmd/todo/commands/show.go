package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show [id]",
		Short: "Show a single task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q", args[0])
			}
			store, err := openStore()
			if err != nil {
				return err
			}
			defer store.Close()

			t, err := store.Get(id)
			if err != nil {
				return err
			}

			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "id:       %d\n", t.ID)
			fmt.Fprintf(w, "title:    %s\n", t.Title)
			fmt.Fprintf(w, "status:   %s\n", t.Status)
			fmt.Fprintf(w, "priority: %s\n", t.Priority)
			if len(t.Tags) > 0 {
				fmt.Fprintf(w, "tags:     %s\n", strings.Join(t.Tags, ", "))
			}
			if !t.Due.IsZero() {
				fmt.Fprintf(w, "due:      %s\n", t.Due.Format("2006-01-02"))
			}
			fmt.Fprintf(w, "created:  %s\n", t.CreatedAt.Format("2006-01-02 15:04"))
			fmt.Fprintf(w, "updated:  %s\n", t.UpdatedAt.Format("2006-01-02 15:04"))
			return nil
		},
	}
}
