package main

import (
	"flag"
	"fmt"

	c "github.com/uncleDecart/ics-api/go/client"
	"github.com/uncleDecart/ics-container/pkg/mock"
)

func main() {
	port := flag.Int("port", 8889, "Port to run mock server")

	artifactMetadataA := "YmF6aW5nYQo="
	fileMetadataA := "Zm9vYmFyCg=="
	fileNameA := "config.yml"
	fileShaA := "406dfb89affce2858e26e209e092cd36358ef41d92c85c53b0dfbab5320174e8"
	sizeA := 42
	blobA := c.BinaryBlob{
		ArtifactMetaData: &artifactMetadataA,
		FileMetaData:     &fileMetadataA,
		FileName:         &fileNameA,
		FileSha:          &fileShaA,
		Size:             &sizeA,
	}

	artifactMetadataB := "YmF6aW5nYQo="
	fileMetadataB := "Zm9vYmFyCg=="
	fileNameB := "configB.yml"
	fileShaB := "406dfb89affce2858e26e209e092cd36358ef41d92c85c53b0dfbab5320174e8"
	sizeB := 42
	blobB := c.BinaryBlob{
		ArtifactMetaData: &artifactMetadataB,
		FileMetaData:     &fileMetadataB,
		FileName:         &fileNameB,
		FileSha:          &fileShaB,
		Size:             &sizeB,
	}

	patchA := "69bac069-f4c3-435a-b4e7-d635b60e9b5a"
	patchB := "69bac069-f4c3-435a-b4e7-777777777777"
	version := "1"

	blobs := []c.BinaryBlob{blobA, blobB}
	envelopesAvailable := []c.PatchEnvelopeDescription{
		{
			PatchID:     &patchA,
			Version:     &version,
			BinaryBlobs: &blobs,
		},
		{
			PatchID:     &patchB,
			Version:     &version,
			BinaryBlobs: &blobs,
		},
	}

	data := map[string]string{
		patchA + ":" + fileNameA: "beforeAA",
		patchA + ":" + fileNameB: "beforeAB",
		patchB + ":" + fileNameA: "beforeBA",
		patchB + ":" + fileNameB: "beforeBB",
	}

	server := mock.CreateMockMetadataServer(&envelopesAvailable, &data)
	server.Addr = fmt.Sprintf(":%d", *port)

	fmt.Printf("Serving on %d \n", *port)

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Failed to run srv %v\n", err)
	}
}
