package tracker

import (
	"encoding/json"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DailyCounts maps "YYYY-MM-DD" → unique visitor count for that day.
type DailyCounts map[string]int

type Tracker struct {
	mu       sync.Mutex
	counts   DailyCounts
	seen     map[string]struct{} // in-memory dedup: "YYYY-MM-DD|ip"
	filePath string
}

// New creates a Tracker, ensuring the data directory exists and loading
// any previously persisted counts. Starts fresh on missing or malformed file.
func New(filePath string) *Tracker {
	t := &Tracker{
		counts:   make(DailyCounts),
		seen:     make(map[string]struct{}),
		filePath: filePath,
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Info("tracker: could not create data dir", "err", err)
	}

	if err := t.load(); err != nil {
		slog.Info("tracker: could not load existing data, starting fresh", "err", err)
		t.counts = make(DailyCounts)
	} else {
		slog.Info("tracker: loaded data", "days", len(t.counts), "file", t.filePath)
	}

	return t
}

// RecordVisit records a visit from ip for today. If this IP has already been
// seen today (in memory), it is a no-op. Persists to disk on a new unique visit.
func (t *Tracker) RecordVisit(ip string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	date := today()
	key := date + "|" + ip

	if _, already := t.seen[key]; already {
		return
	}

	t.seen[key] = struct{}{}
	t.counts[date]++

	if err := t.save(); err != nil {
		slog.Info("tracker: failed to persist counts", "err", err)
	}
}

// Counts returns a safe copy of the current daily counts map.
func (t *Tracker) Counts() DailyCounts {
	t.mu.Lock()
	defer t.mu.Unlock()

	copy := make(DailyCounts, len(t.counts))
	for k, v := range t.counts {
		copy[k] = v
	}
	return copy
}

// ExtractIP returns the best-effort real client IP.
// Checks X-Forwarded-For first (set by Caddy), falls back to remoteAddr.
func ExtractIP(xff, remoteAddr string) string {
	if xff != "" {
		// X-Forwarded-For may be "client, proxy1, proxy2" — take the first field
		parts := strings.SplitN(xff, ",", 2)
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

func (t *Tracker) load() error {
	data, err := os.ReadFile(t.filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &t.counts)
}

// save atomically writes counts to disk using write-to-temp + rename.
func (t *Tracker) save() error {
	data, err := json.Marshal(t.counts)
	if err != nil {
		return err
	}

	dir := filepath.Dir(t.filePath)
	tmp, err := os.CreateTemp(dir, "user_counts_*.json.tmp")
	if err != nil {
		return err
	}

	if _, err = tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err = tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	return os.Rename(tmp.Name(), t.filePath)
}

func today() string {
	return time.Now().UTC().Format("2006-01-02")
}
