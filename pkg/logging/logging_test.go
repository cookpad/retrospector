package logging_test

import (
	"github.com/cookpad/retrospector/pkg/logging"
)

func ExampleLogger() {
	logging.Logger.Info().Msg("hoge")
}
