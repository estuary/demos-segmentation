package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/estuary/connectors/go-types/airbyte"
	"github.com/estuary/demos-segmentation/event-generator/events"
	"github.com/estuary/demos-segmentation/event-generator/throttle"
)

type Config struct {
	MaxEventsPerSecond int64  `json:"maxEventsPerSecond"`
	SegmentCardinality uint64 `json:"SegmentCardinality"`
	UserCardinality    uint64 `json:"userCardinality"`
}

func (c *Config) Validate() error {
	if c.MaxEventsPerSecond <= 0 {
		c.MaxEventsPerSecond = 1
	}
	if c.SegmentCardinality <= 0 {
		return fmt.Errorf("segmentCardinality must be greater than 0")
	}
	if c.UserCardinality <= 0 {
		return fmt.Errorf("userCardinality must be greater than 0")
	}
	return nil
}

const configSchema = `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"title":   "Segmentation Generator Source Spec",
	"type":    "object",
	"required": [
		"segmentCardinality",
		"userCardinality"
	],
	"properties": {
		"segmentCardinality": {
			"type":        "integer",
			"title":       "Number of Segments",
			"description": "Number of unique segments to use when generating events",
			"default":     "10000"
		},
		"userCardinality": {
			"type":        "integer",
			"title":       "Number of Users",
			"description": "Number of unique users to use when generating events",
			"default":     "1000000"
		}
		"maxEventsPerSecond": {
			"type":        "integer",
			"title":       "Number of Events per Second",
			"description": "Maximum number of Events produced per second",
			"default":     "10"
		}
	}
}`

type State struct {
	Cursor int `json:"cursor"`
}

func (c *State) Validate() error {
	return nil
}

func main() {
	airbyte.RunMain(spec, doCheck, doDiscover, doRead)
}

var spec = airbyte.Spec{
	SupportsIncremental:           true,
	SupportedDestinationSyncModes: airbyte.AllDestinationSyncModes,
	ConnectionSpecification:       json.RawMessage(configSchema),
}

func doCheck(args airbyte.CheckCmd) error {
	var result = &airbyte.ConnectionStatus{
		Status: airbyte.StatusSucceeded,
	}

	if err := args.ConfigFile.Parse(new(Config)); err != nil {
		result.Status = airbyte.StatusFailed
		result.Message = err.Error()
	}

	return airbyte.NewStdoutEncoder().Encode(airbyte.Message{
		Type:             airbyte.MessageTypeConnectionStatus,
		ConnectionStatus: result,
	})
}

func doDiscover(args airbyte.DiscoverCmd) error {
	if err := args.ConfigFile.Parse(new(Config)); err != nil {
		return err
	}

	var catalog = new(airbyte.Catalog)
	catalog.Streams = append(catalog.Streams, airbyte.Stream{
		Name:                    "segmentation-events",
		JSONSchema:              json.RawMessage(events.Schema),
		SupportedSyncModes:      airbyte.AllSyncModes,
		SourceDefinedCursor:     true,
		SourceDefinedPrimaryKey: [][]string{{"event"}},
	})

	var encoder = airbyte.NewStdoutEncoder()

	return encoder.Encode(airbyte.Message{
		Type:    airbyte.MessageTypeCatalog,
		Catalog: catalog,
	})
}

func doRead(args airbyte.ReadCmd) error {
	var config Config
	var state State
	var catalog airbyte.ConfiguredCatalog

	if err := args.ConfigFile.Parse(&config); err != nil {
		return err
	} else if err := args.CatalogFile.Parse(&catalog); err != nil {
		return err
	} else if args.StateFile != "" {
		if err := args.StateFile.Parse(&state); err != nil {
			return err
		}
	}

	var enc *json.Encoder = airbyte.NewStdoutEncoder()
	var produceEvent func(*State) error = buildEventProducer(enc, events.NewSource(config.SegmentCardinality, config.UserCardinality))
	var checkpoint func(*State) error = buildCheckpointer(enc)
	var throttler throttle.Throttler = throttle.PerSecond(config.MaxEventsPerSecond)

	for {
		if err := produceEvent(&state); err != nil {
			return err
		}

		if err := checkpoint(&state); err != nil {
			return err
		}

		if !catalog.Tail {
			return nil
		}

		throttler.WaitUntilReady()
	}
}

func buildEventProducer(enc *json.Encoder, eventSource events.Source) func(*State) error {
	var event events.Event

	return func(state *State) error {
		event = eventSource.Next()

		if err := writeEvent(enc, event); err != nil {
			return err
		}

		state.Cursor++

		return nil
	}
}

func buildCheckpointer(enc *json.Encoder) func(*State) error {
	var throttler throttle.Throttler = throttle.New(200 * time.Millisecond)

	return func(latestState *State) error {
		if throttler.IsReady() {
			return writeStateCheckpoint(enc, latestState)
		} else {
			return nil
		}
	}
}

func writeEvent(enc *json.Encoder, event events.Event) error {
	var jsonBody, err = json.Marshal(event)
	if err != nil {
		return err
	}

	if err = enc.Encode(&airbyte.Message{
		Type: airbyte.MessageTypeRecord,
		Record: &airbyte.Record{
			Stream:    "segmentation-events",
			EmittedAt: time.Now().UTC().UnixNano() / int64(time.Millisecond),
			Data:      jsonBody,
		},
	}); err != nil {
		return err
	}

	return nil
}

func writeStateCheckpoint(enc *json.Encoder, state *State) error {
	if jsonBody, err := json.Marshal(state); err != nil {
		return err
	} else if err = enc.Encode(airbyte.Message{
		Type:  airbyte.MessageTypeState,
		State: &airbyte.State{Data: jsonBody},
	}); err != nil {
		return err
	}

	return nil
}
