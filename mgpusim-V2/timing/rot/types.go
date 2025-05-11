// mgpusim-V2/timing/rot/types.go
package rot

import (
	"github.com/sarchlab/akita/v3/sim"
)

// RotMessage is the base interface for all messages sent to/from the RoT
type RotMessage interface {
	sim.Msg
}

// RotRequest represents a request sent to the RoT
type RotRequest struct {
	sim.MsgMeta
	RequestType string
	Data        []byte
}

// Meta returns the meta data associated with the message
func (r *RotRequest) Meta() *sim.MsgMeta {
	return &r.MsgMeta
}

// RotResponse represents a response from the RoT
type RotResponse struct {
	sim.MsgMeta
	ResponseType string
	Data         []byte
	Status       uint32
	ErrorMessage string
}

// Meta returns the meta data associated with the message
func (r *RotResponse) Meta() *sim.MsgMeta {
	return &r.MsgMeta
}
