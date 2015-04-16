// +build ignore

package main

import (
	ui "github.com/gizak/termui"
	"github.com/matt3o12/termui-widgets/hackernews"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	panicIfErr(ui.Init())
	defer ui.Close()
	ui.UseTheme("helloworld")

	width := 100

	topStories := hackernews.NewWidget(hackernews.TopStories)
	topStories.Height = 15
	topStories.Width = width

	mostRecent := hackernews.NewWidget(hackernews.MostRecent)
	mostRecent.Height = 20
	mostRecent.Width = width
	mostRecent.Y = 15

	draw := func() {
		ui.Render(mostRecent, topStories)
	}

	topStories.UpdateEntries(draw)
	mostRecent.UpdateEntries(draw)
	draw()

	<-ui.EventCh()
}
