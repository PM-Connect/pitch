package scaffold

import (
	"os"
	"testing"

	filet "github.com/Flaque/filet"
	"github.com/stretchr/testify/assert"
)

func TestYamlFileLoader(t *testing.T) {
	defer filet.CleanUp(t)

	data := `
    user_input:
        some_var:
            description: Please enter some text.

    files:
        some/path/file.txt:
            mode: 0644
            conditions:
                - field: some_var
                  value: y
                  operator: equal
            template: |-
                A simple multi-line
                text file template.
    `

	filet.File(t, "scaffold.yaml", data)

	fileLoader := new(YamlFileLoader)

	scaffold, err := fileLoader.Get("scaffold.yaml")

	expectedTextFile := "A simple multi-line\ntext file template."

	var expectedPerms os.FileMode

	expectedPerms = 0644

	assert.Nil(t, err)
	assert.Equal(t, "Please enter some text.", scaffold.UserInput["some_var"].Description)
	assert.Equal(t, expectedTextFile, scaffold.Files["some/path/file.txt"].Template)
	assert.Equal(t, expectedPerms, scaffold.Files["some/path/file.txt"].Permissions)
	assert.Equal(t, "some_var", scaffold.Files["some/path/file.txt"].Conditions[0].Field)
	assert.Equal(t, "y", scaffold.Files["some/path/file.txt"].Conditions[0].Value)
	assert.Equal(t, "equal", scaffold.Files["some/path/file.txt"].Conditions[0].Operator)
}
