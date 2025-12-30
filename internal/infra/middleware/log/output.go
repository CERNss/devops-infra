package log

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"devops-infra/internal/constant"
	pathutil "devops-infra/internal/utils/path"
)

type OutputSink interface {
	Stdout() io.Writer
	Stderr() io.Writer
	StdoutPath() string
	StderrPath() string
	Close() error
}

type OutputSinkFactory interface {
	Open(info RuntimeInfo, command string) (OutputSink, error)
}

type RuntimeInfo struct {
	LogDir string
	TraceID string
}

type FileOutputSinkFactory struct {
	Dir string
}

func (f FileOutputSinkFactory) Open(info RuntimeInfo, command string) (OutputSink, error) {
	logDir := f.Dir
	if logDir == "" {
		logDir = info.LogDir
	}
	if logDir == "" {
		logDir = constant.DefaultOutputDir
	}

	resolved, err := pathutil.ResolveUserPath(logDir)
	if err != nil {
		return noopOutputSink{}, err
	}
	if err := os.MkdirAll(resolved, 0o755); err != nil {
		return noopOutputSink{}, err
	}

	identity := info.TraceID
	if identity == "" {
		identity = fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
	}
	stdoutPath := filepath.Join(resolved, "cmd-"+identity+".stdout.log")
	stderrPath := filepath.Join(resolved, "cmd-"+identity+".stderr.log")

	stdoutFile, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return noopOutputSink{}, err
	}

	stderrFile, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		_ = stdoutFile.Close()
		return noopOutputSink{}, err
	}

	return &FileOutputSink{
		stdoutPath: stdoutPath,
		stderrPath: stderrPath,
		stdoutFile: stdoutFile,
		stderrFile: stderrFile,
	}, nil
}

type FileOutputSink struct {
	stdoutPath string
	stderrPath string
	stdoutFile *os.File
	stderrFile *os.File
}

func (s *FileOutputSink) Stdout() io.Writer {
	return s.stdoutFile
}

func (s *FileOutputSink) Stderr() io.Writer {
	return s.stderrFile
}

func (s *FileOutputSink) StdoutPath() string {
	return s.stdoutPath
}

func (s *FileOutputSink) StderrPath() string {
	return s.stderrPath
}

func (s *FileOutputSink) Close() error {
	if s == nil {
		return nil
	}
	if s.stdoutFile != nil {
		_ = s.stdoutFile.Close()
	}
	if s.stderrFile != nil {
		_ = s.stderrFile.Close()
	}
	return nil
}

type noopOutputSink struct{}

func (noopOutputSink) Stdout() io.Writer    { return io.Discard }
func (noopOutputSink) Stderr() io.Writer    { return io.Discard }
func (noopOutputSink) StdoutPath() string   { return "" }
func (noopOutputSink) StderrPath() string   { return "" }
func (noopOutputSink) Close() error         { return nil }

func NoopOutputSink() OutputSink {
	return noopOutputSink{}
}

type CombinedOutputSinkFactory struct {
	Path string
}

func (f CombinedOutputSinkFactory) Open(info RuntimeInfo, command string) (OutputSink, error) {
	path := f.Path
	if path == "" {
		if info.LogDir != "" {
			path = filepath.Join(info.LogDir, "output.log")
		} else {
			path = constant.DefaultOutputFile
		}
	}

	resolved, err := pathutil.ResolveUserPath(path)
	if err != nil {
		return noopOutputSink{}, err
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return noopOutputSink{}, err
	}

	file, err := os.OpenFile(resolved, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return noopOutputSink{}, err
	}

	traceID := info.TraceID
	if traceID == "" {
		traceID = fmt.Sprintf("%d-%d", time.Now().UnixNano(), os.Getpid())
	}

	sharedMu := &sync.Mutex{}
	if err := writeCommandHeader(sharedMu, file, traceID, command); err != nil {
		_ = file.Close()
		return noopOutputSink{}, err
	}

	return &CombinedOutputSink{
		path:         resolved,
		file:         file,
		stdoutWriter: newPrefixedWriter(sharedMu, file, traceID, "stdout"),
		stderrWriter: newPrefixedWriter(sharedMu, file, traceID, "stderr"),
	}, nil
}

type CombinedOutputSink struct {
	path         string
	file         *os.File
	stdoutWriter io.Writer
	stderrWriter io.Writer
}

func (s *CombinedOutputSink) Stdout() io.Writer {
	return s.stdoutWriter
}

func (s *CombinedOutputSink) Stderr() io.Writer {
	return s.stderrWriter
}

func (s *CombinedOutputSink) StdoutPath() string {
	return s.path
}

func (s *CombinedOutputSink) StderrPath() string {
	return s.path
}

func (s *CombinedOutputSink) Close() error {
	if s == nil || s.file == nil {
		return nil
	}
	return s.file.Close()
}

type prefixedWriter struct {
	mu          *sync.Mutex
	w           io.Writer
	traceID     string
	stream      string
	atLineStart bool
}

func newPrefixedWriter(mu *sync.Mutex, w io.Writer, traceID string, stream string) *prefixedWriter {
	return &prefixedWriter{
		mu:          mu,
		w:           w,
		traceID:     traceID,
		stream:      stream,
		atLineStart: true,
	}
}

func (p *prefixedWriter) Write(b []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	consumed := 0
	for len(b) > 0 {
		if p.atLineStart {
			prefix := fmt.Sprintf("%s trace_id=%s stream=%s ", time.Now().Format(time.RFC3339Nano), p.traceID, p.stream)
			if _, err := io.WriteString(p.w, prefix); err != nil {
				return consumed, err
			}
			p.atLineStart = false
		}

		idx := bytes.IndexByte(b, '\n')
		if idx == -1 {
			if _, err := p.w.Write(b); err != nil {
				return consumed, err
			}
			consumed += len(b)
			return consumed, nil
		}

		chunk := b[:idx+1]
		if _, err := p.w.Write(chunk); err != nil {
			return consumed, err
		}
		consumed += len(chunk)
		b = b[idx+1:]
		p.atLineStart = true
	}

	return consumed, nil
}

func writeCommandHeader(mu *sync.Mutex, w io.Writer, traceID string, command string) error {
	mu.Lock()
	defer mu.Unlock()
	ts := time.Now().Format(time.RFC3339Nano)
	_, err := fmt.Fprintf(w, "%s trace_id=%s event=start cmd=%q\n", ts, traceID, command)
	return err
}
