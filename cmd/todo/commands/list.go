package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/frankbardon/todo/internal/filter"
	"github.com/frankbardon/todo/internal/task"
)

func newListCmd() *cobra.Command {
	var (
		status   string
		prio     string
		tag      string
		text     string
		asJSON   bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			f := filter.Filter{Tag: tag, TextContain: text}
			if status != "" {
				s, err := task.ParseStatus(status)
				if err != nil {
					return err
				}
				f.Status = &s
			}
			if prio != "" {
				p, err := task.ParsePriority(prio)
				if err != nil {
					return err
				}
				f.Priority = &p
			}

			store, err := openStore()
			if err != nil {
				return err
			}
			defer store.Close()

			all, err := store.List()
			if err != nil {
				return err
			}
			out := f.Apply(all)
			if asJSON {
				return renderJSON(cmd, out)
			}
			renderTable(cmd, out)
			return nil
		},
	}
	cmd.Flags().StringVarP(&status, "status", "s", "", "filter by status")
	cmd.Flags().StringVarP(&prio, "priority", "p", "", "filter by priority")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "filter by tag")
	cmd.Flags().StringVar(&text, "text", "", "match substring in title")
	cmd.Flags().BoolVar(&asJSON, "json", false, "emit machine-readable JSON")
	return cmd
}

func renderJSON(cmd *cobra.Command, tasks []*task.Task) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(tasks)
}

func renderTable(cmd *cobra.Command, tasks []*task.Task) {
	if len(tasks) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no tasks")
		return
	}
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tPRIO\tDUE\tTAGS\tTITLE")
	now := time.Now()
	for _, t := range tasks {
		due := "-"
		if !t.Due.IsZero() {
			due = t.Due.Format("2006-01-02")
			if t.Overdue(now) {
				due += "*"
			}
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			t.ID, t.Status, t.Priority, due, strings.Join(t.Tags, ","), t.Title)
	}
	w.Flush()
}
