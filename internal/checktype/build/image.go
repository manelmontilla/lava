// Copyright 2023 Adevinta

package build

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"time"

	checkcatalog "github.com/adevinta/vulcan-check-catalog/pkg/model"
	"github.com/docker/docker/client"

	"github.com/adevinta/lava/internal/dockerutil"
)

const (
	// lasModifiedTimeLabel defines the key of the label using [reverse DNS notation].
	//
	// [reverse DNS notation]:https://docs.docker.com/config/labels-custom-metadata/
	lastModifiedTimeLabel = "com.adevinta.vulcan.last_modified_file"

	// checktypeNameLabel defines the key of the label using [reverse DNS notation].
	//
	// [reverse DNS notation]:https://docs.docker.com/config/labels-custom-metadata/
	checktypeNameLabel = "com.adevinta.vulcan.name"

	// checktypeManifest defines the key of the label using [reverse DNS notation].
	//
	// [reverse DNS notation]:https://docs.docker.com/config/labels-custom-metadata/
	checktypeManifest = "com.adevinta.vulcan.manifest"
)

// ErrNoChecktypeImage is returned by the [ParseImage] function when an image
// does not contain the metadata of a checktype.
type ErrNoChecktypeImage struct {
	Image string
}

func (e ErrNoChecktypeImage) Error() string {
	return fmt.Sprintf("invalid metadata in image %s", e.Image)
}

// Image represents the metadata about a checktype stored in a docker image.
// Vulcan checktype.
type Image struct {
	// Name the name of the image in format REPOSITORY:TAG.
	Name string

	// ChecktypeName the name of the checktype that the image contains.
	ChecktypeName string

	// Manifest the manifest of the checktype that the image contains.
	Manifest Manifest

	// LastModified contains the time the code of the checktype stored in the
	// image was modified.
	LastModified time.Time
}

// InspectImage returns the metadata about a checktype stored in an image.
func InspectImage(cli client.APIClient, image string) (Image, error) {
	labels, err := dockerutil.ImageLabels(cli, image)
	if err != nil {
		return Image{}, fmt.Errorf("unable to read image labels: %w", err)
	}
	lastModified, ok := labels[lastModifiedTimeLabel]
	if !ok {
		err := ErrNoChecktypeImage{Image: image}
		return Image{}, fmt.Errorf("%w: label %s not found", err, lastModifiedTimeLabel)
	}
	lastModifiedTime, err := time.Parse(time.RFC822, lastModified)
	if err != nil {
		errNoCheck := ErrNoChecktypeImage{Image: image}
		err := fmt.Errorf("invalid time %s defined in the label %s: %w", lastModified, lastModifiedTimeLabel, errNoCheck)
		return Image{}, err
	}

	ctName, ok := labels[checktypeNameLabel]
	if !ok {
		err := ErrNoChecktypeImage{Image: image}
		return Image{}, fmt.Errorf("label %s not found: %w", checktypeNameLabel, err)
	}

	m, ok := labels[checktypeManifest]
	if !ok {
		err := ErrNoChecktypeImage{Image: image}
		return Image{}, fmt.Errorf("label %s not found: %w", checktypeManifest, err)
	}

	manifest, err := ParseManifest(m)
	if err != nil {
		err := ErrNoChecktypeImage{Image: image}
		return Image{}, fmt.Errorf("invalid checktype manifest: %w", err)
	}

	return Image{
		Name:          image,
		ChecktypeName: ctName,
		Manifest:      manifest,
		LastModified:  lastModifiedTime,
	}, nil
}

// NewImage builds a Docker image representing the checktype defined in the
// specified directory using the specified checktype name. Returns the metadata
// stored in the built image.
func NewImage(ctx context.Context, cli client.APIClient, name, dir string, checktype string) (Image, error) {
	manifestPath := path.Join(dir, ManifestFile)
	manifestContent, err := os.ReadFile(manifestPath)
	if os.IsNotExist(err) {
		return Image{}, fmt.Errorf("no manifest file found in the dir %s", dir)
	}
	if err != nil {
		return Image{}, err
	}
	manifest, err := ParseManifest(string(manifestContent))
	if err != nil {
		return Image{}, fmt.Errorf("invalid manifest file %s: %w", manifestPath, err)
	}

	contents, err := buildTarFromDir(dir)
	if err != nil {
		return Image{}, err
	}

	modif, err := lastModified(dir)
	if err != nil {
		return Image{}, err
	}
	t := modif.Format(time.RFC822)

	labels := map[string]string{
		checktypeNameLabel:    checktype,
		checktypeManifest:     string(manifestContent),
		lastModifiedTimeLabel: t,
	}

	_, err = dockerutil.BuildImage(ctx, cli, contents, []string{name}, labels)
	if err != nil {
		return Image{}, fmt.Errorf("unable to build image for checktype in dir %s: %w", dir, err)
	}

	image := Image{
		ChecktypeName: checktype,
		Name:          name,
		Manifest:      manifest,
	}
	return image, nil
}

// Checktype returns the information of the checktype defined in the image.
func (i Image) Checktype() (checkcatalog.Checktype, error) {
	options, err := i.Manifest.UnmarshalOptions()
	if err != nil {
		return checkcatalog.Checktype{}, fmt.Errorf("unable to unmarshal options: %w", err)
	}
	assetTypes, err := i.Manifest.AssetTypes.Strings()
	if err != nil {
		return checkcatalog.Checktype{}, fmt.Errorf("unable to read asset types: %w", err)
	}
	var requiredVars []any
	for _, r := range i.Manifest.RequiredVars {
		requiredVars = append(requiredVars, r)
	}
	ct := checkcatalog.Checktype{
		Name:         i.ChecktypeName,
		Description:  i.Manifest.Description,
		Image:        i.Name,
		Timeout:      i.Manifest.Timeout,
		Options:      options,
		RequiredVars: requiredVars,
		Assets:       assetTypes,
	}
	return ct, nil
}

func buildTarFromDir(dirPath string) (*bytes.Buffer, error) {
	dir, err := os.Open(path.Clean(dirPath))
	if err != nil {
		return nil, err
	}
	defer dir.Close() // nolint: errcheck

	files, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	tarfileWriter := tar.NewWriter(&output)
	defer tarfileWriter.Close() // nolint: errcheck

	err = addDir(dirPath, "", tarfileWriter, files)
	return &output, err
}
