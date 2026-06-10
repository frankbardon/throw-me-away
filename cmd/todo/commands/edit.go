package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/frankbardon/todo/internal/task"
	"github.com/frankbardon/todo/pkg/dateparse"
	"github.com/frankbardon/todo/pkg/tags"
)

func newEditCmd() *cobra.Command {
	var (
		title string
		prio  string
		due   string
		tagsF string
		clear bool
	)
	cmd := &cobra.Command{
		Use:   "edit [id]",
		Short: "Edit an existing task",
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

			tk, err := store.Get(id)
			if err != nil {
				return err
			}

			if title != "" {
				tk.Title = title
			}
			if prio != "" {
				p, err := task.ParsePriority(prio)
				if err != nil {
					return err
				}
				tk.Priority = p
			}
			if clear {
				tk.Due = time.Time{}
			} else if due != "" {
				d, err := dateparse.Parse(due, time.Now())
				if err != nil {
					return err
				}
				tk.Due = d
			}
			if tagsF != "" {
				tk.Tags = tags.Parse(tagsF)
			}

			if err := store.Update(tk); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "edited #%d: %s\n", tk.ID, tk.Title)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVarP(&prio, "priority", "p", "", "new priority")
	cmd.Flags().StringVar(&due, "due", "", "new due date")
	cmd.Flags().StringVarP(&tagsF, "tag", "t", "", "replace tags (comma-separated)")
	cmd.Flags().BoolVar(&clear, "clear-due", false, "clear the due date")
	return cmd
}
