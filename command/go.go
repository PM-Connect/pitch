package command

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
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

	source: string
		The source url/path to the yaml template.

	directory: string
		The directory to scaffold/bootstrap into.

	-overwrite:
		Automatically overwrite existing files without prompting.

	-dry-run:
		Run without any mutations or file creation.

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
	var verbose, overwrite, dryrun bool

	flags := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	flags.BoolVar(&verbose, "verbose", false, "Turn on verbose output.")
	flags.BoolVar(&overwrite, "overwrite", false, "Automatically overwrite existing files.")
	flags.BoolVar(&dryrun, "dry-run", false, "Run without writing/creating any files.")
	flags.Parse(args)

	args = flags.Args()

	var source, dir string

	if len(args) > 0 {
		source = args[0]
	}

	if len(args) > 1 {
		dir = args[1]
	} else {
		userDirPrompt := promptui.Prompt{
			Label: "Please enter a directory name (use ./ for current path):",
			Validate: func(input string) error {
				if _, err := os.Stat(input); os.IsNotExist(err) {
					return fmt.Errorf("Unable to find directory: %s", input)
				}

				return nil
			},
		}
		userDir, err := userDirPrompt.Run()

		if err != nil {
			c.UI.Error(fmt.Sprintf("%s", err))
			return 1
		}

		dir = userDir
	}

	dir = utils.EnsureSuffix(dir, "/")

	if len(source) == 0 {
		c.UI.Error("A source must be provided as the first argument.")
		return 1
	}

	var loader scaffold.Loader

	if utils.IsValidURL(source) {
		loader = c.URLLoader

		c.verboseOutput(fmt.Sprintf("Using URL loader for source: %s", source), verbose)
	} else {
		loader = c.FileLoader

		c.verboseOutput(fmt.Sprintf("Using File loader for source: %s", source), verbose)
	}

	c.verboseOutput("Loading source.", verbose)

	scf, err := loader.Get(source)

	if err != nil {
		c.UI.Error(fmt.Sprint(err))
		return 1
	}

	validate := *validator.New()

	c.verboseOutput("Validating source template configuration.", verbose)

	err = validate.Struct(scf)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			c.UI.Error(fmt.Sprintf("Error validating source: \n    %s", err))
			return 1
		}
	}

	c.verboseOutput("Requesting variables from user.", verbose)

	for name, variable := range scf.UserInput {
		var value string

		if verbose {
			c.UI.Output(fmt.Sprintf("Fetching variable \"%s\"", name))
		}

		var err error

		if len(variable.Options) > 0 {
			variableValueSelect := promptui.Select{
				Label: variable.Description,
				Items: variable.Options,
			}
			_, value, err = variableValueSelect.Run()
		} else {
			variableValuePrompt := promptui.Prompt{
				Label: variable.Description,
			}
			value, err = variableValuePrompt.Run()
		}

		if err != nil {
			if len(variable.Value) == 0 {
				c.UI.Error(fmt.Sprintf("%s", err))
				return 1
			}
		}

		variable.Value = value

		scf.UserInput[name] = variable
	}

	c.verboseOutput("Parsing file conditions and working out applicable files.", verbose)

	filesToCreate := checkFiles(scf.Files, scf.UserInput)

	c.verboseOutput("Generating files.", verbose)

	for name, file := range filesToCreate {
		name = utils.RemovePrefix(parseName(name, dir, scf.UserInput), "/")

		path := dir + name

		if _, err := os.Stat(path); !os.IsNotExist(err) && !overwrite {
			overwriteConfirm := promptui.Prompt{
				Label:     fmt.Sprintf("File \"%s\" already exists. Overwrite? (y|n)", path),
				IsConfirm: true,
			}
			overwrite, _ := overwriteConfirm.Run()

			if strings.ToLower(overwrite) != "y" {
				continue
			}

			c.UI.Output(fmt.Sprintf("Overwriting file: %s", path))
		}

		if !file.DisableTemplating {
			file.Template = parseTemplate(file.Template, name, path, dir, file.TemplateTags, scf.UserInput)
		}

		c.verboseOutput(fmt.Sprintf("Writing file: %s", path), verbose)

		if !dryrun {
			err := c.Writer.Write(path, file)

			if err != nil {
				c.UI.Error(fmt.Sprintf("Error writing file \"%s\": %s", path, err))
			} else {
				c.UI.Info(fmt.Sprintf("Created file \"%s\".", path))
			}
		}
	}

	return 0
}

func (c *GoCommand) verboseOutput(output string, verbose bool) {
	if verbose {
		c.UI.Output(output)
	}
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
