package recaster

import "fmt"

type OutputDriver interface {
	Execute(map[string]string) error
}

type DummpyOutput struct{}

func (do *DummpyOutput) Execute(templates map[string]string) error {
	fmt.Printf("Dummy driver invoked with %v\n", templates)
	return nil
}
