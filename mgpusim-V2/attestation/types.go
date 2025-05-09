// attestation/types.go
package attestation

type AttestationReport struct {
	GPUId       uint64
	Measurement []byte
	Signature   []byte
	Nonce       []byte
	Timestamp   uint64
}

type AttestationRequest struct {
	Nonce []byte
	GPUId uint64
}

type AttestationResponse struct {
	Report       AttestationReport
	Status       uint32
	ErrorMessage string
}
