package protocol

import (
	"github.com/sarchlab/akita/v3/sim"
	"github.com/sarchlab/mgpusim/v3/attestation"
)

type GPUAttestationReq struct {
	sim.MsgMeta
	Nonce []byte
}

type GPUAttestationRsp struct {
	sim.MsgMeta
	Report attestation.AttestationReport
}

// Meta returns the meta data associated with the message
func (r *GPUAttestationReq) Meta() *sim.MsgMeta {
	return &r.MsgMeta
}

func (r *GPUAttestationRsp) Meta() *sim.MsgMeta {
	return &r.MsgMeta
}
