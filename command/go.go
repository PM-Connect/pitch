package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pm-connect/pitch/scaffold"
	"github.com/pm-connect/pitch/utils"
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
Usage: pitch go <source> <directory>

    Go is used to run the bootstrap/scaffold process.

General Options:

	source:
		The source url/path to the yaml template.

	directory:
		The directory to scaffold/bootstrap into.

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

		if verbose {
			c.UI.Output(fmt.Sprintf("Using URL loader for source: %s", source))
		}
	} else {
		loader = c.FileLoader

		if verbose {
			c.UI.Output(fmt.Sprintf("Using File loader for source: %s", source))
		}
	}

	if verbose {
		c.UI.Output("Loading source.")
	}

	scf, err := loader.Get(source)

	if err != nil {
		c.UI.Error(fmt.Sprint(err))
		return 1
	}

	validate := *validator.New()

	if verbose {
		c.UI.Output("Validating source template configuration.")
	}

	err = validate.Struct(scf)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			c.UI.Error(fmt.Sprintf("Error validating source: \n    %s", err))
			return 1
		}
	}

	if verbose {
		c.UI.Output("Requesting variables from user.")
	}

	for name, variable := range scf.UserInput {
		var value string

		c.UI.Output(fmt.Sprintf("Fetching variable \"%s\"", name))

		for value == "" {
			value, _ = c.UI.Ask(variable.Description)

			if value == "" && variable.Value != "" {
				value = variable.Value
			} else if verbose {
				c.UI.Output("Value is empty and there is no default. Please enter a value.")
			}
		}

		variable.Value = value

		scf.UserInput[name] = variable
	}

	if verbose {
		c.UI.Output("Parsing file conditions and working out applicable files.")
	}

	filesToCreate := checkFiles(scf.Files, scf.UserInput)

	if verbose {
		c.UI.Output("Generating files.")
	}

	for name, file := range filesToCreate {
		name = utils.RemovePrefix(parseName(name, dir, scf.UserInput), "/")

		path := dir + name

		if _, err := os.Stat(path); !os.IsNotExist(err) && !overwrite {
			overwrite, _ := c.UI.Ask(fmt.Sprintf("File \"%s\" already exists. Overwrite? (y|n)", path))
			if strings.ToLower(overwrite) != "y" {
				continue
			}
		}

		if !file.DisableTemplating {
			file.Template = parseTemplate(file.Template, name, path, dir, file.TemplateTags, scf.UserInput)
		}

		if verbose {
			c.UI.Output(fmt.Sprintf("Writing file: %s", path))
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

func checkFiles(files map[string]scaffold.File, userData map[string]scaffold.Variable) map[string]scaffold.File {
	filesToCreate := make(map[string]scaffold.File)

	for name, file := range files {
		if len(file.Conditions) == 0 {
			filesToCreate[name] = file
			continue
		}

		passed := true

		for _, condition := range file.Conditions {
			switch condition.Operator {
			case "equal":
				if userData[condition.Field].Value != condition.Value {
					passed = false
				}
			case "not_equal":
				if userData[condition.Field].Value == condition.Value {
					passed = false
				}
			}
		}

		if passed {
			filesToCreate[name] = file
		}
	}

	return filesToCreate
}

func parseName(name string, dir string, userData map[string]scaffold.Variable) string {
	nameTemplate := fasttemplate.New(name, "%", "%")

	return nameTemplate.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		for name, variable := range userData {
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
	})
}

func parseTemplate(template string, name string, path string, dir string, tags scaffold.TemplateTags, userData map[string]scaffold.Variable) string {
	defaultTemplateTags := scaffold.TemplateTags{
		Open:  "%",
		Close: "%",
	}

	var templateTags scaffold.TemplateTags

	if tags.Open != "" && tags.Close != "" {
		templateTags = tags
	} else {
		templateTags = defaultTemplateTags
	}

	templater := fasttemplate.New(template, templateTags.Open, templateTags.Close)

	return templater.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		for name, variable := range userData {
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
