package main

import (
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
)

//Handler is exporeted for test
func Handler(args *arguments.Arguments, event golambda.Event) (interface{}, error) {
	events, err := event.DecapSNSonSQSMessage()
	if err != nil {
		return nil, err
	}

	repo := args.RepositoryService()

	for _, event := range events {
		var iocChunk retrospector.IOCChunk
		if err := event.Bind(&iocChunk); err != nil {
			return nil, err
		}

		if err := repo.PutIOCSet(iocChunk); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func main() {
	args := arguments.New()
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		return Handler(args, event)
	})
}
