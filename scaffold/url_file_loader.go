package scaffold

import (
	"io/ioutil"
	"net/http"
	"time"

	yaml "gopkg.in/yaml.v2"
)

var urlClient = &http.Client{
	Timeout: time.Second * 10,
}

// URLFileLoader is responsible for returning Scaffolds from yaml files at a remote URL.
type URLFileLoader struct {
}

// Get returns a Scaffold from a given yaml file.
func (l *URLFileLoader) Get(source string) (scaffold Scaffold, err error) {
	response, err := urlClient.Get(source)

	if err != nil {
		return
	}

	buf, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return
	}

	err = yaml.Unmarshal(buf, &scaffold)

	return
}
