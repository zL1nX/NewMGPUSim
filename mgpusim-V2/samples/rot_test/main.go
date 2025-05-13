package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/sarchlab/akita/v3/sim"
	"github.com/sarchlab/mgpusim/v3/protocol"
	"github.com/sarchlab/mgpusim/v3/puf"
	"github.com/sarchlab/mgpusim/v3/samples/runner"
)

func generateRandomChallenge(size int) []byte {
	challenge := make([]byte, size)
	_, err := rand.Read(challenge)
	if err != nil {
		panic(err)
	}
	return challenge
}

func testPUFchallenge(platform *runner.Platform, numChallenges int, challengeSize int, waitCycle int) {
	// Generate random challenges (32 bytes each = 256 bits)
	challenges := make([][]byte, numChallenges)
	for i := 0; i < numChallenges; i++ {
		challenges[i] = generateRandomChallenge(challengeSize)
	}

	for gpuID := 0; gpuID < len(platform.GPUs); gpuID++ {
		gpu := platform.GPUs[gpuID]
		cpPort := gpu.CommandProcessor.ToDriver
		driverPort := platform.Driver.GetPortByName("GPU")

		fmt.Printf("\n=== Testing PUF on GPU %d ===\n", gpuID)

		for i, challenge := range challenges {
			challengeMsg := &puf.PUFChallenge{
				MsgMeta: sim.MsgMeta{
					ID:       sim.GetIDGenerator().Generate(),
					SendTime: platform.Engine.CurrentTime(),
					Src:      driverPort,
					Dst:      cpPort,
				},
				Challenge: challenge,
			}

			fmt.Printf("[*] Challenge %d to GPU %d:\n%s\n",
				i, gpuID, hex.Dump(challenge))

			driverPort.Send(challengeMsg)
			platform.Engine.Run()

			var rspMsg *puf.PUFResponse
			for j := 0; j < waitCycle; j++ {
				msg := driverPort.Retrieve(platform.Engine.CurrentTime())
				if msg != nil {
					rspMsg = msg.(*puf.PUFResponse)
					break
				}
				platform.Engine.Run()
			}

			if rspMsg != nil {
				fmt.Printf("[*] Response from GPU %d:\n%s\n",
					gpuID, hex.Dump(rspMsg.Response))
			} else {
				fmt.Printf("[*] No response received from GPU %d for challenge %d\n",
					gpuID, i)
			}
		}
	}
}

func testRemoteAttestation(platform *runner.Platform, waitCycle int) {
	gpuNum := len(platform.GPUs)
	resp := make([]*protocol.GPUAttestationRsp, gpuNum)

	// Initialize each GPU's attestation capability and register public keys with driver
	for gpuID := 0; gpuID < gpuNum; gpuID++ {
		gpu := platform.GPUs[gpuID]
		publicKey := gpu.RoTComponent.GetPublicKey()
		if publicKey == nil {
			log.Fatalf("Failed to get public key for GPU %d", gpuID)
		}
		// Register the public key with the driver
		platform.Driver.RegisterGPUPublicKey(uint64(gpuID), publicKey)
		fmt.Printf("Registered public key for GPU %d\n", gpuID)
	}

	// Request attestation
	for gpuID := 0; gpuID < gpuNum; gpuID++ {
		gpu := platform.GPUs[gpuID]
		cpPort := gpu.CommandProcessor.ToDriver
		driverPort := platform.Driver.GetPortByName("GPU")

		req, err := platform.Driver.GenerateAttestationReq(uint64(gpuID))
		if err != nil {
			log.Fatal(err)
		}
		req.Dst = cpPort
		driverPort.Send(req)
		platform.Engine.Run()

		for j := 0; j < waitCycle; j++ {
			msg := driverPort.Retrieve(platform.Engine.CurrentTime())
			if msg != nil {
				resp[gpuID] = msg.(*protocol.GPUAttestationRsp)
				break
			}
			platform.Engine.Run()
		}
	}

	for gpuID := 0; gpuID < gpuNum; gpuID++ {
		if resp[gpuID] == nil {
			fmt.Printf("[*] No response received from this GPU %d\n", gpuID)
		} else {
			fmt.Printf("[*] Response from GPU %d (len: %d)\n", gpuID, len(resp[gpuID].Report.Measurement))
		}
		if platform.Driver.VerifyAttestationReport(resp[gpuID]) {
			fmt.Println("[*] Attestation successful")
		} else {
			fmt.Println("[*] Attestation failed")
		}
	}

}

func main() {
	numGPUs := 2
	numChallenges := 1
	challengeSize := 32
	waitCycle := 20

	platform := runner.MakeR9NanoBuilder().WithNumGPU(numGPUs).Build()

	testPUFchallenge(platform, numChallenges, challengeSize, waitCycle)
	testRemoteAttestation(platform, waitCycle)

}
