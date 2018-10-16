package scaffold

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
)

func TestURLFileLoader(t *testing.T) {
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

	http.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write([]byte(data))
	})

	port, err := freeport.GetFreePort()

	assert.Nil(t, err)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

		if err != nil {
			panic(err)
		}
	}()

	loader := new(URLFileLoader)

	scaffold, err := loader.Get(fmt.Sprintf("http://localhost:%d/file", port))

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
