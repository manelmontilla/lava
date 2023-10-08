// Copyright 2023 Adevinta

package build

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/manelmontilla/toml"
)

// ManifestFile defines the name of the manifest file of a checktype.
const ManifestFile = "manifest.toml"

// AssetType defines the valid types of assets a check can accept.
type AssetType int

const (
	// IP represents an IP assettype.
	IP AssetType = iota
	// Hostname represents a hostname assettype.
	Hostname
	// DomainName represents an domain name assettype.
	DomainName
	// AWSAccount represents an AWS account assettype.
	AWSAccount
	// IPRange represents an IP range assettype.
	IPRange
	// DockerImage represents a DockerImage asset type.
	DockerImage
	// WebAddress represents a WebAddress asset type.
	WebAddress
	// GitRepository represents a git repo asset type.
	GitRepository
	// GCPProject represents a GCP Project type.
	GCPProject
)

var assetTypeStrings = map[AssetType]string{
	IP:            "IP",
	Hostname:      "Hostname",
	DomainName:    "DomainName",
	AWSAccount:    "AWSAccount",
	IPRange:       "IPRange",
	DockerImage:   "DockerImage",
	WebAddress:    "WebAddress",
	GitRepository: "GitRepository",
	GCPProject:    "GCPProject",
}

// MarshalText returns the string representation of a AssetType instance.
func (a *AssetType) MarshalText() (text []byte, err error) {
	s, err := a.String()
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// UnmarshalText creates a AssetType from its string representation.
func (a *AssetType) UnmarshalText(text []byte) error {
	val := string(text)
	for k, v := range assetTypeStrings {
		if v == val {
			*a = k
			return nil
		}
	}
	return fmt.Errorf(" %s is not a valid AssetType", val)
}

func (a *AssetType) String() (string, error) {
	if _, ok := assetTypeStrings[*a]; !ok {
		return "", fmt.Errorf("value %d is not a valid AssetType", a)
	}
	return assetTypeStrings[*a], nil
}

// AssetTypes represents an array of asset types supported by a concrete
// checktype.
type AssetTypes []*AssetType

// Strings converts a slice of Assettypes into a slice of strings.
func (a AssetTypes) Strings() ([]string, error) {
	res := []string{}
	for _, s := range a {
		txt, err := s.String()
		if err != nil {
			return nil, err
		}
		res = append(res, txt)
	}
	return res, nil
}

// Manifest contains all the data defined in the manifest.
type Manifest struct {
	Description  string
	Timeout      int
	Options      string
	RequiredVars []string
	QueueName    string
	AssetTypes   AssetTypes
}

// UnmarshalOptions returns the options interpreted as json.
func (m Manifest) UnmarshalOptions() (map[string]interface{}, error) {
	if m.Options == "" {
		return nil, nil
	}
	var options = make(map[string]interface{})
	err := json.Unmarshal([]byte(m.Options), &options)
	if err != nil {
		return nil, err
	}
	return options, nil
}

// ParseManifestFile reads a manifest file.
func ParseManifestFile(path string) (Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("unable to open manifest file %s: %w", path, err)
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return Manifest{}, fmt.Errorf("unable to read manifest file %s: %w", path, err)
	}
	return ParseManifest(string(content))
}

// ParseManifest parses the contents of a manifest file.
func ParseManifest(content string) (Manifest, error) {
	d := Manifest{}
	m, err := toml.Decode(content, &d)
	if err != nil {
		return d, err
	}
	if !m.IsDefined("Description") {
		return d, errors.New("description field is mandatory")
	}

	if m.IsDefined("Options") {
		dummy := make(map[string]any)
		err = json.Unmarshal([]byte(d.Options), &dummy)
		if err != nil {
			err = fmt.Errorf("reading manifest file, the options field is not a valid json: %v", err)
			return d, err
		}
	}
	return d, nil
}
