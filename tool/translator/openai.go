package main

import (
	"context"
	"encoding/json"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
	"golang.org/x/text/message/pipeline"
)

type openaiTranslator openai.Client

func newOpenaiTranslator() openaiTranslator {
	return openaiTranslator(openai.NewClient())
}

func (t openaiTranslator) translate(ctx context.Context, promt string) (*pipeline.Messages, error) {
	resp, err := t.Responses.New(ctx, responses.ResponseNewParams{
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(promt),
		},
		Model:       responses.ChatModelGPT4oMini,
		Temperature: openai.Float(0.2),
	})
	if err != nil {
		return nil, err
	}

	res := &pipeline.Messages{}
	if err := json.Unmarshal([]byte(resp.OutputText()), res); err != nil {
		return res, err
	}

	return res, nil
}
