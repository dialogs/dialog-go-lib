package consumer

import (
	"context"
)

type groupItem struct {
	*Consumer
	closeCtx context.Context
}

func newGroupItem(c *Consumer, closeCtx context.Context) *groupItem {

	i := &groupItem{
		Consumer: c,
		closeCtx: closeCtx,
	}
	c.AddStateObserver(i)

	return i
}

func (i *groupItem) HandleStateEvent(e StateEvent) {
	if e == StateRun {
		go func() {
			<-i.closeCtx.Done()
			i.Stop()
		}()
	}
}
