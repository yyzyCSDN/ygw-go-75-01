package alarm

import (
	"io"
	"sync"

	"aquarecirc/internal/model"
)

type BufferSink struct {
	mu     sync.Mutex
	alarms []model.Alarm
	limit  int
}

func NewBufferSink(limit int) *BufferSink {
	return &BufferSink{limit: limit}
}

func (s *BufferSink) Emit(evt model.Alarm) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alarms = append(s.alarms, evt)
	if len(s.alarms) > s.limit {
		s.alarms = s.alarms[len(s.alarms)-s.limit:]
	}
}

func (s *BufferSink) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.alarms)
}

type ConsoleSink struct {
	writer io.Writer
}

func NewConsoleSink(writer io.Writer) *ConsoleSink {
	return &ConsoleSink{writer: writer}
}

func (s *ConsoleSink) Emit(evt model.Alarm) {
	if s.writer == nil {
		return
	}
	_, _ = io.WriteString(
		s.writer,
		evt.Timestamp.Format("15:04:05")+" alarm "+evt.PondID+" "+evt.Message+"\n",
	)
}
