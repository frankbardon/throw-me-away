package commands

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/frankbardon/todo/internal/filter"
	"github.com/frankbardon/todo/internal/task"
	"github.com/frankbardon/todo/pkg/dateparse"
)

func newListCmd() *cobra.Command {
	var (
		status    string
		prio      string
		tag       string
		text      string
		asJSON    bool
		dueBefore string
		overdue   bool
		sortKey   string
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
			if dueBefore != "" {
				t, err := dateparse.Parse(dueBefore, time.Now())
				if err != nil {
					return fmt.Errorf("parse --due-before: %w", err)
				}
				f.DueBefore = &t
			}
			if overdue {
				f.Overdue = true
			}
			if sortKey != "" && sortKey != "due" {
				return fmt.Errorf("unknown --sort value %q: only \"due\" is supported", sortKey)
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
			if sortKey == "due" {
				sort.SliceStable(out, func(i, j int) bool {
					a, b := out[i], out[j]
					if a.Due == nil && b.Due == nil {
						return a.ID < b.ID
					}
					if a.Due == nil {
						return false
					}
					if b.Due == nil {
						return true
					}
					if a.Due.Equal(*b.Due) {
						return a.ID < b.ID
					}
					return a.Due.Before(*b.Due)
				})
			}
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
	cmd.Flags().StringVar(&dueBefore, "due-before", "", "only tasks due before this date (e.g. tomorrow, 2026-06-10)")
	cmd.Flags().BoolVar(&overdue, "overdue", false, "only overdue, not-done tasks")
	cmd.Flags().StringVar(&sortKey, "sort", "", "sort order; supported: due")
	return cmd
}

func renderJSON(cmd *cobra.Command, tasks []*task.Task) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(tasks)
}

func renderTable(cmd *cobra.Command, tasks []*task.Task) {
	out := cmd.OutOrStdout()
	if len(tasks) == 0 {
		fmt.Fprintln(out, "no tasks")
		return
	}
	tty := isTerminal(out)
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tPRIO\tDUE\tTAGS\tTITLE")
	now := time.Now()
	for _, t := range tasks {
		due := "-"
		if t.Due != nil {
			due = t.Due.Format("2006-01-02")
			if t.Overdue(now) {
				if tty {
					due = ansiRed + due + ansiReset
				} else {
					due = "!" + due
				}
			}
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			t.ID, t.Status, t.Priority, due, strings.Join(t.Tags, ","), t.Title)
	}
	w.Flush()
}
