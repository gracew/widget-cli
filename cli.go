package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "push",
				Usage: "Uploads the current directory to Widget",
				Action: func(c *cli.Context) error {
					apiID := c.Args().Get(0)
					return TarAndUpload(apiID)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func TarAndUpload(apiID string) error {
	file, err := ioutil.TempFile(os.TempDir(), "tar-")
	if err != nil {
		return errors.Wrap(err, "failed to create temporary tar file")
	}

	err = Tar("./", file)
	if err != nil {
		return errors.Wrap(err, "failed to tar current directory")
	}

	err = Upload(apiID, file.Name())
	if err != nil {
		return errors.Wrap(err, "failed to upload tar file")
	}

	return nil
}

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
// adapted from https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func Tar(src string, writers ...io.Writer) error {
	mw := io.MultiWriter(writers...)
	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		if excludeDir(file) {
			return filepath.SkipDir
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})
}

var ExcludedDirs = []string{".git", "node_modules"}

func excludeDir(dir string) bool {
	base := path.Base(dir)
	for _, d := range ExcludedDirs {
		if base == d {
			return true
		}
	}
	return false
}

func Upload(apiID string, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return errors.Wrap(err, "failed to open file "+filename)
	}
	defer file.Close()

	res, err := http.Post(fmt.Sprintf("http://localhost:8080/apis/%s/upload", apiID), "binary/octet-stream", file)
	if err != nil {
		return errors.Wrap(err, "failed to upload file")
	}

	if res.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("upload resulted in status code %d", res.StatusCode))
	}
	return nil
}
