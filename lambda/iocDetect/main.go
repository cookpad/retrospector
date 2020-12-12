package main

import (
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
	"github.com/m-mizutani/retrospector/pkg/service"
)

//Handler is exporeted for test
func Handler(args *arguments.Arguments, event golambda.Event) error {
	events, err := event.DecapSNSonSQSMessage()
	if err != nil {
		return err
	}

	repo := args.RepositoryService()
	alertSvc := args.AlertService()

	for _, event := range events {
		var iocChunk retrospector.IOCChunk
		if err := event.Bind(&iocChunk); err != nil {
			return golambda.WrapError(err).With("event", event)
		}

		for _, ioc := range iocChunk {
			entities, err := repo.GetEntities([]*retrospector.IOC{ioc})
			if err != nil {
				return err
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
				return golambda.WrapError(err).With("ioc", ioc).With("alert", alert)
			}
		}
	}

	return nil
}

func main() {
	golambda.Start(func(event golambda.Event) error {
		return Handler(arguments.New(), event)
	})
}
