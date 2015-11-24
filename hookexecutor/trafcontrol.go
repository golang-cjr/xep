package hookexecutor

import (
	"log"
	"strings"
	"time"
)

type rateLimitCheck int

const (
	RLOK rateLimitCheck = iota
	RLPartialOverflow
	RLOverflow
)

type ShaperConfig struct {
	RatePerMinute int
	RatePer10sec  int
}

type TrafficEvent struct {
	timestamp time.Time
	score     int
}

type Rejecter func(*Message, string)

type TrafficController struct {
	config ShaperConfig

	input  chan *Message
	output chan *Message
	reject Rejecter

	logger *log.Logger
}

func NewTrafficController(config ShaperConfig, reject Rejecter, input, output chan *Message, logger *log.Logger) *TrafficController {
	return &TrafficController{config, input, output, reject, logger}
}

func (tc *TrafficController) Start() {
	go tc.run()
}

func (tc *TrafficController) run() {
	defer func() {
		if err := recover(); err != nil {
			tc.logger.Printf("catched panic in traffic controller loop: %v", err)
		}
	}()

	tsLog := []*TrafficEvent{}

	for {
		msg, ok := <-tc.input
		if !ok {
			return
		}

		event := &TrafficEvent{score: countLines(msg), timestamp: time.Now()}

		status := checkRateLimit(tsLog, event, tc.config.RatePerMinute, time.Minute)
		if status != RLOK {
			// enforce hard global limit
			tc.reject(msg, "ratelimited")
			continue
		}

		status = checkRateLimit(tsLog, event, tc.config.RatePer10sec, 10*time.Second)
		if status == RLOverflow {
			// allow burst overflow for single message
			tc.reject(msg, "burst_ratelimited")
			continue
		}

		select {
		case tc.output <- msg:
		default:
			tc.reject(msg, "busy")
			continue
		}

		tsLog = addToLog(tsLog, event, time.Minute)
	}
}

func checkRateLimit(tsLog []*TrafficEvent, event *TrafficEvent, limit int, period time.Duration) rateLimitCheck {
	target := time.Now().Add(-period)
	total := 0
	for i := len(tsLog) - 1; i > 0; i-- {
		item := tsLog[i]
		if target.After(tsLog[i].timestamp) {
			break
		}

		total += item.score
		if total > limit {
			return RLOverflow
		}
	}

	if total+event.score < limit {
		return RLOK
	} else {
		return RLPartialOverflow
	}
}

func addToLog(tsLog []*TrafficEvent, event *TrafficEvent, period time.Duration) []*TrafficEvent {
	tsLog = append(tsLog, event)

	first := time.Now().Add(-period)
	for idx, event := range tsLog {
		if first.After(event.timestamp) {
			return tsLog[idx:]
		}
	}

	return tsLog[:0]
}

func countLines(msg *Message) int {
	return len(strings.Split(msg.Data["body"], "\n"))
}
