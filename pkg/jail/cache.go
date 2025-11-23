package jail

import (
	"github.com/gobugger/gomarket/internal/config"
	"github.com/gobugger/gomarket/internal/util"
	"github.com/gobugger/gomarket/pkg/rand"
	"image"
	"log/slog"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

const (
	cacheSize      = 100
	updateInterval = 5 * time.Minute
)

var backgrounds []*image.NRGBA

type item struct {
	indexes []int
	src     string
}

func makeItem() item {
	indexes := RandomIndexes()
	img := GetImage(indexes)

	html, err := util.Base64HTMLSrc(img)
	if err != nil {
		slog.Error("base64 encoding failed", slog.Any("error", err))
	}

	return item{indexes: indexes, src: html}
}

var memstore = make([]item, cacheSize)
var mtx sync.Mutex
var ready atomic.Bool

func refreshCache(cache []item) {
	for i := range cache {
		cache[i] = makeItem()
	}
}

func Setup() {
	pattern := filepath.Join(config.StaticDir, "entry_backgrounds/*.webp")
	bgs, err := loadBackgrounds(pattern, 280, 140)
	if err != nil {
		slog.Error("failed to load backgrounds", slog.Any("error", err))
	}
	backgrounds = bgs
}

func Start() {
	refreshCache(memstore)
	ready.Store(true)
	go func() {
		im := make([]item, cacheSize)
		for {
			time.Sleep(updateInterval)
			refreshCache(im)
			mtx.Lock()
			copy(memstore, im)
			mtx.Unlock()
		}
	}()
}

func SetupAndStart() {
	Setup()
	Start()
}

func Get() ([]int, string) {
	item := func() item {
		if ready.Load() && mtx.TryLock() {
			defer mtx.Unlock()
			return memstore[rand.Intn(len(memstore))]
		} else {
			return makeItem()
		}
	}()
	return item.indexes, item.src
}
