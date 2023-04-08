package submit

import (
	"context"

	em "github.com/digisan/event-mgr"
)

var (
	ctx    context.Context
	Cancel context.CancelFunc
)

func init() {
	ctx, Cancel = context.WithCancel(context.Background())
	em.InitDB("./data")
	em.InitEventSpan("TEN_MINUTE", ctx)
}
