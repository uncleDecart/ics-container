package utils

import c "github.com/uncleDecart/ics-api/go/client"

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
