package main

import (
	"flag"
	"fmt"
	"golang.org/x/text/language"
	"golang.org/x/text/message/pipeline"
	"log"
	"strings"
)

var (
	lang      = flag.String("lang", "en-US", "comma separated list of languages to translate to")
	packages  = flag.String("packages", "", "comma separated list of packages to translate")
	out       = flag.String("out", "", "out file")
	overwrite = flag.Bool("overwrite", false, "overwrite")
	srcLang   = flag.String("srclang", "en-US", "the source-code language")
	dir       = flag.String("dir", "locales", "default subdirectory to store translation files")
)

func config() (*pipeline.Config, error) {
	tag, err := language.Parse(*srcLang)
	if err != nil {
		return nil, fmt.Errorf("invalid srclang: %w", err)
	}

	// Use a default value since rewrite and extract don't have an out flag.
	genFile := ""
	if out != nil {
		genFile = *out
	}

	return &pipeline.Config{
		SourceLanguage:      tag,
		Supported:           getLangs(),
		TranslationsPattern: `messages\.(.*)\.json$`,
		GenFile:             genFile,
		Dir:                 *dir,
	}, nil
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	args := flag.Args()

	packages := args[1:]

	cfg, err := config()
	if err != nil {
		panic(wrap(err, "failed to parse config"))
	}

	if err := runUpdate(cfg, packages); err != nil {
		panic(wrap(err, "failed"))
	}
}

func runUpdate(config *pipeline.Config, args []string) error {
	config.Packages = args
	state, err := pipeline.Extract(config)
	if err != nil {
		return wrap(err, "extract failed")
	}
	if err := state.Import(); err != nil {
		return wrap(err, "import failed")
	}
	if err := state.Merge(); err != nil {
		return wrap(err, "merge failed")
	}

	mergePositions(state)

	if err := state.Export(); err != nil {
		return wrap(err, "export failed")
	}
	if *out != "" {
		return wrap(state.Generate(), "generation failed")
	}
	return nil
}

func mergePositions(state *pipeline.State) {
	positions := map[string]string{}

	for _, message := range state.Extracted.Messages {
		positions[message.ID[0]] = message.Position
	}

	for i, messages := range state.Messages {
		for j, message := range messages.Messages {
			if p, ok := positions[message.ID[0]]; ok {
				state.Messages[i].Messages[j].Position = p
			}
		}
	}
}

func getLangs() (tags []language.Tag) {
	if lang == nil {
		return []language.Tag{language.AmericanEnglish}
	}
	for t := range strings.SplitSeq(*lang, ",") {
		if t == "" {
			continue
		}
		tag, err := language.Parse(t)
		if err != nil {
			panic(wrap(err, "failed to parse language"))
		}
		tags = append(tags, tag)
	}
	return tags
}

func wrap(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}
