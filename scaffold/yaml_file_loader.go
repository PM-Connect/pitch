package scaffold

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// YamlFileLoader is responsible for returning Scaffolds from yaml files.
type YamlFileLoader struct {
}

// Get returns a Scaffold from a given yaml file.
func (l *YamlFileLoader) Get(source string) (scaffold Scaffold, err error) {
	if _, err = os.Stat(source); os.IsNotExist(err) {
		return
	}

	data, err := ioutil.ReadFile(source)

	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &scaffold)

	return
}
