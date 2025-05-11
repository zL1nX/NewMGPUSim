package puf

import (
	"crypto/sha256"
	"fmt"

	"github.com/sarchlab/akita/v3/sim"
)

// Port represents a communication endpoint for PUF requests
type Port interface {
	sim.Port
}

// PUFChallenge represents a request to generate a PUF response
type PUFChallenge struct {
	sim.MsgMeta
	Challenge []byte
}

// Meta returns the meta data associated with the message
func (r *PUFChallenge) Meta() *sim.MsgMeta {
	return &r.MsgMeta
}

// PUFResponse represents the response from the PUF
type PUFResponse struct {
	sim.MsgMeta
	Response []byte
}

// Meta returns the meta data associated with the message
func (r *PUFResponse) Meta() *sim.MsgMeta {
	return &r.MsgMeta
}

// SendPUFChallengeEvent represents an event to send a challenge to PUF
type SendPUFChallengeEvent struct {
	*sim.EventBase
	ChallengeMsg *PUFChallenge
	Src          sim.Port
	Dst          sim.Port
}

// PUFResponseEvent represents a response event from PUF
type PUFResponseEvent struct {
	*sim.EventBase
	ResponseMsg *PUFResponse
	Src         sim.Port
	Dst         sim.Port
}

// PUF Component represents a PUF hardware component
type PUF struct {
	*sim.TickingComponent
	challengeWidth uint64 // in bytes
	responseWidth  uint64 // in bytes
	ToCP           sim.Port
	responseDelay  int
	randID         []byte // nonce for unique PUF instances
}

// NewPUF creates a new PUF component
func NewPUF(
	name string,
	engine sim.Engine,
	freq sim.Freq,
	challengeWidth uint64,
	responseWidth uint64,
	randID []byte,
) *PUF {
	newPUF := new(PUF)
	newPUF.TickingComponent = sim.NewTickingComponent(name, engine, freq, newPUF)
	newPUF.challengeWidth = challengeWidth
	newPUF.responseWidth = responseWidth
	newPUF.responseDelay = 10
	newPUF.randID = randID

	newPUF.ToCP = sim.NewLimitNumMsgPort(newPUF, 1, name+".ToCP")
	newPUF.AddPort("ToCP", newPUF.ToCP)

	fmt.Printf("[*] PUF Instance Created: %s with randID: %x\n",
		newPUF.Name(), randID)
	return newPUF
}

func (p *PUF) Tick(now sim.VTimeInSec) bool {
	msg := p.ToCP.Peek()
	if msg == nil {
		return false
	}

	switch req := msg.(type) {
	case *PUFChallenge:
		fmt.Printf("[*] PUF Challenge Received in PUF from CP port (%s): %x\n", p.ToCP.Name(), req.Challenge)

		response := p.GenerateResponse(req.Challenge)
		rspMsg := &PUFResponse{
			MsgMeta: sim.MsgMeta{
				ID:       sim.GetIDGenerator().Generate(),
				SendTime: now + sim.VTimeInSec(float64(p.responseDelay)/float64(p.Freq)),
				Src:      p.ToCP,
				Dst:      req.Meta().Src,
			},
			Response: response,
		}

		p.ToCP.Send(rspMsg)
		p.ToCP.Retrieve(now)
		fmt.Printf("[*] PUF Response Sent from PUF port (%s) to CP port (%s): %x\n", p.ToCP.Name(), req.Src.Name(), response)
		return true
	}

	return false
}

// GetPortByName returns the named port
func (p *PUF) GetPortByName(name string) sim.Port {
	fmt.Println("In PUF GetPortByName", p.Freq)
	switch name {
	case "ToCP":
		return p.ToCP
	default:
		return nil
	}
}

// GenerateResponse generates a PUF response using the nonce
func (p *PUF) GenerateResponse(challenge []byte) []byte {
	// Create a hash instance
	h := sha256.New()

	// Write challenge and randID directly to hash
	h.Write(challenge)
	h.Write(p.randID)

	// Get hash and truncate to desired response width
	hash := h.Sum(nil)
	response := hash[:p.responseWidth]

	return response
}
