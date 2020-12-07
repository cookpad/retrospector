package lambda

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/m-mizutani/retrospector/pkg/errors"
	"github.com/m-mizutani/retrospector/pkg/logging"
)

// Handler is callback function type of lambda.Run()
type Handler func(args *Arguments) error

// Run sets up Arguments and logging tools, then invoke handler with Arguments
func Run(handler Handler) {
	lambda.Start(func(ctx context.Context, event interface{}) error {
		defer errors.FlushSentry()
		logging.Logger.Info().Interface("event", event).Msg("Lambda start")

		args, err := newArguments(event)
		if err != nil {
			return err
		}

		if err := handler(args); err != nil {
			errors.EmitSentry(err)

			log := logging.Logger.Error()
			if e, ok := err.(*errors.Error); ok {
				for key, value := range e.Values {
					log = log.Interface(key, value)
				}
				log = log.Str("stacktrace", e.StackTrace())
			}

			log.Msg(err.Error())
			return err
		}
		return nil
	})
}
