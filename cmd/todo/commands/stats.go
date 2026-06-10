package commands

import (
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/frankbardon/todo/internal/task"
)

func newStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Print a summary of task counts",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := openStore()
			if err != nil {
				return err
			}
			defer store.Close()

			all, err := store.List()
			if err != nil {
				return err
			}

			byStatus := map[task.Status]int{}
			byPriority := map[task.Priority]int{}
			overdue := 0
			now := time.Now()
			for _, t := range all {
				byStatus[t.Status]++
				byPriority[t.Priority]++
				if t.Overdue(now) {
					overdue++
				}
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintf(w, "total\t%d\n", len(all))
			fmt.Fprintf(w, "overdue\t%d\n", overdue)
			fmt.Fprintln(w, "---\t---")
			for _, s := range []task.Status{task.StatusTodo, task.StatusDoing, task.StatusDone} {
				fmt.Fprintf(w, "%s\t%d\n", s, byStatus[s])
			}
			fmt.Fprintln(w, "---\t---")
			for _, p := range []task.Priority{task.PriorityLow, task.PriorityMedium, task.PriorityHigh} {
				fmt.Fprintf(w, "%s\t%d\n", p, byPriority[p])
			}
			return w.Flush()
		},
	}
}
