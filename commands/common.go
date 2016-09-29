package commands

import (
	"encoding/json"
	"fmt"

	"github.com/Shopify/themekit/kit"
)

func drainErrors(errs chan error) {
	for {
		if err := <-errs; err != nil {
			kit.NotifyError(err)
		} else {
			break
		}
	}
}

func mergeEvents(dest chan kit.ThemeEvent, chans []chan kit.ThemeEvent) {
	go func() {
		for _, ch := range chans {
			var ok = true
			for ok {
				if ev, ok := <-ch; ok {
					dest <- ev
				}
			}
			close(ch)
		}
	}()
}

func logEvent(event kit.ThemeEvent, eventLog chan kit.ThemeEvent) {
	go func() {
		eventLog <- event
	}()
}

type basicEvent struct {
	Formatter func(b basicEvent) string
	EventType string `json:"event_type"`
	Target    string `json:"target"`
	Title     string `json:"title"`
	Etype     string `json:"type"`
}

func message(eventLog chan kit.ThemeEvent, content string, args ...interface{}) {
	logEvent(basicEvent{
		Formatter: func(b basicEvent) string { return fmt.Sprintf(content, args...) },
		EventType: "message",
		Title:     "Notice",
		Etype:     "basicEvent",
	}, eventLog)
}

func (b basicEvent) String() string {
	return b.Formatter(b)
}

func (b basicEvent) Successful() bool {
	return true
}

func (b basicEvent) Error() error {
	return nil
}

func (b basicEvent) AsJSON() ([]byte, error) {
	return json.Marshal(b)
}
