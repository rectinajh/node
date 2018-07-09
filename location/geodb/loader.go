package geodb

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"io/ioutil"
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
		defer decompressingReader.Close()
		decompressed, err := ioutil.ReadAll(decompressingReader)
		if err != nil {
			return nil, err
		}
		if len(decompressed) != originalSize {
			return nil, errors.New("original and decompressed data size mismatch")
		}
		return decompressed, nil
	}

	return decoded, nil
}
