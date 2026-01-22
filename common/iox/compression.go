/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package iox

import (
	"io"

	"github.com/klauspost/compress/snappy"
)

type CompressionRw struct {
	reader  io.Reader
	writer  io.Writer
	cReader *snappy.Reader
	cWriter *snappy.Writer
}

func (c *CompressionRw) Close() error {
	if c.cReader != nil {
		PutSnappyReader(c.cReader)
	}
	if c.cWriter != nil {
		_ = c.cWriter.Close()
		PutSnappyWriter(c.cWriter)
	}
	return nil
}

func NewCompressionRw(reader io.Reader, writer io.Writer) *CompressionRw {
	return &CompressionRw{
		reader:  reader,
		writer:  writer,
		cReader: GetSnappyReader(reader),
		cWriter: GetSnappyWriter(writer),
	}
}

func (c *CompressionRw) Reader() io.Reader {
	return c.reader
}

func (c *CompressionRw) Writer() io.Writer {
	return c.writer
}

func (c *CompressionRw) Write(p []byte) (n int, err error) {
	//n, err = c.writer.Write(p)
	n, err = c.cWriter.Write(p)
	_ = c.cWriter.Flush()
	return
}

func (c *CompressionRw) Read(p []byte) (n int, err error) {
	return c.cReader.Read(p)
}
