package main

import (
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
	"github.com/m-mizutani/retrospector/pkg/service"
)

//Handler is exporeted for test
func Handler(args *arguments.Arguments, event golambda.Event) (interface{}, error) {
	events, err := event.DecapSNSonSQSMessage()
	if err != nil {
		return nil, err
	}

	repo := args.RepositoryService()
	alertSvc := args.AlertService()

	for _, event := range events {
		var iocChunk retrospector.IOCChunk
		if err := event.Bind(&iocChunk); err != nil {
			return nil, golambda.WrapError(err).With("event", event)
		}

		for _, ioc := range iocChunk {
			entities, err := repo.DetectEntities([]*retrospector.IOC{ioc})
			if err != nil {
				return nil, err
			}

			if len(entities) == 0 {
				continue
			}

			alert := &service.Alert{
				Cause:    service.AlertCauseIOC,
				Target:   &ioc.Value,
				Entities: entities,
				IOCChunk: retrospector.IOCChunk{ioc},
			}

			if err := alertSvc.EmitToSlack(alert); err != nil {
				return nil, golambda.WrapError(err).With("ioc", ioc).With("alert", alert)
			}

			for _, entity := range entities {
				if err := repo.UpdateEntityDetected(entity); err != nil {
					return nil, err
				}
			}
		}
	}

	return nil, nil
}

func main() {
	golambda.Start(func(event golambda.Event) (interface{}, error) {
		return Handler(arguments.New(), event)
	})
}
