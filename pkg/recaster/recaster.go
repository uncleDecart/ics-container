package recaster

import (
	"encoding/json"
	"fmt"
	"regexp"

	c "github.com/uncleDecart/ics-api/go/client"
	"github.com/uncleDecart/ics-container/pkg/pemanager"
)

// example:
// $(6ba7b810-9dad-11d1-80b4-00c04fd430c8/BinaryBlobName/fieldname)
// groups are needed to extract: patchID, binaryBlob name and field name
var /* const */ TransformerPattern = regexp.MustCompile("\\$\\(([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})/(.*)/(.*)\\)")

type Config struct {
	Name           string
	PatchEnvelopes []string
	Templates      map[string]string

	DriverType   string
	DriverConfig json.RawMessage

	BackupType   string
	BackupConfig json.RawMessage
}

type Recaster struct {
	Name           string   `json:"name"`
	PatchEnvelopes []string `json:"envelopes"`
	Templates      map[string]string

	Driver OutputDriver
	Backup BackupStrategy
}

func (r *Recaster) UnmarshalJSON(data []byte) error {
	var cfg Config
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return err
	}
	r.Name = cfg.Name
	r.PatchEnvelopes = cfg.PatchEnvelopes
	r.Templates = cfg.Templates

	var drv OutputDriver
	switch cfg.DriverType {
	case DriverTypeDummy:
		drv = &DummyOutput{}
	case DriverTypeHTTP:
		drv = &HTTPOutput{}
		err := json.Unmarshal(cfg.DriverConfig, drv)
		if err != nil {
			return err
		}
	}
	r.Driver = drv

	var backup BackupStrategy
	switch cfg.BackupType {
	case BackupTypeNoBackup:
		backup = &NoBackup{}
	case BackupTypeHTTP:
		backup = &HTTPBackup{}
		err := json.Unmarshal(cfg.BackupConfig, backup)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unkown backup")
	}
	r.Backup = backup

	return nil
}

func (r *Recaster) MarshalJSON() ([]byte, error) {
	driverCfg, err := json.Marshal(r.Driver)
	if err != nil {
		return nil, err
	}
	backupCfg, err := json.Marshal(r.Backup)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&Config{
		Name:           r.Name,
		PatchEnvelopes: r.PatchEnvelopes,
		Templates:      r.Templates,

		DriverType:   r.Driver.Type(),
		DriverConfig: driverCfg,

		BackupType:   r.Backup.Type(),
		BackupConfig: backupCfg,
	})
}

func (r *Recaster) IsEqual(rhs Recaster) bool {
	return r.Name == rhs.Name
}

func (r *Recaster) Recast(mgr *pemanager.PatchEnvelopeManager) error {
	if mgr == nil {
		return fmt.Errorf("PatchEnvelopeManager is nil")
	}
	transformed := make(map[string]string)

	if r.Backup.Enabled() {
		for k, v := range r.Backup.Params() {
			t, err := r.TransformString(v, mgr)
			if err != nil {
				return err
			}
			t = r.transformBackupData(t, r.Backup.Data())
			transformed[k] = t
		}

		if err := r.Backup.Load(transformed); err != nil {
			return err
		}
	}

	transformedTemplates := make(map[string]string)

	for k, v := range r.Templates {
		transformed, err := r.TransformString(v, mgr)
		if err != nil {
			return err
		}
		transformedTemplates[k] = transformed
	}

	err := r.Driver.Execute(transformedTemplates)

	if err != nil {
		if r.Backup.Enabled() {
			return r.Backup.Restore(transformed)
		}
		return err
	}

	return nil
}

func (r *Recaster) TransformString(str string, mgr *pemanager.PatchEnvelopeManager) (string, error) {
	errString := ""
	result := TransformerPattern.ReplaceAllStringFunc(str, func(match string) string {
		replacement, err := r.transformSinglePath(match, mgr)
		if err != nil {
			errString = fmt.Sprintf("%s\n%v", errString, err)
			fmt.Printf("Error parsing string %v\n", err)
			return match
		}
		return replacement
	})

	if errString != "" {
		return "", fmt.Errorf("Failed to RecastString %s", errString)
	}

	return result, nil
}

// example:
// $(Backup data)
var /* const */ backupPattern = regexp.MustCompile("\\$\\(Backup data\\)")

func (r *Recaster) transformBackupData(path string, data string) string {
	return backupPattern.ReplaceAllString(path, data)
}

func (r *Recaster) transformSinglePath(path string, mgr *pemanager.PatchEnvelopeManager) (string, error) {
	keys := TransformerPattern.FindStringSubmatch(path)

	if len(keys) != 4 {
		return "", fmt.Errorf("INVALID")
	}

	patchID, blobFileName, fieldName := keys[1], keys[2], keys[3]
	fmt.Printf("Parsing %s %s %s \n", patchID, blobFileName, fieldName)

	patchFound := false
	for _, p := range r.PatchEnvelopes {
		if p == patchID {
			patchFound = true
			break
		}
	}
	if !patchFound {
		return "", fmt.Errorf("INVALID PatchID")
	}

	if fieldName == "data" {
		data, loaded := mgr.GetBlobData(patchID, blobFileName)
		if !loaded {
			return "", fmt.Errorf("FAILED FETCHIND DATA")
		}
		return string(data), nil
	}

	blob := mgr.GetBlobInfo(patchID, blobFileName)
	if blob == nil {
		return "", fmt.Errorf("FAILED TO FETCH BLOB")
	}

	f, err := getFieldByName(*blob, fieldName)
	if err != nil {
		return "", fmt.Errorf("INVALID FIELD")
	}

	return f, nil
}

func getFieldByName(b c.BinaryBlob, name string) (string, error) {
	switch name {
	case "artifactMetadata":
		return *b.ArtifactMetaData, nil
	case "fileMetaData":
		return *b.FileMetaData, nil
	case "fileName":
		return *b.FileName, nil
	case "fileSha":
		return *b.FileSha, nil
	case "size":
		return fmt.Sprintf("%d", *b.Size), nil
	case "url":
		return *b.Url, nil
	}
	return "", fmt.Errorf("INVALID FIELD NAME")
}
