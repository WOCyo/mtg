package record

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
)

type Record struct {
	Type    Type
	Version Version
	Payload bytes.Buffer
}

func (r *Record) String() string {
	return fmt.Sprintf("<tlsRecord(type=%v, version=%v, payload=%s)>",
		r.Type,
		r.Version,
		base64.StdEncoding.EncodeToString(r.Payload.Bytes()))
}

func (r *Record) Reset() {
	r.Payload.Reset()
}

func (r *Record) Read(reader io.Reader) error {
	r.Reset()

	buf := [2]byte{}

	if _, err := io.ReadFull(reader, buf[:1]); err != nil {
		return fmt.Errorf("cannot read type: %w", err)
	}

	r.Type = Type(buf[0])
	if err := r.Type.Valid(); err != nil {
		return fmt.Errorf("invalid type: %w", err)
	}

	if _, err := io.ReadFull(reader, buf[:]); err != nil {
		return fmt.Errorf("cannot read version: %w", err)
	}

	r.Version = Version(binary.BigEndian.Uint16(buf[:]))
	if err := r.Version.Valid(); err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	if _, err := io.ReadFull(reader, buf[:]); err != nil {
		return fmt.Errorf("cannot read payload length: %w", err)
	}

	length := int64(binary.BigEndian.Uint16(buf[:]))
	if _, err := io.CopyN(&r.Payload, reader, length); err != nil {
		return fmt.Errorf("cannot read payload: %w", err)
	}

	return nil
}

func (r *Record) Dump(writer io.Writer) error {
	buf := acquireBytesBuffer()
	defer releaseBytesBuffer(buf)

	bufSlice := [2]byte{byte(r.Type), 0}
	buf.Write(bufSlice[:1])

	binary.BigEndian.PutUint16(bufSlice[:], uint16(r.Version))
	buf.Write(bufSlice[:])

	binary.BigEndian.PutUint16(bufSlice[:], uint16(r.Payload.Len()))
	buf.Write(bufSlice[:])

	buf.Write(r.Payload.Bytes())

	if _, err := buf.WriteTo(writer); err != nil {
		return fmt.Errorf("cannot dump record: %w", err)
	}

	return nil
}
