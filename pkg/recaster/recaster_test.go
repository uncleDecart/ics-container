package recaster_test

import (
	"fmt"
	"testing"

	"github.com/uncleDecart/ics-container/pkg/pemanager"
	"github.com/uncleDecart/ics-container/pkg/recaster"

	"github.com/onsi/gomega"

	c "github.com/uncleDecart/ics-api/go/client"
)

func TestIsEqual(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)

	dummyPeMgr, err := pemanager.NewPatchEnvelopeManager("localhost:2222")
	g.Expect(err).To(gomega.BeNil())

	recA := recaster.NewRecaster(recaster.Config{
		Name:           "Recaster A",
		PatchEnvelopes: []string{"patch1", "patch2"},
	}, dummyPeMgr)

	recB := recaster.NewRecaster(recaster.Config{
		Name:           "Recaster B",
		PatchEnvelopes: []string{"patch1", "patch2"},
	}, dummyPeMgr)

	recC := recaster.NewRecaster(recaster.Config{
		Name:           "Recaster A",
		PatchEnvelopes: []string{"patch3", "patch4"},
	}, dummyPeMgr)

	g.Expect(recA.IsEqual(recB)).To(gomega.BeFalse())
	g.Expect(recB.IsEqual(recA)).To(gomega.BeFalse())

	g.Expect(recA.IsEqual(recC)).To(gomega.BeTrue())
	g.Expect(recC.IsEqual(recA)).To(gomega.BeTrue())
}

func TestTransformString(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)

	patchID := "69bac069-f4c3-435a-b4e7-d635b60e9b5a"
	artifactMetadata := "artifactMetadata"
	fileMetadata := "fileMetadata"
	fileName := "fileName"
	fileSha := "fileSha"
	size := 42
	url := "localhost:2222/patch/download"
	blob := c.BinaryBlob{
		ArtifactMetaData: &artifactMetadata,
		FileMetaData:     &fileMetadata,
		FileName:         &fileName,
		FileSha:          &fileSha,
		Size:             &size,
		Url:              &url,
	}
	env := []c.PatchEnvelopeDescription{
		{
			PatchID:     &patchID,
			BinaryBlobs: &[]c.BinaryBlob{blob},
		},
	}
	data := map[string][]byte{
		patchID: []byte{42, 42, 42, 42},
	}
	peMgr, err := pemanager.NewPatchEnvelopeManagerPreInit("localhost:2222", data, env)
	g.Expect(err).To(gomega.BeNil())

	input := fmt.Sprintf("color : $(%s/%s/size)", patchID, fileName)
	expected := "color : 42"

	recB := recaster.NewRecaster(recaster.Config{
		Name:           "Recaster A",
		PatchEnvelopes: []string{patchID},
	}, peMgr)

	transformed, err := recB.TransformString(input)
	g.Expect(err).To(gomega.BeNil())
	g.Expect(transformed).To(gomega.BeEquivalentTo(expected))
}
