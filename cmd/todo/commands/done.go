package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "done [id]",
		Short: "Mark a task as done",
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
			tk.MarkDone()
			if err := store.Update(tk); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "done #%d: %s\n", tk.ID, tk.Title)
			return nil
		},
	}
}
