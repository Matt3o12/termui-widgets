package hackernews

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gizak/termui"
)

const timeFormat = "Jan 2, 2006 at 3:04pm (MST)"

const errorTolerance = 5

// Entry represents a simple hacker news entry. For now it only
// includes the default
type Entry struct {
	Title string
	Time  time.Time
	ID    int
}

func (e Entry) String() string {
	return fmt.Sprintf("\"%v\" on %v", e.Title, e.Time.Format(timeFormat))
}

// An EntryCache holds instances for all entries so they can be reused and
// don't need to be loaded a second time.
type EntryCache struct {
	entries    map[int]*Entry
	oldEntries map[int]*Entry
}

// Put an entry to the given id.
func (cache *EntryCache) Put(entry Entry) {
	if cache.entries == nil {
		cache.entries = make(map[int]*Entry)
	}

	cache.entries[entry.ID] = &entry
}

// Get returns an entry for the given id.
func (cache *EntryCache) Get(id int) *Entry {
	entry := cache.entries[id]
	if entry == nil && cache.oldEntries != nil {
		entry = cache.oldEntries[id]

		if entry != nil {
			cache.entries[id] = entry
		}
	}

	return entry
}

// GC removes all entries that had not been accessed since
// the last time this method has been called.
func (cache *EntryCache) GC() {
	cache.oldEntries = cache.entries
	cache.entries = make(map[int]*Entry)
}

type uiWidget interface {
	SetWidth(int)
	SetX(int)
	SetY(int)

	Buffer() []termui.Point
}

// All available Widget Types.
const (
	MostRecent WidgetType = iota
	TopStories
)

// WidgetType can either be `TopStories` or `MostRecent` stories.
type WidgetType int

func (t WidgetType) String() string {
	switch t {
	case MostRecent:
		return "Most Recent"

	case TopStories:
		return "Top Stories"
	}

	return "Unkown"
}

// Widget the hacker news widget used for showing current hackernews entries.
type Widget struct {
	termui.Block

	Type            WidgetType
	Ready           bool
	err             error
	ErrorCount      int
	EntryOrder      []int
	Entries         map[int]Entry
	RefreshInterval time.Duration

	Cache *EntryCache

	TextFgColor, TextBgColor termui.Attribute

	lastUpdatedDebug time.Time
}

// NewWidget creates a hacker-news new widget.
func NewWidget(widgetType WidgetType) *Widget {
	block := termui.NewBlock()
	return &Widget{
		Block:           *block,
		Ready:           false,
		Type:            widgetType,
		Cache:           &EntryCache{},
		RefreshInterval: 5 * time.Second,
		TextFgColor:     termui.Theme().ParTextFg,
		TextBgColor:     termui.Theme().ParTextBg,
		err:             nil,
	}
}

// EntriesMap returns a map containing all entries.
func (w *Widget) EntriesMap() map[int]Entry {
	entryMap := make(map[int]Entry)
	for _, entry := range w.Entries {
		entryMap[entry.ID] = entry
	}

	return entryMap
}

// Buffer draws the widget on the screen.
func (w Widget) Buffer() []termui.Point {
	var wrappedWidget uiWidget

	wrapperHeight := w.Height - 3
	if w.Error() != nil && !w.shouldIgnoreError() {
		textWidget := termui.NewPar(fmt.Sprintf("Error: %v", w.Error().Error()))
		textWidget.Height = wrapperHeight
		textWidget.HasBorder = false

		wrappedWidget = textWidget
	} else if !w.Ready {
		textWidget := termui.NewPar("Loading entries, please wait...")
		textWidget.Height = wrapperHeight
		textWidget.HasBorder = false

		wrappedWidget = textWidget
	} else {
		listWidget := termui.NewList()
		listWidget.Height = wrapperHeight
		listWidget.HasBorder = false

		items := make([]string, w.EntriesToDisplay())

		// addtional width: width that needs to be added so that they're
		// formatted properly.
		addWidth := int(math.Log10(float64(w.EntriesToDisplay()))) + 1
		for i := 0; i < w.EntriesToDisplay() && i < len(w.EntryOrder); i++ {
			entryID := w.EntryOrder[i]
			entry, hasEntry := w.Entries[entryID]

			var message string
			if hasEntry {
				message = entry.String()
			} else {
				message = "Loading, please wait"
			}

			dwidth := int(math.Log10(float64(i+1))) + 1
			width := strconv.Itoa(addWidth - dwidth)
			f := strings.Replace("[%v]%{width}v %v...", "{width}", width, 1)
			items[i] = fmt.Sprintf(f, i+1, "", message)
		}

		listWidget.Items = items
		wrappedWidget = listWidget
	}

	wrappedWidget.SetWidth(w.Width - 4)
	wrappedWidget.SetX(w.X + 2)
	wrappedWidget.SetY(w.Y + 1)

	w.Border.Label = fmt.Sprintf("Hacker News (%v)", w.Type.String())
	buffer := append(w.Block.Buffer(), wrappedWidget.Buffer()...)

	return buffer
}

func (w *Widget) loadEntries() ([]int, error) {
	switch w.Type {
	case MostRecent:
		return LoadMostRecentIDs()

	case TopStories:
		return LoadTopIDs()

	default:
		return nil, fmt.Errorf("Unkown Widget-Type: %v", w.Type)
	}
}

// EntriesToDisplay returns the number of items that should be displayed.
func (w *Widget) EntriesToDisplay() int {
	// 3 is the height that is lost because of the borders.
	return w.GetHeight() - 3
}

func fetchEntry(i, id int, e chan Entry, errs chan error) {
	if loadedEntry, err := LoadEntry(id); err != nil {
		errs <- err
	} else {
		e <- loadedEntry
	}
}

// resetWidget sets the error to nil, ready to false, sets the entry order to
// nothing and removes all entries.
func (w *Widget) resetWidget() {
	w.lastUpdatedDebug = time.Now()
	w.Entries = make(map[int]Entry)
	w.EntryOrder = nil
	w.Ready = false
	w.SetError(nil)
}

// shouldIgnoreError says whether or not a error should be ignored.
// Sometimes, the API limits the rates and shows us an error. However,
// if we already have enough data, we can still show the user the view.
// The w.ErrorCount may not exceed errorTolerance.
func (w *Widget) shouldIgnoreError() bool {
	return w.ErrorCount < errorTolerance && len(w.EntryOrder) > 0
}

// SetError sets last error for widget and increases the error count.
// If error is nil, the error count will be reset and the error
// of the widget will be set to nil.
func (w *Widget) SetError(err error) {
	if err == nil {
		w.err = nil
		w.ErrorCount = 0
	} else {
		w.err = err
		w.ErrorCount++
	}
}

// GetError returns the last error. Nil, if there wasn't an error.
func (w *Widget) Error() error {
	return w.err
}

func (w *Widget) updateEntries(readyCallback func()) {
	if entries, err := w.loadEntries(); err != nil {
		w.SetError(err)
	} else {
		w.resetWidget()
		w.EntryOrder = entries
		entryChan := make(chan Entry)
		errorChan := make(chan error)

		totalEntries := w.EntriesToDisplay()
		cacheHits := 0
		for i := 0; i < len(entries) && i < totalEntries; i++ {
			entryID := entries[i]

			if cachedEntry := w.Cache.Get(entryID); cachedEntry != nil {
				w.Entries[entryID] = *cachedEntry
				cacheHits++
				w.Ready = true
			} else {
				go fetchEntry(i, entryID, entryChan, errorChan)
			}
		}

		for i := cacheHits; i < len(entries) && i < totalEntries; i++ {
			select {
			case theEntry := <-entryChan:
				w.Entries[theEntry.ID] = theEntry
				w.Cache.Put(theEntry)
				w.Ready = true
			case entryErr := <-errorChan:
				w.SetError(entryErr)
			}

			readyCallback()
		}

		w.Ready = true
		w.Cache.GC()
	}

	readyCallback()
}

// UpdateEntries updates all entries and sets the status to ready. This function
// is non-blocking.
func (w *Widget) UpdateEntries(readyCallback func()) {
	go func() {
		for {
			w.updateEntries(readyCallback)
			time.Sleep(w.RefreshInterval)
		}
	}()
}
