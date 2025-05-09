package main

import (
	"fmt"
	"log"

	"github.com/sarchlab/akita/v3/sim"
	"github.com/sarchlab/mgpusim/v3/protocol"
	"github.com/sarchlab/mgpusim/v3/samples/runner"
)

func main() {

	gpuNum := 1
	platform := runner.MakeR9NanoBuilder().WithNumGPU(gpuNum).Build()
	resp := make([]*protocol.GPUAttestationRsp, gpuNum)
	waitCycle := 20

	// Initialize each GPU's attestation capability and register public keys with driver
	for gpuID := 0; gpuID < gpuNum; gpuID++ {
		gpu := platform.GPUs[gpuID]
		publicKey, err := gpu.CommandProcessor.InitAttestation(uint64(gpuID))
		if err != nil {
			log.Fatalf("Failed to initialize attestation for GPU %d: %v", gpuID, err)
		}

		// Register the public key with the driver
		platform.Driver.RegisterGPUPublicKey(uint64(gpuID), publicKey)
		fmt.Printf("Registered public key for GPU %d\n", gpuID)
	}

	startTime := platform.Engine.CurrentTime()

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

	endTime := platform.Engine.CurrentTime()
	freq := sim.GHz // 1 GHz can be replaced with the actual frequency of the CPU
	elapsedCycles := freq.Cycle(endTime - startTime)
	fmt.Printf("[*] Elapsed time: %d cycles\n", elapsedCycles)
}
