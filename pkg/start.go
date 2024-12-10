package pkg

import (
	"log"
)

type Starter interface {
	Start() error
}

func SetUp(input ...Starter) error {
	for _, config := range input {
		err := config.Start()
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
