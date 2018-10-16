package scaffold

import "os"

// Scaffold is the configuration for a directory.
type Scaffold struct {
	UserInput map[string]Variable `yaml:"user_input" validate:"dive"`
	Files     map[string]File     `yaml:"files" validate:"required,dive"`
}

// Variable is the config for a variable of mixed types (multi-select, text, etc).
type Variable struct {
	Description string `yaml:"description" validate:"required"`
	Value       string `yaml:"value"`
}

// File is the structure of a file template.
type File struct {
	Permissions       os.FileMode  `yaml:"mode" validate:"required"`
	DirPermissions    os.FileMode  `yaml:"dir_mode"`
	Template          string       `yaml:"template" validate:"required"`
	Conditions        []Condition  `yaml:"conditions"`
	TemplateTags      TemplateTags `yaml:"template_tags"`
	DisableTemplating bool         `yaml:"disable_templating"`
}

// TemplateTags configures the templating syntax within a file.
type TemplateTags struct {
	Open  string `yaml:"open" validate:"omitempty,required"`
	Close string `yaml:"close" validate:"omitempty,required"`
}

// Condition is a condition structure to apply to files.
type Condition struct {
	Field    string `yaml:"field" validate:"required"`
	Value    string `yaml:"value" validate:"required"`
	Operator string `yaml:"operator" validate:"required,oneof=equal not_equal"`
}

// Loader is the interface for loading scaffolds.
type Loader interface {
	Get(source string) (Scaffold, error)
}

// Writer is the interface for writing scaffolds.
type Writer interface {
	Write(path string, file File) error
}
