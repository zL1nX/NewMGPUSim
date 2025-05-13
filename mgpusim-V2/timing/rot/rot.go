package rot

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math"

	"github.com/sarchlab/akita/v3/sim"
	"github.com/sarchlab/mgpusim/v3/attestation"
	"github.com/sarchlab/mgpusim/v3/protocol"
	"github.com/sarchlab/mgpusim/v3/puf"
)

// RoT represents the Root of Trust hardware component
type RoT struct {
	*sim.TickingComponent

	gpuID         uint64
	privateKey    *ecdsa.PrivateKey
	publicKey     *ecdsa.PublicKey
	pufRandID     []byte // nonce for unique PUF instances
	responseDelay int

	// PUF parameters
	challengeWidth uint64 // in bytes
	responseWidth  uint64 // in bytes

	// PUF component
	pufComponent *puf.PUF

	PUF   sim.Port
	cpRoT sim.Port

	// Communication ports
	ToCP  sim.Port
	ToPUF sim.Port

	toPUFSender sim.BufferedSender
}

// generateKeypair creates a new ECDSA keypair for the RoT
func generateKeypair() (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	// In a real implementation, this would use a hardware-backed key
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return key, &key.PublicKey
}

// GetPublicKey returns the RoT's public key
func (r *RoT) GetPublicKey() *ecdsa.PublicKey {
	return r.publicKey
}

func (r *RoT) SetPUF(puf *puf.PUF) {
	r.pufComponent = puf
}

// NewRoT creates a new Root of Trust module
func NewRoT(
	name string,
	engine sim.Engine,
	freq sim.Freq,
	gpuID uint64,
	challengeWidth uint64,
	responseWidth uint64,
) *RoT {
	rot := &RoT{
		gpuID:          gpuID,
		challengeWidth: challengeWidth,
		responseWidth:  responseWidth,
		responseDelay:  10,
	}
	rot.TickingComponent = sim.NewTickingComponent(name, engine, freq, rot)

	// Generate private/public key pair for the RoT
	rot.privateKey, rot.publicKey = generateKeypair()

	// Generate a random value as the PUF identifier
	rot.pufRandID = make([]byte, 16)
	rand.Read(rot.pufRandID)

	// Initialize standalone PUF component
	pufName := name + ".PUF"
	rot.pufComponent = puf.NewPUF(pufName, engine, freq, challengeWidth, responseWidth, rot.pufRandID)

	rot.createPorts()
	rot.connectRoTToPUF()
	fmt.Printf("[*] Root of Trust Created: %s (GPU %d) with PUF ID: %x\n",
		rot.Name(), gpuID, rot.pufRandID)
	return rot
}

func (r *RoT) createPorts() {
	// Create communication ports
	r.ToCP = sim.NewLimitNumMsgPort(r, 1, r.Name()+".ToCP")
	r.AddPort("ToCP", r.ToCP)

	// Create dedicated port for PUF communication
	r.ToPUF = sim.NewLimitNumMsgPort(r, 1, r.Name()+".ToPUF")
	r.AddPort("ToPUF", r.ToPUF)

	r.toPUFSender = sim.NewBufferedSender(
		r.ToPUF,
		sim.NewBuffer(r.Name()+".ToPUFSenderBuffer", math.MaxInt32),
	)
}

func (r *RoT) connectRoTToPUF() {
	rotConn := sim.NewDirectConnection(r.Name()+".RoTToPUF", r.Engine, r.Freq)
	rotConn.PlugIn(r.ToPUF, 1)
	rotConn.PlugIn(r.pufComponent.ToRoT, 1)
	r.PUF = r.pufComponent.ToRoT
}

// Tick processes incoming requests to the RoT
func (r *RoT) Tick(now sim.VTimeInSec) bool {
	madeProgress := false

	madeProgress = r.sendMsgsOut(now) || madeProgress
	madeProgress = r.processRequestFromCP(now) || madeProgress
	madeProgress = r.processResponseFromPUF(now) || madeProgress

	return madeProgress
}

func (r *RoT) sendMsgsOut(now sim.VTimeInSec) bool {
	madeProgress := false

	madeProgress = r.sendMsgsOutFromPort(now, r.toPUFSender) || madeProgress

	return madeProgress
}

func (r *RoT) sendMsgsOutFromPort(
	now sim.VTimeInSec,
	sender sim.BufferedSender,
) (madeProgress bool) {
	for {
		ok := sender.Tick(now)
		if ok {
			madeProgress = true
		} else {
			return madeProgress
		}
	}
}

func (r *RoT) processRequestFromCP(now sim.VTimeInSec) bool {
	msg := r.ToCP.Peek()
	if msg == nil {
		return false
	}

	switch req := msg.(type) {
	case *puf.PUFChallenge:
		fmt.Printf("[*] PUF Challenge Received in RoT from CP: %x\n", req.Challenge)

		// Forward the challenge to the PUF component
		challengeMsg := &puf.PUFChallenge{
			MsgMeta: sim.MsgMeta{
				ID:       sim.GetIDGenerator().Generate(),
				SendTime: now,
				Src:      r.ToPUF,
				Dst:      r.PUF,
			},
			Challenge: req.Challenge,
		}
		r.cpRoT = req.Meta().Src // save the source port of the challenge message (RoT in CP)
		r.toPUFSender.Send(challengeMsg)
		r.ToCP.Retrieve(now)

		fmt.Printf("[*] PUF Challenge sent to PUF port (%s): %x\n", r.pufComponent.Name(), challengeMsg.Challenge)

		return true
	case *protocol.GPUAttestationReq:
		fmt.Printf("[*] Attestation Request Received in RoT from CP\n")
		report := attestation.GenerateAttestationReport(r.privateKey, req.Nonce, r.gpuID)

		rspMsg := &protocol.GPUAttestationRsp{
			MsgMeta: sim.MsgMeta{
				ID:       sim.GetIDGenerator().Generate(),
				SendTime: now + sim.VTimeInSec(float64(r.responseDelay)/float64(r.Freq)),
				Src:      r.ToCP,
				Dst:      req.Meta().Src,
			},
			Report: report,
		}

		r.ToCP.Send(rspMsg)
		r.ToCP.Retrieve(now)
		fmt.Printf("[*] Attestation Response Sent from RoT to CP\n")
		return true
	}
	return false
}

func (r *RoT) processResponseFromPUF(now sim.VTimeInSec) bool {
	msg := r.ToPUF.Peek()
	if msg == nil {
		return false
	}
	switch rsp := msg.(type) {
	case *puf.PUFResponse:
		// Forward the response back to the CP
		responseMsg := &puf.PUFResponse{
			MsgMeta: sim.MsgMeta{
				ID:       sim.GetIDGenerator().Generate(),
				SendTime: now,
				Src:      r.ToCP,
				Dst:      r.cpRoT,
			},
			Response: rsp.Response,
		}

		r.ToCP.Send(responseMsg)
		r.ToPUF.Retrieve(now)
		fmt.Printf("[*] PUF Response Forwarded from RoT to CP: %x\n", rsp.Response)
		return true
	}

	return false
}

// GetPortByName returns the port with the given name
func (r *RoT) GetPortByName(name string) sim.Port {
	switch name {
	case "ToCP":
		return r.ToCP
	case "ToPUF":
		return r.ToPUF
	default:
		return nil
	}
}
