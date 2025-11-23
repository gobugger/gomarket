package main

import (
	"context"
	"golang.org/x/text/message/pipeline"
)

type translator interface {
	translate(ctx context.Context, promt string) (*pipeline.Messages, error)
}
