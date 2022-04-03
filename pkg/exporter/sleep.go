package exporter

import (
	"context"
	"time"
)

// SleepWithContext - реализация sleep для совместимости с контекстом приложения
func SleepWithContext(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
	case <-timer.C:
	}
}
