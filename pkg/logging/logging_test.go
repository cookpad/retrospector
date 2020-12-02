package logging_test

import (
	"github.com/m-mizutani/retrospector/pkg/logging"
)

func ExampleLogger() {
	logging.Logger.Info().Msg("hoge")
}
