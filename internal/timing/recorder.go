package timing

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

// Recorder collects per-transaction latencies (delivery - origin) in memory
// and flushes them to a file on demand. Safe for concurrent use.
type Recorder struct {
	mu        sync.Mutex
	latencies []float64 // milliseconds
}

func NewRecorder() *Recorder {
	return &Recorder{}
}

// Record adds one observation. originTime is from MsgTransaction.OriginTime;
// deliveryTime is from QueueItem.DeliveryTime.
func (r *Recorder) Record(originTime, deliveryTime time.Time) {
	ms := float64(deliveryTime.Sub(originTime).Nanoseconds()) / 1e6
	r.mu.Lock()
	r.latencies = append(r.latencies, ms)
	r.mu.Unlock()
}

// Flush writes all recorded latencies (one per line, in milliseconds) to path.
// Intended to be called once at node shutdown.
func (r *Recorder) Flush(path string) error {
	r.mu.Lock()
	snapshot := make([]float64, len(r.latencies))
	copy(snapshot, r.latencies)
	r.mu.Unlock()

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("timing: create %s: %w", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, ms := range snapshot {
		fmt.Fprintf(w, "%.3f\n", ms)
	}
	return w.Flush()
}
