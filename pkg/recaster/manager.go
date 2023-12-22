package recaster

import "github.com/uncleDecart/ics-container/pkg/pemanager"

type RecasterManager struct {
	Recasters []*Recaster
	PeMgr     *pemanager.PatchEnvelopeManager
}

func (rMgr *RecasterManager) Put(cfg Config) {
	r := NewRecaster(cfg, rMgr.PeMgr)

	for idx, rec := range rMgr.Recasters {
		if rec.IsEqual(r) {
			rMgr.Recasters[idx] = r
			return
		}
	}

	rMgr.Recasters = append(rMgr.Recasters, r)
}

func (rMgr *RecasterManager) Delete(cfg Config) bool {
	r := NewRecaster(cfg, rMgr.PeMgr)

	for idx, rec := range rMgr.Recasters {
		if rec.IsEqual(r) {
			rMgr.Recasters[idx] = rMgr.Recasters[len(rMgr.Recasters)-1]
			rMgr.Recasters = rMgr.Recasters[:len(rMgr.Recasters)-1]
			return true
		}
	}
	return false
}

func (rMgr *RecasterManager) GetIdx(name string) int {
	for idx, rec := range rMgr.Recasters {
		if rec.Name == name {
			return idx
		}
	}
	return -1
}
