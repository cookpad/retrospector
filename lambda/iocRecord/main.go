package main

import (
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
)

//Handler is exporeted for test
func Handler(args *arguments.Arguments, event golambda.Event) error {
	events, err := event.DecapSNSonSQSMessage()
	if err != nil {
		return err
	}

	repo := args.RepositoryService()

	for _, event := range events {
		var iocChunk retrospector.IOCChunk
		if err := event.Bind(&iocChunk); err != nil {
			return err
		}

		if err := repo.PutIOCSet(iocChunk); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	args := arguments.New()
	golambda.Start(func(event golambda.Event) error {
		return Handler(args, event)
	})
}
