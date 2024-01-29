package recaster

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/uncleDecart/ics-container/pkg/pemanager"
)

type RecasterManager struct {
	Recasters   []Recaster `json:"recasters"`
	PeServerURL string     `json:"pe-server"`
	RefreshRate uint       `json:"refresh-rate"`

	ConfigPath string                          `json:"-"`
	PeMgr      *pemanager.PatchEnvelopeManager `json:"-"`
}

func (rMgr *RecasterManager) Put(r Recaster) error {
	defer func() {
		rMgr.saveToFile()
		r.Recast(rMgr.PeMgr)
	}()

	for idx, rec := range rMgr.Recasters {
		if rec.IsEqual(r) {
			rMgr.Recasters[idx] = r
			return nil
		}
	}

	rMgr.Recasters = append(rMgr.Recasters, r)

	return nil
}

func (rMgr *RecasterManager) Delete(r Recaster) bool {
	for idx, rec := range rMgr.Recasters {
		if rec.IsEqual(r) {
			rMgr.Recasters[idx] = rMgr.Recasters[len(rMgr.Recasters)-1]
			rMgr.Recasters = rMgr.Recasters[:len(rMgr.Recasters)-1]
			return rMgr.saveToFile() == nil
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

type RecasterStatus struct {
	Name    string
	Code    DriverStatus
	Message string
}

func (rMgr *RecasterManager) Status() []RecasterStatus {
	result := make([]RecasterStatus, len(rMgr.Recasters))

	for idx, rec := range rMgr.Recasters {
		code, msg := rec.Driver.Status()
		result[idx] = RecasterStatus{
			Name:    rec.Name,
			Code:    code,
			Message: msg,
		}
	}

	return result
}

func (rMgr *RecasterManager) Fetch() error {
	if rMgr.PeMgr == nil {
		return fmt.Errorf("PE Manager is nil")
	}

	changedEnv, err := rMgr.PeMgr.Fetch()
	if err != nil {
		return err
	}

	for _, patchID := range changedEnv {
		for _, rec := range rMgr.Recasters {
			for _, pe := range rec.PatchEnvelopes {
				if pe == patchID {
					if err = rec.Recast(rMgr.PeMgr); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (rMgr *RecasterManager) Update() error {
	for _, rec := range rMgr.Recasters {
		if err := rec.Recast(rMgr.PeMgr); err != nil {
			return err
		}
	}

	return nil
}

func (rm *RecasterManager) saveToFile() error {
	if rm.ConfigPath == "" {
		return nil
	}

	updatedJSON, err := json.MarshalIndent(rm, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(rm.ConfigPath, updatedJSON, 0644)
}
