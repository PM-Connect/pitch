package command

import (
	"os"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/pm-connect/pitch/scaffold"
	input "github.com/tcnksm/go-input"
)

func generalOptionsUsage() string {
	helpText := `
    -verbose
        Enables verbose logging.
    `

	return strings.TrimSpace(helpText)
}

// Commands creates all of the possible commands that can be run.
func Commands() map[string]cli.CommandFactory {
	meta := Meta{}

	meta.UI = &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	meta.UI = &cli.ColoredUi{
		Ui:         meta.UI,
		ErrorColor: cli.UiColorRed,
		WarnColor:  cli.UiColorYellow,
		InfoColor:  cli.UiColorGreen,
	}

	meta.Input = &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	return map[string]cli.CommandFactory{
		"go": func() (cli.Command, error) {
			return &GoCommand{
				Meta:       meta,
				FileLoader: new(scaffold.YamlFileLoader),
				URLLoader:  new(scaffold.URLFileLoader),
				Writer:     new(scaffold.IoWriter),
			}, nil
		},
	}
}
