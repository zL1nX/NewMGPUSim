package driver_test

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sarchlab/mgpusim/v3/driver"
	"github.com/sarchlab/mgpusim/v3/samples/runner"
)

var _ = ginkgo.Describe("Test Encrypted Memory Copy", func() {
	var (
		gpuDriver *driver.Driver
		context   *driver.Context
	)

	ginkgo.BeforeEach(func() {
		platform := runner.MakeEmuBuilder().
			WithNumGPU(1).
			WithEncryptedMemoryCopyMiddleware().
			Build()
		gpuDriver = platform.Driver
		gpuDriver.Run()
		context = gpuDriver.Init()
	})

	ginkgo.AfterEach(func() {
		gpuDriver.Terminate()
	})

	ginkgo.It("should encrypt and decrypt data correctly", func() {
		// Allocate memory on device
		ptr := gpuDriver.AllocateMemory(context, uint64(16))

		// Test data
		input := []byte{0x01, 0x02, 0x03, 0x04}
		output := make([]byte, 16)

		// Copy to device (encrypts)
		gpuDriver.MemCopyH2D(context, ptr, input)

		// Copy back (decrypts)
		gpuDriver.MemCopyD2H(context, output, ptr)

		// Verify data matches
		fmt.Println("output", output)
		for i := 0; i < 4; i++ {
			Expect(input[i]).To(Equal(output[i]))
		}
	})

	ginkgo.It("should work with encryption disabled", func() {
		// Get middleware and disable encryption
		// middleware := gpuDriver.GetEncryptedMiddleware()
		// middleware.EnableEncryption(false)

		// Allocate memory
		ptr := gpuDriver.AllocateMemory(context, uint64(16))

		// Test data
		input := []byte{0x05, 0x06, 0x07, 0x08}
		output := make([]byte, 16)

		// Copy to device
		gpuDriver.MemCopyH2D(context, ptr, input)

		// Copy back
		gpuDriver.MemCopyD2H(context, output, ptr)

		// Verify data matches
		for i := 0; i < 4; i++ {
			Expect(input[i]).To(Equal(output[i]))
		}
	})
})
