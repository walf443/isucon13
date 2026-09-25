package usecase

import (
	"fmt"
)

type fakeLogger struct {
	lines []string
	// warnLines は Warnf で出力された行
	warnLines []string
}

func (l *fakeLogger) Warnf(format string, args ...interface{}) {
	l.warnLines = append(l.warnLines, fmt.Sprintf(format, args...))
}

func (l *fakeLogger) Infof(format string, args ...interface{}) {
	l.lines = append(l.lines, fmt.Sprintf(format, args...))
}
