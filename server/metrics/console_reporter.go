package metrics

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/g-brook/brook/common/log"
)

type ConsoleReporter struct {
	registry *Registry
	interval time.Duration
	stopCh   chan struct{}
	once     sync.Once
}

func NewConsoleReporter(registry *Registry, interval time.Duration) *ConsoleReporter {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &ConsoleReporter{
		registry: registry,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (r *ConsoleReporter) Start() {
	if r == nil || r.registry == nil {
		return
	}
	go r.loop()
}

func (r *ConsoleReporter) Stop() {
	if r == nil {
		return
	}
	r.once.Do(func() {
		close(r.stopCh)
	})
}

func (r *ConsoleReporter) loop() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.report()
		case <-r.stopCh:
			return
		}
	}
}

func (r *ConsoleReporter) report() {
	snapshots := r.registry.Snapshot()
	if len(snapshots) == 0 {
		return
	}

	sort.Slice(snapshots, func(i, j int) bool {
		if snapshots[i].Port == snapshots[j].Port {
			return snapshots[i].ID < snapshots[j].ID
		}
		return snapshots[i].Port < snapshots[j].Port
	})

	log.Info("[monitor]\n%s", formatConsoleTable(snapshots))
}

func formatConsoleTable(snapshots []TunnelTrafficSnapshot) string {
	if len(snapshots) == 0 {
		return ""
	}

	lines := make([]string, 0, len(snapshots)+2)
	lines = append(lines,
		"ID                TYPE  PORT   CONN CLIENT IN        OUT       IN_RATE   OUT_RATE  LATENCY  RUNTIME LAST_SEEN",
		"----------------  ----  -----  ---- ------ --------  --------  --------  --------  -------  ------- ---------",
	)
	for _, snap := range snapshots {
		lines = append(lines, formatConsoleRow(snap))
	}
	return joinLines(lines)
}

func formatConsoleRow(snap TunnelTrafficSnapshot) string {
	runtime := "-"
	if !snap.Runtime.IsZero() {
		runtime = shortDuration(time.Since(snap.Runtime))
	}
	lastSeen := "-"
	if !snap.LastSeen.IsZero() {
		lastSeen = shortDuration(time.Since(snap.LastSeen))
	}
	return fmt.Sprintf(
		"%-16s  %-4s  %5d  %4d %6d %-8s  %-8s  %-8s  %-8s  %-7s  %-7s %-9s",
		truncate(snap.ID, 16),
		truncate(snap.Type, 4),
		snap.Port,
		snap.Connections,
		snap.Clients,
		humanBytes(snap.InBytes),
		humanBytes(snap.OutBytes),
		humanRate(snap.InRateBps),
		humanRate(snap.OutRateBps),
		formatLatency(snap.LatencyMs),
		truncate(runtime, 7),
		truncate(lastSeen, 9),
	)
}

func formatLatency(ms float64) string {
	if ms <= 0 {
		return "-"
	}
	if ms < 1000 {
		return fmt.Sprintf("%.0fms", ms)
	}
	return fmt.Sprintf("%.2fs", ms/1000)
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

func shortDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func truncate(v string, size int) string {
	if len(v) <= size {
		return v
	}
	if size <= 1 {
		return v[:size]
	}
	return v[:size-1] + "~"
}

func humanRate(v float64) string {
	return humanBytes(uint64(v))
}

func humanBytes(v uint64) string {
	const unit = 1024
	if v < unit {
		return fmt.Sprintf("%dB", v)
	}
	div, exp := uint64(unit), 0
	for n := v / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(v)/float64(div), "KMGTPE"[exp])
}
