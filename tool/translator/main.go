package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/text/message/pipeline"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

var localeDir = flag.String("ld", "./internal/translations/locales", "locales dir")
var chunkSize = flag.Int("cs", 30, "chunk size")

func main() {
	flag.Parse()

	t := newOpenaiTranslator()

	entries, err := os.ReadDir(*localeDir)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		locale := entry.Name()
		outf := filepath.Join(*localeDir, locale, "out.gotext.json")
		msgf := filepath.Join(*localeDir, locale, "messages.gotext.json")

		fmt.Printf("translating to %s from file %s\n", locale, outf)

		if err := translateFile(outf, msgf, t); err != nil {
			fmt.Printf("translation of %s failed: %s\n", outf, err)
			os.Exit(1)
		}

		fmt.Printf("translations written to %s\n", msgf)
	}
}

func translateFile(fin, fout string, t translator) error {
	content, err := os.ReadFile(fin)
	if err != nil {
		panic(err)
	}

	messages := pipeline.Messages{}

	if err := json.Unmarshal(content, &messages); err != nil {
		return err
	}

	chunks := slices.Collect(
		slices.Chunk(
			slices.DeleteFunc(
				slices.Clone(messages.Messages),
				func(msg pipeline.Message) bool {
					return !msg.Translation.IsEmpty()
				}),
			*chunkSize),
	)

	fmt.Printf("%d chunks to translate\n", len(chunks))

	c := make(chan *pipeline.Messages)

	go func() {
		wg := sync.WaitGroup{}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		for _, chunk := range chunks {
			if ctx.Err() != nil {
				break
			}

			promt, err := getPromt(&pipeline.Messages{
				Language: messages.Language,
				Messages: chunk,
				Macros:   messages.Macros,
			})
			if err != nil {
				fmt.Printf("failed to create promt: %s\n", err)
				break
			}

			wg.Add(1)
			go func() {
				msgs, err := t.translate(ctx, promt)
				if err != nil {
					fmt.Printf("failed to translate: %s\n", err)
					cancel()
				} else {
					c <- msgs
				}
				wg.Done()
			}()
		}

		wg.Wait()
		close(c)
	}()

	messageMap := map[string]pipeline.Message{}
	done := 0
	printProgress(0, len(chunks))
	for translated := range c {
		for _, msg := range translated.Messages {
			messageMap[msg.ID[0]] = msg
		}
		done++
		printProgress(done, len(chunks))
	}
	if done > 0 {
		fmt.Println()
	}

	// Update messages with translations
	for i, msg := range messages.Messages {
		if msg.Translation.IsEmpty() {
			if translated, ok := messageMap[msg.ID[0]]; ok && !translated.Translation.IsEmpty() {
				messages.Messages[i].Translation = translated.Translation
			}
		}
	}

	result, err := json.MarshalIndent(&messages, "  ", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(fout, result, 0666); err != nil {
		return err
	}

	return nil
}

func printProgress(done, total int) {
	fmt.Printf("\rprogress %05d/%d\t%.1f%%", done, total, 100*float64(done)/float64(total))
}
