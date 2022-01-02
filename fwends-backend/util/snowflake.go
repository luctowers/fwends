package util

import (
	"errors"
	"sync/atomic"
	"time"
)

// see https://en.wikipedia.org/wiki/Snowflake_ID

// TODO: make these configurable
const _ = 41 // timestampBits
const machineBits = 10
const sequenceBits = 12

type SnowflakeGenerator struct {
	sequence int64
	base     int64
}

func NewSnowflakeGenerator(machine int64) (*SnowflakeGenerator, error) {
	g := new(SnowflakeGenerator)
	g.sequence = 0
	if machine >= (1 << machineBits) {
		return nil, errors.New("snowflake machine id out of range")
	}
	g.base = machine << sequenceBits
	return g, nil
}

func (g *SnowflakeGenerator) GenID() int64 {
	id := g.base
	seq := atomic.AddInt64(&g.sequence, 1) - 1
	id |= seq & ((1 << (sequenceBits + 1)) - 1)
	id |= time.Now().UnixMilli() << sequenceBits << machineBits
	return id
}
