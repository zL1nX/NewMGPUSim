package driver

// import (
// 	"bytes"
// 	"crypto/aes"
// 	"crypto/cipher"
// 	"encoding/binary"
// 	"log"

// 	"github.com/sarchlab/akita/v3/sim"
// )

// // encryptedMemoryCopyMiddleware handles memory copy commands and adds encryption/decryption
// type encryptedMemoryCopyMiddleware struct {
// 	driver *Driver

// 	// Configuration
// 	encryptionKey    []byte
// 	encryptionIV     []byte
// 	enableEncryption bool

// 	// Processing state
// 	cyclesPerEncrypt int
// 	cyclesPerDecrypt int
// 	cyclesLeft       int
// 	pendingCommands  map[string]Command
// }

// // NewEncryptedMemoryCopyMiddleware creates a new middleware for encrypted memory transfers
// func NewEncryptedMemoryCopyMiddleware(
// 	driver *Driver,
// ) *encryptedMemoryCopyMiddleware {
// 	m := &encryptedMemoryCopyMiddleware{
// 		driver: driver,
// 		// Default AES-128 key and IV (in production, these should be securely generated)
// 		encryptionKey: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
// 			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
// 		encryptionIV: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
// 			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f},
// 		enableEncryption: true,
// 		cyclesPerEncrypt: 100, // Simulate encryption overhead
// 		cyclesPerDecrypt: 100, // Simulate decryption overhead
// 		pendingCommands:  make(map[string]Command),
// 	}
// 	return m
// }

// // SetKey sets the encryption key
// func (m *encryptedMemoryCopyMiddleware) SetKey(key []byte) {
// 	if len(key) != 16 {
// 		log.Panic("Key must be 16 bytes for AES-128")
// 	}
// 	m.encryptionKey = key
// }

// // SetIV sets the encryption initialization vector
// func (m *encryptedMemoryCopyMiddleware) SetIV(iv []byte) {
// 	if len(iv) != 16 {
// 		log.Panic("IV must be 16 bytes for AES-128")
// 	}
// 	m.encryptionIV = iv
// }

// // EnableEncryption enables or disables encryption
// func (m *encryptedMemoryCopyMiddleware) EnableEncryption(enable bool) {
// 	m.enableEncryption = enable
// }

// // ProcessCommand handles commands and adds encryption/decryption
// func (m *encryptedMemoryCopyMiddleware) ProcessCommand(
// 	now sim.VTimeInSec,
// 	cmd Command,
// 	queue *CommandQueue,
// ) (processed bool) {
// 	switch cmd := cmd.(type) {
// 	case *MemCopyH2DCommand:
// 		return m.processMemCopyH2DCommand(now, cmd, queue)
// 	case *MemCopyD2HCommand:
// 		return m.processMemCopyD2HCommand(now, cmd, queue)
// 	case *LaunchKernelCommand:
// 		return m.processMemCopyD2DCommand(now, cmd, queue)
// 	}

// 	// Pass to next middleware if not handled
// 	return false
// }

// func (m *encryptedMemoryCopyMiddleware) processMemCopyH2DCommand(
// 	now sim.VTimeInSec,
// 	cmd *MemCopyH2DCommand,
// 	queue *CommandQueue,
// ) bool {
// 	if !m.enableEncryption {
// 		return false
// 	}

// 	// Create a buffer from the source data
// 	buffer := bytes.NewBuffer(nil)
// 	err := binary.Write(buffer, binary.LittleEndian, cmd.Src)
// 	if err != nil {
// 		panic(err)
// 	}
// 	rawBytes := buffer.Bytes()

// 	// Encrypt the data
// 	encryptedBytes, err := m.encryptData(rawBytes)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Create a new command with encrypted data
// 	encryptedCmd := &MemCopyH2DCommand{
// 		ID:  cmd.ID,
// 		Dst: cmd.Dst,
// 		Src: encryptedBytes,
// 	}

// 	// Pass to the next middleware
// 	m.cyclesLeft = m.cyclesPerEncrypt
// 	queue.IsRunning = true

// 	return m.next.ProcessCommand(now, encryptedCmd, queue)
// }

// func (m *encryptedMemoryCopyMiddleware) processMemCopyD2HCommand(
// 	now sim.VTimeInSec,
// 	cmd *MemCopyD2HCommand,
// 	queue *CommandQueue,
// ) bool {
// 	if !m.enableEncryption {
// 		return m.next.ProcessCommand(now, cmd, queue)
// 	}

// 	// This requires special handling since we need to decrypt after the data is copied
// 	// First, let the next middleware handle the command
// 	processed := m.next.ProcessCommand(now, cmd, queue)

// 	if processed {
// 		// After the transfer is done, decrypt the data
// 		// This would typically happen in the Tick method after the transfer completes
// 		m.cyclesLeft = m.cyclesPerDecrypt
// 	}

// 	return processed
// }

// func (m *encryptedMemoryCopyMiddleware) processMemCopyD2DCommand(
// 	now sim.VTimeInSec,
// 	cmd *LaunchKernelCommand,
// 	queue *CommandQueue,
// ) bool {
// 	if !m.enableEncryption {
// 		return m.next.ProcessCommand(now, cmd, queue)
// 	}

// 	// For D2D transfers, we simulate encryption and decryption overhead
// 	m.cyclesLeft = m.cyclesPerEncrypt + m.cyclesPerDecrypt
// 	queue.IsRunning = true

// 	// Pass the command through without modification
// 	// In a real implementation, we would need to:
// 	// 1. Read from source address
// 	// 2. Encrypt the data
// 	// 3. Copy to a temporary location
// 	// 4. Launch the kernel to copy the encrypted data
// 	// 5. Then decrypt at the destination
// 	// However, this would require significant architecture changes

// 	return m.next.ProcessCommand(now, cmd, queue)
// }

// func (m *encryptedMemoryCopyMiddleware) Tick(
// 	now sim.VTimeInSec,
// ) (madeProgress bool) {
// 	if m.cyclesLeft > 0 {
// 		m.cyclesLeft--
// 		return true
// 	}

// 	// Delegate to next middleware
// 	if m.next != nil {
// 		if processor, ok := m.next.(TickingComponent); ok {
// 			return processor.Tick(now)
// 		}
// 	}

// 	return false
// }

// func (m *encryptedMemoryCopyMiddleware) Pause() {
// 	if m.next != nil {
// 		if pausable, ok := m.next.(PausableComponent); ok {
// 			pausable.Pause()
// 		}
// 	}
// }

// func (m *encryptedMemoryCopyMiddleware) Resume() {
// 	if m.next != nil {
// 		if pausable, ok := m.next.(PausableComponent); ok {
// 			pausable.Resume()
// 		}
// 	}
// }

// func (m *encryptedMemoryCopyMiddleware) encryptData(data []byte) ([]byte, error) {
// 	// Pad the data to 16-byte boundary if needed
// 	padding := 16 - (len(data) % 16)
// 	if padding < 16 {
// 		paddingBytes := bytes.Repeat([]byte{byte(padding)}, padding)
// 		data = append(data, paddingBytes...)
// 	}

// 	// Set up AES encryption
// 	block, err := aes.NewCipher(m.encryptionKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Use CBC mode for encryption
// 	mode := cipher.NewCBCEncrypter(block, m.encryptionIV)

// 	// Create output buffer
// 	encrypted := make([]byte, len(data))

// 	// Encrypt data
// 	mode.CryptBlocks(encrypted, data)

// 	return encrypted, nil
// }

// func (m *encryptedMemoryCopyMiddleware) decryptData(data []byte) ([]byte, error) {
// 	// Set up AES decryption
// 	block, err := aes.NewCipher(m.encryptionKey)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Use CBC mode for decryption
// 	mode := cipher.NewCBCDecrypter(block, m.encryptionIV)

// 	// Create output buffer
// 	decrypted := make([]byte, len(data))

// 	// Decrypt data
// 	mode.CryptBlocks(decrypted, data)

// 	// Remove padding
// 	padding := int(decrypted[len(decrypted)-1])
// 	if padding > 0 && padding <= 16 {
// 		decrypted = decrypted[:len(decrypted)-padding]
// 	}

// 	return decrypted, nil
// }
