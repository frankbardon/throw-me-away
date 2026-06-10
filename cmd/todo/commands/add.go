package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/frankbardon/todo/internal/task"
	"github.com/frankbardon/todo/pkg/dateparse"
	"github.com/frankbardon/todo/pkg/tags"
)

func newAddCmd() *cobra.Command {
	var (
		prio  string
		due   string
		tagsF string
	)
	cmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := joinArgs(args)

			p, err := task.ParsePriority(prio)
			if err != nil {
				return err
			}

			var dueT time.Time
			if due != "" {
				t, err := dateparse.Parse(due, time.Now())
				if err != nil {
					return err
				}
				dueT = t
			}

			tk, err := task.New(title, p, tags.Parse(tagsF), dueT)
			if err != nil {
				return err
			}

			store, err := openStore()
			if err != nil {
				return err
			}
			defer store.Close()

			created, err := store.Add(tk)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added #%d: %s\n", created.ID, created.Title)
			return nil
		},
	}
	cmd.Flags().StringVarP(&prio, "priority", "p", "medium", "priority: low|medium|high")
	cmd.Flags().StringVar(&due, "due", "", "due date (today, tomorrow, friday, in 3 days, 2026-12-31)")
	cmd.Flags().StringVarP(&tagsF, "tag", "t", "", "comma-separated tags")
	return cmd
}

func joinArgs(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += " "
		}
		out += a
	}
	return out
}
