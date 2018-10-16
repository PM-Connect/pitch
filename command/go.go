package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/PM-Connect/pitch/scaffold"
	"github.com/PM-Connect/pitch/utils"
	"github.com/valyala/fasttemplate"
	validator "gopkg.in/go-playground/validator.v9"
)

// GoCommand runs the build to prepare the project for deployment.
type GoCommand struct {
	Meta
	FileLoader scaffold.Loader
	URLLoader  scaffold.Loader
	Writer     scaffold.Writer
}

// Help displays help output for the command.
func (c *GoCommand) Help() string {
	helpText := `
Usage: pitch go

    Go is used to run the bootstrap/scaffold process.

General Options:

    ` + generalOptionsUsage() + `
    `

	return strings.TrimSpace(helpText)
}

// Synopsis displays the command synopsis.
func (c *GoCommand) Synopsis() string { return "Build the project according to the config." }

// Name returns the name of the command.
func (c *GoCommand) Name() string { return "build" }

// Run starts the build procedure.
func (c *GoCommand) Run(args []string) int {
	var verbose, overwrite bool

	flags := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	flags.BoolVar(&verbose, "verbose", false, "Turn on verbose output.")
	flags.BoolVar(&overwrite, "overwrite", false, "Automatically overwrite existing files.")
	flags.Parse(args)

	args = flags.Args()

	var source, dir string

	if len(args) > 0 {
		source = args[0]
	}

	if len(args) > 1 {
		dir = args[1]
	}

	if len(source) == 0 {
		c.UI.Error("A source must be provided as the first argument.")
		return 1
	}

	if len(dir) == 0 {
		userDir, err := c.UI.Ask("Please enter a directory name (use ./ for current path):")

		if err != nil {
			c.UI.Error(fmt.Sprintf("%s", err))
			return 1
		}

		dir = userDir
	}

	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	var loader scaffold.Loader

	if utils.IsValidURL(source) {
		loader = c.URLLoader
	} else {
		loader = c.FileLoader
	}

	scf, err := loader.Get(source)

	if err != nil {
		c.UI.Error(fmt.Sprint(err))
		return 1
	}

	validate := *validator.New()

	err = validate.Struct(scf)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			c.UI.Error(fmt.Sprintf("Error validating source: \n    %s", err))
			return 1
		}
	}

	for name, variable := range scf.UserInput {
		var value string

		for value == "" {
			value, _ = c.UI.Ask(variable.Description)

			if value == "" && variable.Value != "" {
				value = variable.Value
			}
		}

		variable.Value = value

		scf.UserInput[name] = variable
	}

	filesToCreate := make(map[string]scaffold.File)

	for name, file := range scf.Files {
		if len(file.Conditions) == 0 {
			filesToCreate[name] = file
			continue
		}

		passed := true

		for _, condition := range file.Conditions {
			switch condition.Operator {
			case "equal":
				if scf.UserInput[condition.Field].Value != condition.Value {
					passed = false
				}
			case "not_equal":
				if scf.UserInput[condition.Field].Value == condition.Value {
					passed = false
				}
			}
		}

		if passed {
			filesToCreate[name] = file
		}
	}

	for name, file := range filesToCreate {
		nameTemplate := fasttemplate.New(name, "%", "%")

		name = nameTemplate.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
			for name, variable := range scf.UserInput {
				if name == tag {
					return w.Write([]byte(variable.Value))
				}
			}

			switch tag {
			case "dir":
				return w.Write([]byte(dir))
			default:
				return w.Write([]byte(tag))
			}

			return w.Write([]byte(tag))
		})

		if strings.HasPrefix(name, "/") {
			name = strings.TrimPrefix(name, "/")
		}

		path := dir + name

		if _, err := os.Stat(path); !os.IsNotExist(err) && !overwrite {
			overwrite, _ := c.UI.Ask(fmt.Sprintf("File \"%s\" already exists. Overwrite? (y|n)", path))
			if strings.ToLower(overwrite) != "y" {
				continue
			}
		}

		if !file.DisableTemplating {
			defaultTemplateTags := scaffold.TemplateTags{
				Open:  "%",
				Close: "%",
			}

			var templateTags scaffold.TemplateTags

			if file.TemplateTags.Open != "" && file.TemplateTags.Close != "" {
				templateTags = file.TemplateTags
			} else {
				templateTags = defaultTemplateTags
			}

			template := fasttemplate.New(file.Template, templateTags.Open, templateTags.Close)

			file.Template = template.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
				for name, variable := range scf.UserInput {
					if name == tag {
						return w.Write([]byte(variable.Value))
					}
				}

				switch tag {
				case "current_file":
					return w.Write([]byte(name))
				case "current_file_path":
					return w.Write([]byte(path))
				case "dir":
					return w.Write([]byte(dir))
				default:
					return w.Write([]byte(tag))
				}
			})
		}

		err := c.Writer.Write(path, file)

		if err != nil {
			c.UI.Error(fmt.Sprintf("Error writing file \"%s\": %s", path, err))
		} else {
			c.UI.Info(fmt.Sprintf("Created file \"%s\".", path))
		}
	}

	return 0
}
