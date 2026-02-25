package trace

import "errors"

type multiTraceSink struct {
	sinks []TraceSink
}

func NewMultiTraceSink(sinks ...TraceSink) TraceSink {
	filtered := make([]TraceSink, 0, len(sinks))
	for _, sink := range sinks {
		if sink == nil {
			continue
		}
		filtered = append(filtered, sink)
	}
	if len(filtered) == 0 {
		return NoopTraceSink()
	}
	if len(filtered) == 1 {
		return filtered[0]
	}
	return &multiTraceSink{sinks: filtered}
}

func (m *multiTraceSink) OnCommand(event TraceEvent) {
	for _, sink := range m.sinks {
		sink.OnCommand(event)
	}
}

func (m *multiTraceSink) Close() error {
	var errs []error
	for _, sink := range m.sinks {
		closable, ok := sink.(CloseableTraceSink)
		if !ok {
			continue
		}
		if err := closable.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
