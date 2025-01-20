package io

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p/core/network"
)

// WriteProtoBuffered writes a protobuf message to a libp2p stream using buffered I/O.
func WriteProtoBuffered(stream network.Stream, message proto.Message) error {
	writer := bufio.NewWriter(stream)

	// Serialize the protobuf message
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal protobuf message: %w", err)
	}

	// Write the length of the data as a 4-byte header
	length := uint32(len(data))
	if err := binary.Write(writer, binary.BigEndian, length); err != nil {
		return fmt.Errorf("failed to write data length: %w", err)
	}

	// Write the serialized data
	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("failed to write protobuf data: %w", err)
	}

	// Flush the buffer to ensure all data is sent
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	return nil
}

// ReadProtoBuffered reads a protobuf message from a libp2p stream using buffered I/O.
func ReadProtoBuffered(stream network.Stream, message proto.Message) error {
	reader := bufio.NewReader(stream)

	// Read the 4-byte length header
	var length uint32
	if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
		return fmt.Errorf("failed to read data length: %w", err)
	}

	// Read the data of the given length
	data := make([]byte, length)
	if _, err := io.ReadFull(reader, data); err != nil {
		return fmt.Errorf("failed to read protobuf data: %w", err)
	}

	// Unmarshal the data into the protobuf message
	if err := proto.Unmarshal(data, message); err != nil {
		return fmt.Errorf("failed to unmarshal protobuf data: %w", err)
	}

	return nil
}
