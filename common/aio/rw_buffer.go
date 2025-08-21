package aio

import (
	"bytes"
	"io"
)

type RWBuffer struct {
	buf    *bytes.Buffer
	closed bool
}

// This function writes a byte slice to the RWBuffer
func (r *RWBuffer) Write(p []byte) (n int, err error) {
	// Check if the RWBuffer is closed
	if r.closed {
		// Return an error if the RWBuffer is closed
		return 0, io.ErrClosedPipe
	}
	// Write the byte slice to the buffer
	return r.buf.Write(p)
}

// This function reads data from the RWBuffer and stores it in the byte slice p
func (r *RWBuffer) Read(p []byte) (n int, err error) {
	// Check if the RWBuffer is closed
	if r.closed {
		// If it is, return an error
		return 0, io.ErrClosedPipe
	}
	// Otherwise, read the data from the buffer and store it in p
	return r.buf.Read(p)
}

// Close This function closes the RWBuffer and sets the closed flag to true
func (r *RWBuffer) Close() error {
	// Set the closed flag to true
	r.closed = true
	// Return nil to indicate that the function was successful
	return nil
}

// NewRWBuffer This function creates a new ReadWriteCloser object
func NewRWBuffer() io.ReadWriteCloser {
	// Create a new RWBuffer object
	return &RWBuffer{
		// Initialize the buffer with a new bytes buffer
		buf: bytes.NewBuffer(nil),
		// Set the closed flag to false
		closed: false,
	}
}
