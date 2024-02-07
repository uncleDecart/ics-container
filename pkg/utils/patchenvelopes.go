package utils

import (
	"encoding/base64"
	"fmt"

	c "github.com/uncleDecart/ics-api/go/client"
)

func FindBinaryBlob(pe c.PatchEnvelopeDescription, name string) *c.BinaryBlob {
	for _, b := range *pe.BinaryBlobs {
		if *b.FileName == name {
			return &b
		}
	}
	return nil
}

func FindPatchEnvelope(pes []c.PatchEnvelopeDescription, patchID string) *c.PatchEnvelopeDescription {
	for _, pe := range pes {
		if *pe.PatchID == patchID {
			return &pe
		}
	}
	return nil
}

func DecodeBase64(data []byte) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(decoded, data)
	if err != nil {
		return []byte{}, err
	}
	fmt.Println(string(decoded))
	decoded = decoded[:n]
	return decoded, nil
}
