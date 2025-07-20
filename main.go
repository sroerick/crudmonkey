package main

import (
	"crudmonkey/tui"

	sodaCmd "github.com/gobuffalo/pop/v6/soda/cmd"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use: "cm",
	}

	// wire up Pop/Soda as a subcommand
	sodaCmd.RootCmd.Use = "soda"
	root.AddCommand(sodaCmd.RootCmd)

	tuiCmd := &cobra.Command{
		Use:   "tui",
		Short: "Run the interactive TUI",
		Run: func(cmd *cobra.Command, args []string) {
			tui.Tui()
		},
	}
	root.AddCommand(tuiCmd)

	cobra.CheckErr(root.Execute())
}
