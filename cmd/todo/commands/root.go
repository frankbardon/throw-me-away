package commands

import (
	"github.com/spf13/cobra"

	"github.com/frankbardon/todo/internal/config"
	"github.com/frankbardon/todo/internal/storage"
)

var (
	flagConfigPath string
)

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "todo",
		Short:         "A terminal task tracker",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&flagConfigPath, "config", "", "path to tasks JSON file (defaults to XDG data dir)")

	root.AddCommand(
		newAddCmd(),
		newListCmd(),
		newDoneCmd(),
		newDeleteCmd(),
		newShowCmd(),
		newEditCmd(),
		newClearDoneCmd(),
		newStatsCmd(),
	)
	return root
}

func Execute() error {
	return newRoot().Execute()
}

func openStore() (storage.Store, error) {
	path := flagConfigPath
	if path == "" {
		c, err := config.Default()
		if err != nil {
			return nil, err
		}
		path = c.StorePath
	}
	return storage.NewJSONStore(path)
}
