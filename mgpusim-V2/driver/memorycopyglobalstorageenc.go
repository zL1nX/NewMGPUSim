package driver

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/sarchlab/akita/v3/sim"
)

// defaultMemoryCopyMiddleware handles memory copy commands and related
// communication.
// encryptedGlobalStorageMemoryCopyMiddleware handles memory copy commands with encryption/decryption
type encryptedGlobalStorageMemoryCopyMiddleware struct {
	driver *Driver

	// Encryption configuration
	encryptionKey    []byte
	encryptionIV     []byte
	enableEncryption bool

	// Performance modeling
	cyclesPerEncrypt int
	cyclesPerDecrypt int
	cyclesLeft       int

	// Command tracking
	currentCmd   Command
	currentQueue *CommandQueue
}

func NewEncryptedGlobalStorageMemoryCopyMiddleware(
	driver *Driver,
) *encryptedGlobalStorageMemoryCopyMiddleware {
	m := &encryptedGlobalStorageMemoryCopyMiddleware{
		driver: driver,
		// Default AES-128 key and IV
		encryptionKey: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
		encryptionIV: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
		enableEncryption: true,
		cyclesPerEncrypt: 100,
		cyclesPerDecrypt: 100,
	}
	return m
}

// SetKey sets the encryption key
func (m *encryptedGlobalStorageMemoryCopyMiddleware) SetKey(key []byte) {
	if len(key) != 16 {
		log.Panic("Key must be 16 bytes for AES-128")
	}
	m.encryptionKey = key
}

// SetIV sets the initialization vector
func (m *encryptedGlobalStorageMemoryCopyMiddleware) SetIV(iv []byte) {
	if len(iv) != 16 {
		log.Panic("IV must be 16 bytes for AES-128")
	}
	m.encryptionIV = iv
}

// EnableEncryption enables or disables encryption
func (m *encryptedGlobalStorageMemoryCopyMiddleware) EnableEncryption(enable bool) {
	m.enableEncryption = enable
}

func (m *encryptedGlobalStorageMemoryCopyMiddleware) encryptData(data []byte) ([]byte, error) {
	// Pad data to AES block size (16 bytes)
	padding := aes.BlockSize - (len(data) % aes.BlockSize)
	if padding < aes.BlockSize {
		paddingBytes := bytes.Repeat([]byte{byte(padding)}, padding)
		data = append(data, paddingBytes...)
	}

	// Create AES cipher
	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return nil, err
	}

	// Use CBC mode for encryption
	mode := cipher.NewCBCEncrypter(block, m.encryptionIV)

	// Create output buffer and encrypt
	encrypted := make([]byte, len(data))
	mode.CryptBlocks(encrypted, data)

	return encrypted, nil
}

func (m *encryptedGlobalStorageMemoryCopyMiddleware) decryptData(data []byte) ([]byte, error) {
	// Create AES cipher
	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		return nil, err
	}

	// Use CBC mode for decryption
	mode := cipher.NewCBCDecrypter(block, m.encryptionIV)

	// Create output buffer and decrypt
	decrypted := make([]byte, len(data))
	mode.CryptBlocks(decrypted, data)

	// Remove padding
	padding := int(decrypted[len(decrypted)-1])
	if padding > 0 && padding <= aes.BlockSize {
		decrypted = decrypted[:len(decrypted)-padding]
	}

	return decrypted, nil
}

func (m *encryptedGlobalStorageMemoryCopyMiddleware) ProcessCommand(
	now sim.VTimeInSec,
	cmd Command,
	queue *CommandQueue,
) (processed bool) {
	switch cmd := cmd.(type) {
	case *MemCopyH2DCommand:
		return m.processMemCopyH2DCommand(now, cmd, queue)
	case *MemCopyD2HCommand:
		return m.processMemCopyD2HCommand(now, cmd, queue)
	}

	return false
}

func (m *encryptedGlobalStorageMemoryCopyMiddleware) processMemCopyH2DCommand(
	now sim.VTimeInSec,
	cmd *MemCopyH2DCommand,
	queue *CommandQueue,
) bool {
	buffer := bytes.NewBuffer(nil)
	err := binary.Write(buffer, binary.LittleEndian, cmd.Src)
	if err != nil {
		panic(err)
	}
	rawBytes := buffer.Bytes()

	encryptedBytes, err := m.encryptData(rawBytes)
	if err != nil {
		panic(err)
	}
	rawBytes = encryptedBytes

	fmt.Println("rawBytes", rawBytes)

	offset := uint64(0)
	addr := uint64(cmd.Dst)
	sizeLeft := uint64(len(rawBytes))
	for sizeLeft > 0 {
		page, found := m.driver.pageTable.Find(queue.Context.pid, addr)
		if !found {
			panic("page not found")
		}

		pAddr := page.PAddr + (addr - page.VAddr)
		sizeLeftInPage := page.PageSize - (addr - page.VAddr)
		sizeToCopy := sizeLeftInPage
		if sizeLeft < sizeLeftInPage {
			sizeToCopy = sizeLeft
		}

		m.driver.globalStorage.Write(pAddr, rawBytes[offset:offset+sizeToCopy])

		sizeLeft -= sizeToCopy
		addr += sizeToCopy
		offset += sizeToCopy
	}

	// If encryption is enabled, track the current command and queue
	// if m.enableEncryption {
	// 	m.cyclesLeft = m.cyclesPerEncrypt
	// 	m.currentCmd = cmd
	// 	m.currentQueue = queue
	// 	queue.IsRunning = true
	// 	return true
	// }

	queue.IsRunning = false
	queue.Dequeue()

	return true
}

func (m *encryptedGlobalStorageMemoryCopyMiddleware) processMemCopyD2HCommand(
	now sim.VTimeInSec,
	cmd *MemCopyD2HCommand,
	queue *CommandQueue,
) bool {
	cmd.RawData = make([]byte, binary.Size(cmd.Dst))

	offset := uint64(0)
	addr := uint64(cmd.Src)
	sizeLeft := uint64(len(cmd.RawData))
	fmt.Println("sizeLeft", sizeLeft)
	for sizeLeft > 0 {
		page, found := m.driver.pageTable.Find(queue.Context.pid, addr)
		if !found {
			panic("page not found")
		}

		pAddr := page.PAddr + (addr - page.VAddr)
		sizeLeftInPage := page.PageSize - (addr - page.VAddr)
		sizeToCopy := sizeLeftInPage
		if sizeLeft < sizeLeftInPage {
			sizeToCopy = sizeLeft
		}

		data, _ := m.driver.globalStorage.Read(pAddr, sizeToCopy)
		fmt.Println("D2H data", data)
		copy(cmd.RawData[offset:], data)

		sizeLeft -= sizeToCopy
		addr += sizeToCopy
		offset += sizeToCopy
	}
	decryptedData, err := m.decryptData(cmd.RawData)
	if err != nil {
		panic(err)
	}

	copy(cmd.RawData, decryptedData)

	fmt.Println("decryptedData", cmd.RawData)
	buf := bytes.NewReader(cmd.RawData)
	err = binary.Read(buf, binary.LittleEndian, cmd.Dst)
	if err != nil {
		panic(err)
	}

	// If encryption is enabled, track the current command and queue
	// if m.enableEncryption {
	// 	m.cyclesLeft = m.cyclesPerDecrypt
	// 	m.currentCmd = cmd
	// 	m.currentQueue = queue
	// 	queue.IsRunning = true
	// 	return true
	// }

	queue.IsRunning = false
	queue.Dequeue()
	return true
}

// func (m *encryptedGlobalStorageMemoryCopyMiddleware) Tick(
// 	now sim.VTimeInSec,
// ) (madeProgress bool) {
// 	if m.cyclesLeft > 0 {
// 		m.cyclesLeft--

// 		// If encryption/decryption finished, complete the command
// 		if m.cyclesLeft == 0 {
// 			// Need to store a reference to the current queue and command
// 			if m.currentQueue != nil {
// 				m.currentQueue.IsRunning = false
// 				m.currentQueue.Dequeue()
// 				m.currentQueue = nil
// 			}
// 		}

// 		return true
// 	}

// 	return false
// }

func (m *encryptedGlobalStorageMemoryCopyMiddleware) Tick(
	now sim.VTimeInSec,
) (madeProgress bool) {
	return false
}
