package osc

import (
	"bytes"
	"log"
	"net"
)

// padString pads a string to a multiple of 4 bytes by adding null bytes (0x00).
func PadString(input string) []byte {
	// Start by appending a null terminator to the input string
	strWithNull := input + "\x00"

	length := len(strWithNull)

	padding := (4 - (length % 4)) % 4

	// Append the necessary number of null bytes (0-3)
	paddedString := strWithNull + string(bytes.Repeat([]byte{'\x00'}, padding))

	result := []byte(paddedString)
	return result
}

// createOSCPacket constructs the OSC packet with address, type tags, and arguments,
func CreateOSCPacket(address, argument string) []byte {
	var buf bytes.Buffer

	// Write the OSC address (e.g., "/action")
	buf.Write(PadString(address))

	// Write the OSC type tag (e.g., ",s" for a string argument)
	buf.Write(PadString(",s"))

	// Write the OSC argument (e.g., "_S&M_INS_MARKER_PLAY")
	buf.Write(PadString(argument))

	return buf.Bytes()
}

func SendOSC(ip string, port int, commandID string, udp_client net.PacketConn) {
	packet := CreateOSCPacket("/action", commandID)

	RemoteAddr := net.UDPAddr{IP: net.ParseIP(ip), Port: port}

	_, err := udp_client.WriteTo(packet, &RemoteAddr)
	if err != nil {
		log.Printf("ALALALA")
	}
}
