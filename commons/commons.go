package commons

import (
	"errors"
	"io/ioutil"
	"os"
)

func NewFile(path string, source string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = ioutil.WriteFile(path, []byte(source), 0644)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("The file already exist")
}
