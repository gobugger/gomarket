package captcha

import (
	"context"
	"github.com/alexedwards/scs/v2"
	"github.com/gobugger/gomarket/internal/util"
	"github.com/gobugger/gomarket/pkg/rand"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	SolutionKey    = "captcha_solution"
	cacheSize      = 100
	updateInterval = 5 * time.Minute
)

type item struct {
	solution Solution
	htmlSrc  string
}

func makeItem() item {
	captcha := New()

	src, err := util.Base64HTMLSrc(captcha.Image)
	if err != nil {
		slog.Error("failed to convert captcha image into base64", slog.Any("error", err))
	}

	return item{solution: captcha.Solution, htmlSrc: src}
}

var memstore = make([]item, cacheSize)
var mtx sync.Mutex
var ready atomic.Bool

func refreshCache(cache []item) {
	for i := range cache {
		cache[i] = makeItem()
	}
}

func init() {
	go func() {
		refreshCache(memstore)
		ready.Store(true)
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

func TemplFieldSrc(ctx context.Context, session *scs.SessionManager) string {
	i := func() item {
		if ready.Load() && mtx.TryLock() {
			defer mtx.Unlock()
			return memstore[rand.Intn(len(memstore))]
		} else {
			return makeItem()
		}
	}()

	session.Put(ctx, SolutionKey, i.solution)
	src, _ := strings.CutPrefix(string(i.htmlSrc), `src="`)
	src, _ = strings.CutSuffix(src, `"`)
	return src
}
