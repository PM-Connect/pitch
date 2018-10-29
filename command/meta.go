package command

import (
	"github.com/mitchellh/cli"
	input "github.com/tcnksm/go-input"
)

// Meta contains the meta options for functionally for neraly every command.
type Meta struct {
	UI    cli.Ui
	Input *input.UI
}
