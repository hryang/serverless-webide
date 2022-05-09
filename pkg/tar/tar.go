package tar

// Reference: https://github.com/mimoo/eureka/blob/master/folders.go

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	//gzip "github.com/klauspost/pgzip"
	"compress/gzip"
)

// Check for path traversal and correct forward slashes
func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}

// Extract the tar.gz stream data and write to the local file.
// src: the source of the tar.gz stream
// dst: the destination of the local file.
func ExtractTarGz(src io.Reader, dst string) error {
	uncompressedStream, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("ExtractTarGz: new gzip reader failed: %v", err)
	}

	tarReader := tar.NewReader(uncompressedStream)

	for {
		header, err := tarReader.Next()
		// Reach the end of the stream.
		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("ExtractTarGz: new tar reader failed: %v", err)
		}

		if !validRelPath(header.Name) {
			return fmt.Errorf("ExtractTarGz: tar containerd invalid name: %s", header.Name)
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		// If it's a directory and does not exist, then create it with 0755 permission.
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return fmt.Errorf("ExtractTarGz: Mkdir() failed: %v", err)
				}
			}
		// If it's a file, create it with same permission.
		case tar.TypeReg:
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("ExtractTarGz: OpenFile() failed: %v", err)
			}
			if _, err := io.Copy(fileToWrite, tarReader); err != nil {
				return fmt.Errorf("ExtractTarGz: write file failed: %v", err)
			}

			// Manually close here after each file operation. defering would cause each file
			// close to wait until all operations have completed.
			fileToWrite.Close()
		}
	}

	return nil
}

// Compress a file or directory as tar.gz and write to the destination io stream.
// src: the source of the file or directory.
// dst: the destination of the io stream.
func TarGz(src string, dst io.Writer) error {
	gzipWriter := gzip.NewWriter(dst)
	tarWriter := tar.NewWriter(gzipWriter)

	fi, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("TarGz: stat %s failed: %v", src, err)
	}
	mode := fi.Mode()
	if mode.IsRegular() { // handle regular file
		header, err := tar.FileInfoHeader(fi, src)
		if err != nil {
			return fmt.Errorf("TarGz: get %s file info failed: %v", src, err)
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("TarGz: write file header failed: %v", err)
		}
		data, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("TarGz: open file %s failed: %v", src, err)
		}
		if _, err := io.Copy(tarWriter, data); err != nil {
			return fmt.Errorf("TarGz: write tar failed: %v", err)
		}
	} else if mode.IsDir() { // handle directory
		filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
			// Generate the tar header.
			header, err := tar.FileInfoHeader(info, path)
			if err != nil {
				return fmt.Errorf("TarGz: get %s file info header failed: %v", path, err)
			}

			header.Name, err = filepath.Rel(src, path)
			if err != nil {
				return fmt.Errorf("TarGz: get relative path failed. base path: %s target path: %s error: %v", src, path, err)
			}

			// Write tar header.
			if err := tarWriter.WriteHeader(header); err != nil {
				return fmt.Errorf("TarGz: write tar header failed: %v", err)
			}

			// Write regular file.
			if !info.IsDir() {
				data, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("TarGz: open %s file failed: %v", path, err)
				}
				if _, err := io.Copy(tarWriter, data); err != nil {
					return fmt.Errorf("TarGz: write tar stream failed: %v", err)
				}
			}

			return nil
		})
	} else {
		return fmt.Errorf("TarGz: file type not supported: %s", mode.String())
	}

	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("TarGz: close tar writer failed: %v", err)
	}

	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("TarGz: close gzip writer failed: %v", err)
	}

	return nil
}
