package commands

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete [id]",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete a task",
		Args:    cobra.ExactArgs(1),
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

			if err := store.Delete(id); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted #%d\n", id)
			return nil
		},
	}
}
