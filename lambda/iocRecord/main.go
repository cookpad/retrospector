package main

import (
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/lambda"
)

func handler(args *lambda.Arguments) error {
	events, err := args.DecapSNSoverSQSEvent()
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
	lambda.Run(handler)
}
