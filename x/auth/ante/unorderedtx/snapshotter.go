package unorderedtx

import (
	"encoding/binary"
	"errors"
	"io"
	"time"

	snapshot "cosmossdk.io/store/snapshots/types"
)

const (
	txHashSize  = 32
	heightSize  = 8
	timeoutSize = 8
	chunkSize   = txHashSize + heightSize + timeoutSize
)

var _ snapshot.ExtensionSnapshotter = &Snapshotter{}

const (
	// SnapshotFormat defines the snapshot format of exported unordered transactions.
	// No protobuf envelope, no metadata.
	SnapshotFormat = 1

	// SnapshotName defines the snapshot name of exported unordered transactions.
	SnapshotName = "unordered_txs"
)

type Snapshotter struct {
	m *Manager
}

func NewSnapshotter(m *Manager) *Snapshotter {
	return &Snapshotter{m: m}
}

func (s *Snapshotter) SnapshotName() string {
	return SnapshotName
}

func (s *Snapshotter) SnapshotFormat() uint32 {
	return SnapshotFormat
}

func (s *Snapshotter) SupportedFormats() []uint32 {
	return []uint32{SnapshotFormat}
}

func (s *Snapshotter) SnapshotExtension(height uint64, payloadWriter snapshot.ExtensionPayloadWriter) error {
	// export all unordered transactions as a single blob
	return s.m.exportSnapshot(height, payloadWriter)
}

func (s *Snapshotter) RestoreExtension(height uint64, format uint32, payloadReader snapshot.ExtensionPayloadReader) error {
	if format == SnapshotFormat {
		return s.restore(height, payloadReader)
	}

	return snapshot.ErrUnknownFormat
}

func (s *Snapshotter) restore(height uint64, payloadReader snapshot.ExtensionPayloadReader) error {
	// the payload should be the entire set of unordered transactions
	payload, err := payloadReader()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return io.ErrUnexpectedEOF
		}

		return err
	}

	if len(payload)%chunkSize != 0 {
		return errors.New("invalid unordered txs payload length")
	}

	var i int
	for i < len(payload) {
		var txHash TxHash
		copy(txHash[:], payload[i:i+txHashSize])

		timestamp := binary.BigEndian.Uint64(payload[i+txHashSize+heightSize : i+chunkSize])
		// need to come up with a way to fetch blocktime to filter out expired txs
		//
		// right now we dont have access block time at this flow, so we would just include the expired txs
		// and let it be purge during purge loop
		if timestamp != 0 {
			s.m.AddTimestamp(txHash, time.Unix(int64(timestamp), 0))
			i += chunkSize
			continue
		}

		ttl := binary.BigEndian.Uint64(payload[i+txHashSize : i+txHashSize+heightSize])
		if height < ttl {
			// only add unordered transactions that are still valid, i.e. unexpired
			s.m.Add(txHash, ttl)
		}

		i += chunkSize
	}

	return nil
}
