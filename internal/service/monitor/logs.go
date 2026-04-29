package monitor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/fsnotify/fsnotify"
)

// logFollower tails a file and emits each line on out.
// Closes out when its context is cancelled.
type logFollower struct {
	path string
	file *os.File
	out  chan<- string
	w    *fsnotify.Watcher
}

func newLogFollower(path string, out chan<- string) (*logFollower, error) {
	if path == "" {
		return nil, ErrLogPathEmpty
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open log: %w", err)
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("fsnotify: %w", err)
	}
	if err := w.Add(path); err != nil {
		_ = f.Close()
		_ = w.Close()
		return nil, fmt.Errorf("fsnotify add: %w", err)
	}
	return &logFollower{path: path, file: f, out: out, w: w}, nil
}

// run blocks until ctx is done. Emits each newline-terminated chunk on out.
func (l *logFollower) run(ctx context.Context) {
	defer func() {
		_ = l.file.Close()
		_ = l.w.Close()
	}()
	rd := bufio.NewReader(l.file)
	emit := func() {
		for {
			line, err := rd.ReadString('\n')
			if len(line) > 0 {
				// strip trailing \n
				if line[len(line)-1] == '\n' {
					line = line[:len(line)-1]
				}
				select {
				case l.out <- line:
				case <-ctx.Done():
					return
				}
			}
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				return
			}
		}
	}
	emit() // initial flush of pre-existing content
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-l.w.Events:
			if !ok {
				return
			}
			if ev.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				emit()
			}
		case _, ok := <-l.w.Errors:
			if !ok {
				return
			}
			// transient watch error; keep going.
		}
	}
}
