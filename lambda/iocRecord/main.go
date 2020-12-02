package main

import (
	"github.com/m-mizutani/retrospector/pkg/lambda"
)

func handler(args *lambda.Arguments) error {
	return nil
}

func main() {
	lambda.Run(handler)
}
