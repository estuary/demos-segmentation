package main

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	// Universe of vendors, mapped to via modulo from each segment.
	vendorCardinality = 10
	// Zipfian skew of the generated segment distribution. Must be > 1.
	// Smaller values sample more universally from the universe of segments,
	// while larger values increase the relative frequency of some segments
	// over others. Try values in range 1.0001 to 1.5.
	segmentSkew = 1.1
	// userStdDevClip is the multiple of the standard deviation by which we
	// scale and clip the distribution. Smaller values sample more universally
	// from the universe of users, while larger values increase the relative
	// frequency of some users over others. Try values in range 5.0 to 20.0.
	userStdDevClip = 10.0
	// Each event is "add" or "remove" sampled with uniform probability density.
	// Adds are more frequent than removes.
	addProbability = 0.7
	// Each event is novel or a repetition of previous event, sampled uniformly.
	// Novel events are more frequent, but there are many repeats (e.x. browser reload).
	repeatProability = 0.4
	// Size of the reservoir from which we select repeated events.
	repetitionReservoirSize = 20000
)

type Segment struct {
	Vendor int    `json:"vendor"`
	Name   string `json:"name"`
}

type Event struct {
	EventID   string  `json:"event"`
	Timestamp string  `json:"timestamp"`
	User      string  `json:"user"`
	Segment   Segment `json:"segment"`
	Remove    bool    `json:"remove,omitempty"`
}

const eventSchema = `
{
  "type": "object",
  "properties": {
    "event": {"type": "string"},
    "timestamp": {"type": "string"},
    "user": {"type": "string"},
    "segment": {
      "type": "object",
      "properties": {
        "vendor": {"type": "number"},
        "name": {"type": "string"}
      }
    },
    "remove": {"type": ["boolean", null]}
  },
  "required": ["event", "timestamp", "user", "segment"]
}`

type sample struct {
	segment, user int
	add           bool
}

type eventSource struct {
	rnd             *rand.Rand
	rndSegment      *rand.Zipf
	reservoir       []sample
	userCardinality uint64
	// We'll generate timestamps seeded from the present time,
	// but uniformly incremented by 10ms with each generated event.
	now    time.Time
	tickCh *time.Ticker
}

func newEventSource(segmentCardinality uint64, userCardinality uint64) eventSource {
	var rnd = rand.New(rand.NewSource(8675309))

	return eventSource{
		rnd:             rnd,
		rndSegment:      rand.NewZipf(rnd, segmentSkew, 1, segmentCardinality),
		reservoir:       make([]sample, 0, repetitionReservoirSize),
		userCardinality: userCardinality,
		now:             time.Now(),
		tickCh:          time.NewTicker(time.Second),
	}
}

func (gen *eventSource) next() Event {
	var cur sample

	if len(gen.reservoir) == 0 || gen.rnd.Float32() > repeatProability {
		// Generate a novel sample.
		cur = sample{
			// Sample segment ∈ [0, SegmentCardinality].
			segment: int(gen.rndSegment.Uint64()),
			// Sample from positive half of the normal distribution, then scaled and clip to p ∈ [0, 1].
			// Then, project to user ∈ [0, UserCardinality].
			user: int(float64(gen.userCardinality) * math.Min(1.0, math.Abs(gen.rnd.NormFloat64())/userStdDevClip)),
			// Sample uniformly ∈ [0, 1] and project to "add" or "remove".
			add: gen.rnd.Float32() < addProbability,
		}
	} else {
		// Sample from reservoir.
		cur = gen.reservoir[gen.rnd.Intn(len(gen.reservoir))]
	}

	// Update sample reservoir.
	if len(gen.reservoir) == cap(gen.reservoir) {
		gen.reservoir[gen.rnd.Intn(len(gen.reservoir))] = cur
	} else {
		gen.reservoir = append(gen.reservoir, cur)
	}

	// Read a random UUID.
	var uuid [16]byte
	gen.rnd.Read(uuid[:])

	// Maybe update current event time.
	select {
	case gen.now = <-gen.tickCh.C:
	default:
	}

	var event = Event{
		EventID:   encodeHexUUID(uuid),
		Timestamp: gen.now.Format(time.RFC3339),
		User:      fmt.Sprintf("usr-%06x", cur.user),
		Remove:    !cur.add,
	}

	event.Segment.Vendor = 1 + (cur.segment % vendorCardinality)
	event.Segment.Name = fmt.Sprintf("seg-%X", cur.segment)

	return event
}

func encodeHexUUID(uuid [16]byte) string {
	var buf [36]byte

	hex.Encode(buf[0:8], uuid[:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], uuid[10:])

	return string(buf[:])
}
