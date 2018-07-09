package geodb

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
)

// LoadData returns emmbeded database as byte array
func EncodedDataLoader(data string, originalSize int, compressed bool) ([]byte, error) {
	decoded, err := base64.RawStdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	if compressed {
		reader := bytes.NewReader(decoded)
		decompressingReader, err := gzip.NewReader(reader)
		if err != nil {
			return nil, err
		}
		decompressed := make([]byte, originalSize)
		readed, err := decompressingReader.Read(decompressed)
		if err != nil {
			return nil, err
		}
		if readed != originalSize {
			return nil, errors.New("original and decompressed size mismatch")
		}
		return decompressed, nil
	}

	return decoded, nil
}
