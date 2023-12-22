package recaster

import (
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

	Templates map[string]string

	Driver OutputDriver
}

type Recaster struct {
	Config

	mgr *pemanager.PatchEnvelopeManager
}

func NewRecaster(cfg Config, mgr *pemanager.PatchEnvelopeManager) *Recaster {
	result := &Recaster{cfg, mgr}
	return result
}

func (r *Recaster) IsEqual(rhs *Recaster) bool {
	return r.Name == rhs.Name
}

func (r *Recaster) Recast() error {
	transformedTemplates := make(map[string]string)

	for k, v := range r.Templates {
		transformed, err := r.TransformString(v)
		if err != nil {
			return err
		}
		transformedTemplates[k] = transformed
	}

	r.Driver.Execute(transformedTemplates)

	return nil
}

func (r *Recaster) TransformString(str string) (string, error) {
	errString := ""
	result := TransformerPattern.ReplaceAllStringFunc(str, func(match string) string {
		replacement, err := r.transformSinglePath(match)
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

func (r *Recaster) transformSinglePath(path string) (string, error) {
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
		data, loaded := r.mgr.GetBlobData(patchID, blobFileName)
		if !loaded {
			return "", fmt.Errorf("FAILED FETCHIND DATA")
		}
		return string(data), nil
	}

	blob := r.mgr.GetBlobInfo(patchID, blobFileName)
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
