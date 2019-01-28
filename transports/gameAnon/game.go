/*
 * Copyright (c) 2014, Yawning Angel <yawning at torproject dot org>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *  * Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 *
 *  * Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

// Package obfs4 provides an implementation of the Tor Project's obfs4
// obfuscation protocol.
package gameAnon

import (
	
	"fmt"
	"bytes"
	"net"
	"syscall"
	"time"
	"encoding/binary"

	"git.torproject.org/pluggable-transports/obfs4.git/common/log"
	"git.torproject.org/pluggable-transports/goptlib.git"
	
	
	
	"../base"
)

const (
	transportName = "GameAnon"

	nodeIDArg     = "node-id"
	
	kArg          = "k"

	MaximumFramePayloadLength  = 1400

	DataPacketType = uint8(0)
	DummyPacketType = uint8(1)

	HeaderSize = 3

	tickTime = 50
	






)

// biasedDist controls if the probability table will be ScrambleSuit style or
// uniformly distributed.




type gameClientArgs struct {
	K int
}

// Transport is the obfs4 implementation of the base.Transport interface.
type Transport struct{}

// Name returns the name of the obfs4 transport protocol.
func (t *Transport) Name() string {
	return transportName
}

// ClientFactory returns a new gameClientFactory instance.
func (t *Transport) ClientFactory(stateDir string) (base.ClientFactory, error) {
	cf := &gameClientFactory{transport: t}
	return cf, nil
}

// ServerFactory returns a new gameServerFactory instance.
func (t *Transport) ServerFactory(stateDir string, args *pt.Args) (base.ServerFactory, error) {

	// Store the arguments that should appear in our descriptor for the clients.
	ptArgs := pt.Args{}

	



	sf := &gameServerFactory{t, &ptArgs,4}	
	return sf, nil
}

type gameClientFactory struct {
	transport base.Transport
}

func (cf *gameClientFactory) Transport() base.Transport {
	return cf.transport
}

func (cf *gameClientFactory) ParseArgs(args *pt.Args) (interface{}, error) {
	

	return &gameClientArgs{2}, nil
}

func (cf *gameClientFactory) Dial(network, addr string, dialFn base.DialFunc, args interface{}) (net.Conn, error) {
	// Validate args before bothering to open connection.


	ca, ok := args.(*gameClientArgs)
	if !ok {
		return nil, fmt.Errorf("invalid argument type for args")
	}

	conn, err := dialFn(network, addr)
	if err != nil {
		return nil, err
	}
	dialConn := conn

	if conn, err = newGameClientConn(conn, ca); err != nil {
		dialConn.Close()
		return nil, err
	}
	return conn, nil
}

type gameServerFactory struct {
	transport base.Transport
	args      *pt.Args
	K			int
}

func (sf *gameServerFactory) Transport() base.Transport {
	return sf.transport
}

func (sf *gameServerFactory) Args() *pt.Args {
	return sf.args
}

func (sf *gameServerFactory) WrapConn(conn net.Conn) (net.Conn, error) {
	// Not much point in having a separate newObfs4ServerConn routine when
	// wrapping requires using values from the factory instance.

	// Generate the session keypair *before* consuming data from the peer, to
	// attempt to mask the rejection sampling due to use of Elligator2.  This
	// might be futile, but the timing differential isn't very large on modern
	// hardware, and there are far easier statistical attacks that can be
	// mounted as a distinguisher.
	log.Infof("Wrap Conn")
	c := &gameConn{conn, true,sf.K,bytes.NewBuffer(nil),make(chan bool),time.NewTicker(time.Millisecond * tickTime),make([]byte, MaximumFramePayloadLength),make([]byte, 1600*6),bytes.NewBuffer(nil),bytes.NewBuffer(nil),bytes.NewBuffer(nil),8000}

	go c.PeriodicWrite()
	

	return c, nil
}




type gameConn struct {
	net.Conn
	isServer bool
	K int
	writeBuffer *bytes.Buffer
	done chan bool
	writeTicker *time.Ticker
	writeDummyBuffer []byte
	readDummyBuffer []byte
	receiveBuffer        *bytes.Buffer
	receiveDecodedBuffer        *bytes.Buffer
	transmitBuffer *bytes.Buffer
	writeRateByte    int
}

func newGameClientConn(conn net.Conn, args *gameClientArgs) (c *gameConn, err error) {
	// Generate the initial protocol polymorphism distribution(s).


	log.Infof("CLient Conn")
	// Allocate the client structure.
	c = &gameConn{conn, false, args.K,bytes.NewBuffer(nil),make(chan bool),time.NewTicker(time.Millisecond * tickTime),make([]byte, MaximumFramePayloadLength),make([]byte, 1600*6),bytes.NewBuffer(nil),bytes.NewBuffer(nil),bytes.NewBuffer(nil),8000}

	// Start the handshake timeout.
	// deadline := time.Now().Add(clientHandshakeTimeout)
	// if err = conn.SetDeadline(deadline); err != nil {
	// 	return nil, err
	// }

	// if err = c.clientHandshake(args.nodeID, args.publicKey, args.sessionKey); err != nil {
	// 	return nil, err
	// }

	// // Stop the handshake timeout.
	// if err = conn.SetDeadline(time.Time{}); err != nil {
	// 	return nil, err
	// }
	go c.PeriodicWrite()
	
	return c, nil
}

func (conn *gameConn) Read(b []byte) (n int, err error) {
	
	
	
	
	n,err = conn.Conn.Read(conn.readDummyBuffer)

	if n == 0 {
		return
	}
	

	conn.receiveBuffer.Write(conn.readDummyBuffer[:n])
	
	



	var decoded [MaximumFramePayloadLength+HeaderSize]byte
	
	
	
	for conn.receiveBuffer.Len() > HeaderSize {

		bn , _ := conn.receiveBuffer.Read(decoded[:])
		
		payloadLen := int( binary.BigEndian.Uint16(decoded[1:]))
	
		if payloadLen+HeaderSize <=bn {
			if decoded[0] == DataPacketType { 
	
				conn.receiveDecodedBuffer.Write(decoded[HeaderSize:HeaderSize+payloadLen])
			}

			if payloadLen + HeaderSize < bn {
	
				resetting := conn.receiveBuffer.Bytes()
				
				conn.receiveBuffer.Reset()
				conn.receiveBuffer.Write(decoded[HeaderSize+payloadLen:bn])
				conn.receiveBuffer.Write(resetting)
			}

		} else {
	
			
			conn.receiveBuffer.Write(decoded[:bn])
			
			break



		}


	}
	if conn.receiveDecodedBuffer.Len() == 0 {
		return 0, nil
	}



	return conn.receiveDecodedBuffer.Read(b)
}

func (conn *gameConn) Write(b []byte) (n int, err error) {
	
	
	return conn.writeBuffer.Write(b)
}

func (conn *gameConn) PeriodicWrite() (err error) {
	log.Infof("Periodic writer")

	ticker := conn.writeTicker.C
	for {
		select {
		case <- conn.done:
			log.Infof("Done")
			return 
		case <- ticker:
			
			conn.transmitBuffer.Reset()
			
			var pkt [MaximumFramePayloadLength+HeaderSize]byte
			
			for conn.writeBuffer.Len() > 0  && conn.transmitBuffer.Len()< (conn.writeRateByte-HeaderSize) {
				bn,err := conn.writeBuffer.Read(conn.writeDummyBuffer)

				remaining := conn.writeRateByte-HeaderSize-conn.transmitBuffer.Len()



				



				if err == nil {
						
					pkt[0] = DataPacketType
					if bn < remaining {
						binary.BigEndian.PutUint16(pkt[1:], uint16((bn)))

						if bn > 0 {
							copy(pkt[3:], conn.writeDummyBuffer[:bn])
						}
						conn.transmitBuffer.Write(pkt[:bn+HeaderSize])
					} else {
						binary.BigEndian.PutUint16(pkt[1:], uint16((remaining)))

						if bn > 0 {
							copy(pkt[3:], conn.writeDummyBuffer[:remaining])
						}
						conn.transmitBuffer.Write(pkt[:remaining+HeaderSize])

						rem := conn.writeBuffer.Bytes()
						conn.writeBuffer.Reset()
						conn.writeBuffer.Write(conn.writeDummyBuffer[remaining:bn])
						conn.writeBuffer.Write(rem)
						

					}

				}
				
				


			}
			dummypkt := make([]byte, MaximumFramePayloadLength+HeaderSize)
			for conn.transmitBuffer.Len()< conn.writeRateByte-HeaderSize {

				dummypkt[0] = DummyPacketType
				

				dn := min ((conn.writeRateByte-conn.transmitBuffer.Len()-HeaderSize),MaximumFramePayloadLength)
				
				binary.BigEndian.PutUint16(dummypkt[1:], uint16((dn)))

				if dn > 0 {
					copy(dummypkt[3:], conn.writeDummyBuffer[:dn])
				}
				conn.transmitBuffer.Write(dummypkt[:dn+HeaderSize])
				


			}
			conn.Conn.Write(conn.transmitBuffer.Bytes())

			
		}
	}
	return 
}



func (conn *gameConn) Close() ( err error) {
	log.Infof("Closing MY Connection ")
	return conn.Conn.Close()
}


func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}


func (conn *gameConn) SetDeadline(t time.Time) error {
	return syscall.ENOTSUP
}

func (conn *gameConn) SetWriteDeadline(t time.Time) error {
	return syscall.ENOTSUP
}

func init() {
	
}

var _ base.ClientFactory = (*gameClientFactory)(nil)
var _ base.ServerFactory = (*gameServerFactory)(nil)
var _ base.Transport = (*Transport)(nil)
var _ net.Conn = (*gameConn)(nil)
