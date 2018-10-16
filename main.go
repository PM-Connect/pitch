package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/PM-Connect/pitch/command"
	"github.com/mitchellh/cli"
)

func main() {
	os.Exit(Run(os.Args[1:]))
}

// Run the cli application.
func Run(args []string) int {
	return RunCustom(args)
}

// RunCustom runs the custom setup cli app.
func RunCustom(args []string) int {
	commands := command.Commands()

	cli := &cli.CLI{
		Name:                       "pitch",
		Version:                    "1.0.0",
		Args:                       args,
		Commands:                   commands,
		Autocomplete:               true,
		AutocompleteNoDefaultFlags: true,
		HelpFunc:                   groupedHelpFunc(cli.BasicHelpFunc("pitch")),
		HelpWriter:                 os.Stdout,
	}

	exitCode, err := cli.Run()

	if err != nil {
		log.Printf("Error executing CLI: %s", err.Error())
		return 1
	}

	return exitCode
}

func groupedHelpFunc(f cli.HelpFunc) cli.HelpFunc {
	return func(commands map[string]cli.CommandFactory) string {
		var b bytes.Buffer

		tw := tabwriter.NewWriter(&b, 0, 2, 6, ' ', 0)

		fmt.Fprintf(tw, "Usage: pitch [-version] [-help] [-verbose] [-autocomplete-(un)install] <command> [args]\n\n")
		fmt.Fprintf(tw, "Common commands:\n")
		for k := range commands {
			printCommand(tw, k, commands[k])
		}

		tw.Flush()

		return strings.TrimSpace(b.String())
	}
}

func printCommand(w io.Writer, name string, cmdFn cli.CommandFactory) {
	cmd, err := cmdFn()

	if err != nil {
		log.Panicf("failed to load %q command: %s", name, err)
	}
	fmt.Fprintf(w, "    %s\t%s\n", name, cmd.Synopsis())
}
