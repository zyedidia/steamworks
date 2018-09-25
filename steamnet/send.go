package steamnet

import (
	"errors"
	"unsafe"

	"github.com/BenLubar/steamworks"
	"github.com/BenLubar/steamworks/internal"
)

// Reliability specifies the send type of SendPacket.
// Typically Unreliable is what you want for UDP-like packets, Reliable for TCP-like packets.
type Reliability = internal.EP2PSend

const (
	// Basic UDP send. Packets can't be bigger than 1200 bytes (your typical MTU size). Can be lost, or arrive out of order (rare).
	// The sending API does have some knowledge of the underlying connection, so if there is no NAT-traversal accomplished or
	// there is a recognized adjustment happening on the connection, the packet will be batched until the connection is open again.
	Unreliable = internal.EP2PSend_Unreliable

	// As above, but if the underlying P2P connection isn't yet established the packet will just be thrown away.
	// Using this on the first packet sent to a remote host almost guarantees the packet will be dropped.
	// This is only really useful for kinds of data that should never buffer up, e.g. voice payload packets
	UnreliableNoDelay = internal.EP2PSend_UnreliableNoDelay

	// Reliable message send. Can send up to 1MB of data in a single message.
	// Does fragmentation/re-assembly of messages under the hood, as well as a sliding window for efficient sends of large chunks of data.
	Reliable = internal.EP2PSend_Reliable

	// As above, but applies the Nagle algorithm to the send - sends will accumulate until the current MTU size (typically ~1200 bytes, but can change)
	// or ~200ms has passed (Nagle algorithm). This is useful if you want to send a set of smaller messages but have the coalesced into a single packet.
	//
	// Since the reliable stream is all ordered, you can do several small message sends with ReliableWithBuffering and then do a normal Reliable
	// to force all the buffered data to be sent.
	ReliableWithBuffering = internal.EP2PSend_ReliableWithBuffering
)

// Possible errors that can be returned by SendPacket.
var (
	ErrTargetUserInvalid = errors.New("steamnet: target Steam ID is invalid")
	ErrPacketTooLarge    = errors.New("steamnet: packet is too large for the send type")
	ErrBufferFull        = errors.New("steamnet: too many bytes are queued to be sent")
)

// SendPacket sends a P2P packet to the specified user.
//
// This is a session-less API which automatically establishes NAT-traversing or Steam relay server connections.
//
// The first packet send may be delayed as the NAT-traversal code runs.
//
// See Reliability for descriptions of the different ways of sending packets.
//
// The type of data you send is arbitrary, you can use an off the shelf system like Protocol Buffers or
// Cap'n Proto to encode your packets in an efficient way, or you can create your own messaging system.
//
// Note that a nil return value does not mean the packet was successfully received. If the packet is not received
// after a timeout of 20 seconds, an error will be sent to the function registered with RegisterErrorCallback.
func SendPacket(user steamworks.SteamID, data []byte, sendType Reliability, channel int32) error {
	if !internal.SteamAPI_ISteamNetworking_SendP2PPacket(internal.SteamID(user), unsafe.Pointer(&data[0]), uint32(len(data)), sendType, channel) {
		if !user.IsValid() {
			return ErrTargetUserInvalid
		}
		if (sendType < Reliable && len(data) > 1200) || len(data) > 1<<20 {
			return ErrPacketTooLarge
		}
		return ErrBufferFull
	}

	return nil
}