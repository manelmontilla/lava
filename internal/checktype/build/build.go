// Copyright 2023 Adevinta

// Package build exposes the types and functions needed for building a
// checktype from code.
package build

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	checkcatalog "github.com/adevinta/vulcan-check-catalog/pkg/model"

	"github.com/adevinta/lava/internal/containers"
)

// Code represents a dir containing the definition of a checktype.
type Code string

// Build builds the code of a checktype defined in a directory. If the code was
// not modified since the last time it was build locally, it doesn't rebuild
// the check. Returns the data representing the checktype.
func (c Code) Build(ctx context.Context, rt containers.Runtime) (checkcatalog.Checktype, error) {
	bLog := slog.Default().With("directory", c)
	cli, err := containers.NewDockerdClient(rt)
	if err != nil {
		return checkcatalog.Checktype{}, fmt.Errorf("unable to get Docker client: %w", err)
	}

	modified, err := c.isModified(bLog, cli)
	if err != nil {
		return checkcatalog.Checktype{}, err
	}
	if !modified {
		bLog.Info("no changes in checktype, reusing image", "image", c.imageName())
		image, err := InspectImage(cli, c.imageName())
		if err != nil {
			return checkcatalog.Checktype{}, err
		}
		return image.Checktype()
	}
	dir := string(c)
	bLog.Info("compiling checktype")
	// Run go build in the checktype dir.
	if err := goBuildDir(dir); err != nil {
		return checkcatalog.Checktype{}, err
	}
	// Build a tar file with the docker image contents.
	bLog.Info("building image for checktype")

	image, err := NewImage(ctx, cli, c.imageName(), string(c), c.name())
	if err != nil {
		return checkcatalog.Checktype{}, err
	}
	return image.Checktype()
}

func (c Code) isModified(logger *slog.Logger, cli containers.DockerdClient) (bool, error) {
	logger = logger.With("image", c)
	image, err := InspectImage(cli, c.imageName())

	noCheckImageErr := ErrNoChecktypeImage{}
	if errors.As(err, &noCheckImageErr) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	dirTime, err := lastModified(string(c))
	if err != nil {
		err := fmt.Errorf("error: %+w, getting the last modification time for the checktype in %s", err, string(c))
		return false, err
	}
	logger.Debug("checking if the code of the checktype was modified", "image-modified.time", image.LastModified, "dir-modified-time", dirTime)
	modified := dirTime.Equal(image.LastModified)
	return modified, nil
}

func (c Code) imageName() string {
	name := c.name()
	return fmt.Sprintf("%s-%s", name, "local")
}

// name returns the name of the checktype represented in the code directory.
func (c Code) name() string {
	dir := string(c)
	return path.Base(dir)
}

func goBuildDir(dir string) error {
	args := []string{"build", "-a", "-ldflags", "-extldflags -static", "."}
	cmd := exec.Command("go", args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOOS=linux", "CGO_ENABLED=0")
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func addDir(sourceDir string, currentPath string, writer *tar.Writer, finfo []os.FileInfo) error {
	for _, file := range finfo {
		tarPath := path.Join(currentPath, file.Name())
		// If file is a dir we recurse.
		if file.IsDir() {
			absPath := path.Join(sourceDir, tarPath)
			dir, err := os.Open(absPath)
			if err != nil {
				return err
			}

			files, err := dir.Readdir(0)
			if err != nil {
				return err
			}

			err = addDir(sourceDir, tarPath, writer, files)
			if err != nil {
				return err
			}
			continue
		}
		// File is not a dir, add to the the Tar.
		h, err := tar.FileInfoHeader(file, tarPath)
		if err != nil {
			return err
		}

		h.Name = tarPath
		if err = writer.WriteHeader(h); err != nil {
			return err
		}

		absFilePath := path.Join(sourceDir, tarPath)

		var content []byte
		content, err = os.ReadFile(absFilePath)
		if err != nil {
			return err
		}

		if _, err = writer.Write(content); err != nil {
			return err
		}
	}
	return nil
}

// lastModified returns the newest last modified time of all the files in the tree
// rooted at the specified dir.
func lastModified(dir string) (time.Time, error) {
	var latest *time.Time
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		modtime := info.ModTime()
		if latest == nil || modtime.After(*latest) {
			latest = &modtime
		}
		return nil
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("error walking through the dir %s", dir)
	}
	if latest == nil {
		return time.Time{}, fmt.Errorf("the dir %s is empty", dir)
	}
	return *latest, nil
}
