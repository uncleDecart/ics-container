package recaster_test

import (
	"context"
	"fmt"
	"net"
	"sort"
	"testing"

	"github.com/onsi/gomega"
	"github.com/uncleDecart/ics-container/pkg/mock"
	"github.com/uncleDecart/ics-container/pkg/pemanager"
	"github.com/uncleDecart/ics-container/pkg/recaster"

	c "github.com/uncleDecart/ics-api/go/client"
)

func TestDelete(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)

	recA := recaster.Recaster{
		Name:           "Recaster A",
		PatchEnvelopes: []string{"patch1", "patch2"},
	}
	recB := recaster.Recaster{
		Name:           "Recaster B",
		PatchEnvelopes: []string{"patch1", "patch2"},
	}

	recC := recaster.Recaster{
		Name:           "Recaster C",
		PatchEnvelopes: []string{"patch3", "patch4"},
	}

	rMgr := recaster.RecasterManager{
		Recasters: []recaster.Recaster{recA, recB, recC},
	}

	deleted := rMgr.Delete(recA)
	g.Expect(deleted).To(gomega.BeTrue())

	recD := recaster.Recaster{
		Name:           "Recaster D",
		PatchEnvelopes: []string{"patch3", "patch4"},
	}
	deleted = rMgr.Delete(recD)
	g.Expect(deleted).To(gomega.BeFalse())

	// We don't care about order after deleting
	// so in order to compare we sort resulting array
	sort.Slice(rMgr.Recasters, func(i, j int) bool {
		return rMgr.Recasters[i].Name < rMgr.Recasters[j].Name
	})

	g.Expect(rMgr.Recasters).To(gomega.BeEquivalentTo(
		[]recaster.Recaster{recB, recC},
	))
}

func TestPut(t *testing.T) {
	t.Parallel()
	g := gomega.NewGomegaWithT(t)

	dummyPeMgr, err := pemanager.NewPatchEnvelopeManager("localhost:2222")
	g.Expect(err).To(gomega.BeNil())

	recB := recaster.Recaster{
		Name:           "Recaster B",
		PatchEnvelopes: []string{"patch1", "patch2"},
	}

	recC := recaster.Recaster{
		Name:           "Recaster C",
		PatchEnvelopes: []string{"patch3", "patch4"},
	}

	rMgr := recaster.RecasterManager{
		Recasters: []recaster.Recaster{recB, recC},
		PeMgr:     dummyPeMgr,
	}

	recA := recaster.Recaster{
		Name:           "Recaster A",
		PatchEnvelopes: []string{"patch1", "patch2"},
	}

	err = rMgr.Put(recA)
	g.Expect(err).To(gomega.BeNil())

	g.Expect(rMgr.Recasters).To(gomega.BeEquivalentTo(
		[]recaster.Recaster{recB, recC, recA},
	))

	newC := recaster.Recaster{
		Name:           "Recaster C",
		PatchEnvelopes: []string{"patch3", "patch4", "patch42"},
	}

	err = rMgr.Put(newC)
	g.Expect(err).To(gomega.BeNil())

	g.Expect(rMgr.Recasters).To(gomega.BeEquivalentTo(
		[]recaster.Recaster{recB, newC, recA},
	))
}

func TestFetch(t *testing.T) {
	t.Parallel()

	g := gomega.NewGomegaWithT(t)

	l, closeFn := mock.CreateListener()
	defer closeFn()

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
		patchA + ":" + fileNameA: "YmVmb3JlQUE=",
		patchA + ":" + fileNameB: "YmVmb3JlQUI=",
		patchB + ":" + fileNameA: "YmVmb3JlQkE=",
		patchB + ":" + fileNameB: "YmVmb3JlQkI=",
	}

	server := mock.CreateMockMetadataServer(&envelopesAvailable, &data)
	go func() {
		if err := server.Serve(l); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	port := l.Addr().(*net.TCPAddr).Port
	fmt.Printf("looking at %d\n", port)
	dummyPeMgr, err := pemanager.NewPatchEnvelopeManager(fmt.Sprintf("http://127.0.0.1:%d", port))
	g.Expect(err).To(gomega.BeNil())

	templatesB := map[string]string{
		"section1": fmt.Sprintf("$(%s/%s/data)", patchA, fileNameA),
		"section2": fmt.Sprintf("$(%s/%s/data)", patchA, fileNameB),
		"section3": fmt.Sprintf("$(%s/%s/data)", patchB, fileNameA),
		"section4": fmt.Sprintf("$(%s/%s/data)", patchB, fileNameB),
	}
	recB := recaster.Recaster{
		Name:           "Recaster B",
		PatchEnvelopes: []string{patchA, patchB},
		Templates:      templatesB,
		Driver:         &recaster.DummyOutput{},
		Backup:         &recaster.NoBackup{},
	}
	g.Expect(err).To(gomega.BeNil())

	templatesC := map[string]string{
		"section1": fmt.Sprintf("$(%s/%s/fileMetaData)", patchA, fileNameA),
		"section2": fmt.Sprintf("$(%s/%s/fileSha)", patchB, fileNameB),
	}
	recC := recaster.Recaster{
		Name:           "Recaster C",
		PatchEnvelopes: []string{patchA, patchB},
		Templates:      templatesC,
		Driver:         &recaster.DummyOutput{},
		Backup:         &recaster.NoBackup{},
	}
	g.Expect(err).To(gomega.BeNil())

	rMgr := recaster.RecasterManager{
		Recasters: []recaster.Recaster{recB, recC},
		PeMgr:     dummyPeMgr,
	}
	g.Expect(err).To(gomega.BeNil())

	err = rMgr.Fetch()
	g.Expect(err).To(gomega.BeNil())

	expected := map[string]string{
		"section1": "beforeAA",
		"section2": "beforeAB",
		"section3": "beforeBA",
		"section4": "beforeBB",
	}
	statusCode, statusStr := recB.Driver.Status()
	g.Expect(statusCode).To(gomega.BeEquivalentTo(recaster.DriverStatusOK))
	g.Expect(statusStr).To(gomega.BeEquivalentTo(fmt.Sprint(expected)))

	expected = map[string]string{
		"section1": fileMetadataA,
		"section2": fileShaA,
	}
	statusCode, statusStr = recC.Driver.Status()
	g.Expect(statusCode).To(gomega.BeEquivalentTo(recaster.DriverStatusOK))
	g.Expect(statusStr).To(gomega.BeEquivalentTo(fmt.Sprint(expected)))

	err = server.Shutdown(context.TODO())
	g.Expect(err).To(gomega.BeNil())
}
