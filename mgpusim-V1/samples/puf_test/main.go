package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/sarchlab/akita/v3/sim"
	"github.com/sarchlab/mgpusim/v3/samples/runner"
	"github.com/sarchlab/mgpusim/v3/samples/runner/puf"
)

func generateRandomChallenge(size int) []byte {
	challenge := make([]byte, size)
	_, err := rand.Read(challenge)
	if err != nil {
		panic(err)
	}
	return challenge
}

func main() {
	numGPUs := 4
	numChallenges := 3
	waitCycle := 5

	platform := runner.MakeR9NanoBuilder().WithNumGPU(numGPUs).Build()

	// Generate random challenges (32 bytes each = 256 bits)
	challenges := make([][]byte, numChallenges)
	for i := 0; i < numChallenges; i++ {
		challenges[i] = generateRandomChallenge(32)
	}

	for gpuID := 0; gpuID < numGPUs; gpuID++ {
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
