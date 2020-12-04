package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/m-mizutani/retrospector"
	"github.com/m-mizutani/retrospector/pkg/errors"
	"github.com/m-mizutani/retrospector/pkg/lambda"
)

// Handler is exporeted for test
func Handler(args *lambda.Arguments) error {
	recvEvents, err := args.DecapSNSoverSQSEvent()
	if err != nil {
		return err
	}

	repoSvc := args.RepositoryService()
	entitySvc := args.EntityService()

	for _, event := range recvEvents {
		var s3Event events.S3Event
		if err := event.Bind(&s3Event); err != nil {
			return err
		}

		for _, s3Record := range s3Event.Records {
			rq := entitySvc.NewReadQueue(s3Record.AWSRegion, s3Record.S3.Bucket.Name, s3Record.S3.Object.Key)
			var entities []*retrospector.Entity
			for {
				entity := rq.Read()
				if entity != nil {
					break
				}
				entities = append(entities, entity)
			}

			if err := rq.Error(); err != nil {
				return errors.With(err, "s3", s3Record)
			}

			if err := repoSvc.PutEntities(entities); err != nil {
				return errors.With(err, "s3", s3Record)
			}
		}
	}

	return nil
}

func main() {
	lambda.Run(Handler)
}
