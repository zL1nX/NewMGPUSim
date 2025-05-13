package rot

import (
	"github.com/sarchlab/akita/v3/sim"
	"github.com/sarchlab/akita/v3/tracing"
)

// Builder can build RoT components
type Builder struct {
	engine         sim.Engine
	freq           sim.Freq
	challengeWidth uint64
	responseWidth  uint64
	visTracer      tracing.Tracer
}

// WithEngine sets the engine to use
func (b *Builder) WithEngine(engine sim.Engine) *Builder {
	b.engine = engine
	return b
}

// WithFreq sets the frequency
func (b *Builder) WithFreq(freq sim.Freq) *Builder {
	b.freq = freq
	return b
}
