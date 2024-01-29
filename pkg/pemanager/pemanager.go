package pemanager

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/uncleDecart/ics-api/go/client"
	c "github.com/uncleDecart/ics-api/go/client"
	"github.com/uncleDecart/ics-container/pkg/utils"
)

type PatchEnvelopeManager struct {
	sync.Mutex
	envelopesAvailable []c.PatchEnvelopeDescription

	data   map[string][]byte
	client *c.Client
}

type PatchEnvelopeView struct {
	Envelopes []c.PatchEnvelopeDescription
}

func NewPatchEnvelopeManager(endpoint string) (*PatchEnvelopeManager, error) {
	c, err := c.NewClient(endpoint)
	if err != nil {
		return nil, err
	}

	peMgr := &PatchEnvelopeManager{
		client: c,
		data:   make(map[string][]byte),
	}

	return peMgr, nil
}

// Needed for testing
func NewPatchEnvelopeManagerPreInit(endpoint string, data map[string][]byte, env []c.PatchEnvelopeDescription) (*PatchEnvelopeManager, error) {
	c, err := c.NewClient(endpoint)
	if err != nil {
		return nil, err
	}

	peMgr := &PatchEnvelopeManager{
		client:             c,
		data:               data,
		envelopesAvailable: env,
	}

	return peMgr, nil
}

func (mgr *PatchEnvelopeManager) Envelopes() []string {
	mgr.Lock()
	defer mgr.Unlock()

	result := make([]string, len(mgr.envelopesAvailable))

	for i := range mgr.envelopesAvailable {
		result[i] = *mgr.envelopesAvailable[i].PatchID
	}

	return result
}

func (mgr *PatchEnvelopeManager) GetBlobInfo(patchID, fileName string) *c.BinaryBlob {
	mgr.Lock()
	defer mgr.Unlock()

	patch := utils.FindPatchEnvelope(mgr.envelopesAvailable, patchID)
	if patch == nil {
		return nil
	}
	if blob := utils.FindBinaryBlob(*patch, fileName); blob != nil {
		fmt.Println(*blob.FileName)
		return blob
	}
	return nil
}

func (mgr *PatchEnvelopeManager) GetBlobData(patchID, fileName string) ([]byte, bool) {
	mgr.Lock()
	defer mgr.Unlock()

	val, ok := mgr.data[getDataKey(patchID, fileName)]
	return val, ok
}

func (mgr *PatchEnvelopeManager) View() PatchEnvelopeView {
	return PatchEnvelopeView{
		Envelopes: mgr.envelopesAvailable[:],
	}
}

func (mgr *PatchEnvelopeManager) Fetch() ([]string, error) {
	mgr.Lock()
	defer mgr.Unlock()

	changedEnvelopes, err := mgr.fetchDescription()

	if err != nil {
		return []string{}, err
	}

	if err := mgr.fetchAllBinaryBlobs(); err != nil {
		return []string{}, err
	}

	return changedEnvelopes, nil
}

func (mgr *PatchEnvelopeManager) fetchDescription() ([]string, error) {
	response, err := mgr.client.GetAvailablePatchEnvelopes(context.TODO())
	if err != nil {
		return []string{}, err
	}
	defer response.Body.Close()

	pes, err := client.ParseGetAvailablePatchEnvelopesResponse(response)
	if err != nil {
		return []string{}, err
	}

	changedEnvelopes := descDiff(mgr.envelopesAvailable, *pes.JSON200)

	mgr.envelopesAvailable = *pes.JSON200

	return changedEnvelopes, nil
}

func descDiff(desc1, desc2 []c.PatchEnvelopeDescription) []string {
	diff := make([]string, 0)

	// Create a map to store the elements of the first slice
	elements := make(map[string]struct{})
	for _, v := range desc1 {
		key := *v.PatchID + ":" + *v.Version
		elements[key] = struct{}{}
	}

	// Check each element in the second slice
	// If an element is not in the map, it is added to the diff slice
	for _, v := range desc2 {
		key := *v.PatchID + ":" + *v.Version
		if _, exists := elements[key]; !exists {
			diff = append(diff, *v.PatchID)
		}
	}

	return diff
}

func (mgr *PatchEnvelopeManager) fetchAllBinaryBlobs() error {
	for _, pe := range mgr.envelopesAvailable {
		for _, blob := range *pe.BinaryBlobs {
			if err := mgr.fetchBinaryBlob(pe, blob); err != nil {
				return err
			}
		}
	}

	return nil
}

func (mgr *PatchEnvelopeManager) fetchBinaryBlob(pe c.PatchEnvelopeDescription, blob c.BinaryBlob) error {
	resp, err := mgr.client.DownloadPatchArchiveFile(context.TODO(), *pe.PatchID, *blob.FileName)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	mgr.data[getDataKey(*pe.PatchID, *blob.FileName)] = bodyBytes
	return nil
}

func getDataKey(patchID, fileName string) string {
	return patchID + ":" + fileName
}
