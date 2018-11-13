package dtls

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/x509"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
)

const initialTickerInterval = time.Second
const finalTickerInternal = 90 * time.Second

type handshakeMessageHandler func(*Conn) error
type timerThread func(*Conn)

// Conn represents a DTLS connection
type Conn struct {
	lock           sync.RWMutex    // Internal lock (must not be public)
	nextConn       net.Conn        // Embedded Conn, typically a udpconn we read/write from
	fragmentBuffer *fragmentBuffer // out-of-order and missing fragment handling
	handshakeCache *handshakeCache // caching of handshake messages for verifyData generation
	decrypted      chan []byte     // Decrypted Application Data, pull by calling `Read`
	workerTicker   *time.Ticker

	outboundEpoch          uint16
	outboundSequenceNumber uint64 // uint48

	currFlight                          *flight
	cipherSuite                         *cipherSuite // nil if a cipherSuite hasn't been chosen
	localRandom, remoteRandom           handshakeRandom
	localCertificate, remoteCertificate *x509.Certificate
	localKeypair, remoteKeypair         *namedCurveKeypair
	cookie                              []byte
	localVerifyData                     []byte // cached VerifyData

	keys                *encryptionKeys
	localGCM, remoteGCM cipher.AEAD

	handshakeMessageHandler handshakeMessageHandler
	timerThread             timerThread
}

func createConn(nextConn net.Conn, timerThread timerThread, handshakeMessageHandler handshakeMessageHandler, isClient bool) *Conn {
	c := &Conn{
		nextConn:                nextConn,
		currFlight:              newFlight(isClient),
		fragmentBuffer:          newFragmentBuffer(),
		handshakeCache:          newHandshakeCache(),
		handshakeMessageHandler: handshakeMessageHandler,
		timerThread:             timerThread,

		decrypted:    make(chan []byte),
		workerTicker: time.NewTicker(initialTickerInterval),
	}
	c.localRandom.populate()
	c.localKeypair, _ = generateKeypair(namedCurveX25519)

	go c.readThread()
	go c.timerThread(c)
	return c
}

// Dial establishes a DTLS connection over an existing conn
func Dial(conn net.Conn) (*Conn, error) {
	return createConn(conn, clientTimerThread, clientHandshakeHandler /*isClient*/, true), nil
}

// Server listens for incoming DTLS connections
func Server(conn net.Conn) (*Conn, error) {
	return createConn(conn, serverTimerThread, serverHandshakeHandler /*isClient*/, false), nil
}

// Read reads data from the connection.
func (c *Conn) Read(p []byte) (n int, err error) {
	out := <-c.decrypted
	if len(p) < len(out) {
		return 0, errBufferTooSmall
	}

	copy(p, out)
	return len(p), nil
}

// Write writes len(p) bytes from p to the DTLS connection
func (c *Conn) Write(p []byte) (n int, err error) {
	return // TODO encrypt + send ApplicationData
}

// Close closes the connection.
func (c *Conn) Close() error {
	c.nextConn.Close() // TODO Is there a better way to stop read in readThread?
	return nil
}

// Pulls from nextConn
func (c *Conn) readThread() {
	b := make([]byte, 8192)
	for {
		i, err := c.nextConn.Read(b)
		if err != nil {
			panic(err)
		}
		if err := c.handleIncoming(b[:i]); err != nil {
			panic(err)
		}
	}
}

func (c *Conn) internalSend(pkt *recordLayer, shouldEncrypt bool) {
	raw, err := pkt.marshal()
	if err != nil {
		panic(err)
	}

	if h, ok := pkt.content.(*handshake); ok {
		c.handshakeCache.push(raw[recordLayerHeaderSize:], pkt.recordLayerHeader.epoch,
			h.handshakeHeader.messageSequence /* isLocal */, true)
	}

	if shouldEncrypt {
		payload := raw[recordLayerHeaderSize:]
		raw = raw[:recordLayerHeaderSize]

		nonce := append(append([]byte{}, c.keys.clientWriteIV[:4]...), make([]byte, 8)...)
		if _, err := rand.Read(nonce[4:]); err != nil {
			panic(err)
		}

		var additionalData [13]byte
		// SequenceNumber MUST be set first
		// we only want uint48, clobbering an extra 2 (using uint64, Golang doesn't have uint48)
		binary.BigEndian.PutUint64(additionalData[:], pkt.recordLayerHeader.sequenceNumber)
		binary.BigEndian.PutUint16(additionalData[:], pkt.recordLayerHeader.epoch)
		copy(additionalData[8:], raw[:3])
		binary.BigEndian.PutUint16(additionalData[len(additionalData)-2:], uint16(len(payload)))

		encryptedPayload := c.localGCM.Seal(nil, nonce, payload, additionalData[:])
		encryptedPayload = append(nonce[4:], encryptedPayload...)
		raw = append(raw, encryptedPayload...)

		// Update recordLayer size to include explicit nonce
		binary.BigEndian.PutUint16(raw[recordLayerHeaderSize-2:], uint16(len(raw)-recordLayerHeaderSize))
	}

	c.nextConn.Write(raw)
}

func (c *Conn) handleIncoming(buf []byte) error {
	pkts, err := unpackDatagram(buf)
	if err != nil {
		return err
	}

	for _, p := range pkts {
		pushSuccess, err := c.fragmentBuffer.push(p)
		if err != nil {
			return err
		} else if pushSuccess {
			// This was a fragmented buffer, therefore a handshake
			return c.handshakeMessageHandler(c)
		}

		r := &recordLayer{}
		if err := r.unmarshal(p); err != nil {
			return err
		}
		switch content := r.content.(type) {
		case *alert:
			return fmt.Errorf(spew.Sdump(content))
		default:
			return fmt.Errorf("Unhandled contentType %d", content.contentType())
		}
	}
	return nil
}
