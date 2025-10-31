package logger

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
)

var (
	endcollor = "\033[0m"
	magenta   = "\033[35m"
	blue      = "\033[34m"
	yellow    = "\033[33m"
	red       = "\033[31m"
	cyan      = "\033[36m"
	white     = "\033[97m"
)

type prettyHandler struct {
	slog.Handler
	l *log.Logger
}

func newHandler(w io.Writer, opts *slog.HandlerOptions) *prettyHandler {
	h := &prettyHandler{
		Handler: slog.NewJSONHandler(w, opts),
		l:       log.New(w, "", 0),
	}
	return h
}

func (h *prettyHandler) Handle(ctx context.Context, r slog.Record) error {
	paint := func(str string, color string) string {
		return color + str + endcollor
	}
	level := r.Level.String()
	switch r.Level {
	case slog.LevelDebug:
		level = paint(level, magenta)
	case slog.LevelInfo:
		level = paint(level, blue)
	case slog.LevelWarn:
		level = paint(level, yellow)
	case slog.LevelError:
		level = paint(level, red)
	}

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})

	b, err := json.MarshalIndent(fields, "", " ")
	if err != nil {
		return err
	}

	ts := r.Time.Format("[01-02-2006 15:04:05]")
	msg := paint(r.Message, cyan)

	h.l.Println(ts, level, msg, paint(string(b), white))

	return nil
}
