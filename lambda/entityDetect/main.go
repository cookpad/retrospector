package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/golambda"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/arguments"
	"github.com/m-mizutani/retrospector/pkg/service"
)

// Handler is exporeted for test
func Handler(args *arguments.Arguments, event golambda.Event) (interface{}, error) {
	recvEvents, err := event.DecapSNSonSQSMessage()
	if err != nil {
		return nil, err
	}

	alertSvc := args.AlertService()
	repoSvc := args.RepositoryService()
	entitySvc := args.EntityService()

	for _, event := range recvEvents {
		var s3Event events.S3Event
		if err := event.Bind(&s3Event); err != nil {
			return nil, err
		}

		for _, s3Record := range s3Event.Records {
			rq := entitySvc.NewReadQueue(s3Record.AWSRegion, s3Record.S3.Bucket.Name, s3Record.S3.Object.Key)
			entityMap := make(map[retrospector.Value][]*retrospector.Entity)

			for {
				entity := rq.Read()
				if entity == nil {
					break
				}

				entityMap[entity.Value] = append(entityMap[entity.Value], entity)
			}

			if err := rq.Error(); err != nil {
				return nil, golambda.WrapError(err).With("s3", s3Record)
			}

			for value, entities := range entityMap {
				matched, err := repoSvc.GetIOCSet([]*retrospector.Entity{
					{Value: value},
				})
				if err != nil {
					return nil, golambda.WrapError(err).With("s3", s3Record)
				}
				if len(matched) == 0 {
					continue
				}

				alert := &service.Alert{
					Cause:    service.AlertCauseEntity,
					Target:   &value,
					Entities: entities,
					IOCChunk: matched,
				}
				if err := alertSvc.EmitToSlack(alert); err != nil {
					return nil, golambda.WrapError(err).With("alert", alert).With("s3", s3Record)
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
