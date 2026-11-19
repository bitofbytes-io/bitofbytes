package status

import "time"

const (
	// uptimeWindow is the rolling period the uptime percentage covers.
	uptimeWindow = 24 * time.Hour
	// minUptimeSamples is how many polls in the window are needed before a
	// percentage is shown; fewer would let one failed poll read as 0%.
	minUptimeSamples = 10
	// historyCapacity holds one sample per poll over the window, plus an hour
	// of headroom. Polls start every pollInterval but samples are stamped when
	// a check finishes, so completions can fall closer together than that, and
	// a full ring must never drop a sample that is still inside the window.
	historyCapacity = int(uptimeWindow/pollInterval) + 60
)

type sample struct {
	at time.Time
	up bool
}

// ring is a fixed-capacity buffer of poll samples. Once full, a new sample
// overwrites the oldest. It is not safe for concurrent use; Checker guards it.
type ring struct {
	samples []sample
	next    int // index the next sample is written to
	full    bool
}

func newRing(capacity int) ring { return ring{samples: make([]sample, capacity)} }

func (r *ring) add(at time.Time, up bool) {
	if len(r.samples) == 0 {
		return
	}
	r.samples[r.next] = sample{at: at, up: up}
	r.next++
	if r.next == len(r.samples) {
		r.next, r.full = 0, true
	}
}

// count returns how many samples taken within window up to now were up and how
// many there were in total. Older samples and ones dated after now are ignored.
func (r *ring) count(now time.Time, window time.Duration) (up, total int) {
	n := r.next
	if r.full {
		n = len(r.samples)
	}
	cutoff := now.Add(-window)
	for _, s := range r.samples[:n] {
		if s.at.After(cutoff) && !s.at.After(now) {
			total++
			if s.up {
				up++
			}
		}
	}
	return up, total
}

// uptimePermille is up/total as an integer number of tenths of a percent,
// rounded down so a target with any failure never reads as 100%. ok is false
// until there are minUptimeSamples samples.
func uptimePermille(up, total int) (permille int, ok bool) {
	if total < minUptimeSamples {
		return 0, false
	}
	return up * 1000 / total, true
}
