package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/frankbardon/todo/internal/task"
)

func newClearDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear-done",
		Short: "Delete all tasks that are marked done",
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
			removed := 0
			for _, t := range all {
				if t.Status != task.StatusDone {
					continue
				}
				if err := store.Delete(t.ID); err != nil {
					return err
				}
				removed++
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed %d done task(s)\n", removed)
			return nil
		},
	}
}
