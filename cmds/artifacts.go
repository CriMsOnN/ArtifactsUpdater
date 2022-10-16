package cmds

import (
	"archive/tar"
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"github.com/xi2/xz"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Configuration struct {
	Artifacts_windows string `json:"artifacts_windows"`
	Artifacts_linux   string `json:"artifacts_linux"`
}

type Artifact struct {
	Recommended_Download string `json:"recommended_download"`
	Optional_Download    string `json:"optional_download"`
	Latest_Download      string `json:"latest_download"`
	Critical_Download    string `json:"critical_download"`
}

func GetArtifactUrl(c *Configuration, osType string, artifact string) string {
	var downloadURL string

	if osType == "win32" {
		downloadURL = c.Artifacts_windows
	} else if osType == "linux" {
		downloadURL = c.Artifacts_linux
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		panic(err)
	}

	var a Artifact
	jsonError := json.NewDecoder(resp.Body).Decode(&a)
	if jsonError != nil {
		panic(jsonError)
	}

	switch artifact {
	case "Recommended_Download":
		return a.Recommended_Download
	case "Optional_Download":
		return a.Optional_Download
	case "Latest_Download":
		return a.Latest_Download
	case "Critical_Download":
		return a.Critical_Download
	}

	return downloadURL
}

func DownloadFile(url string, artifactsType string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return
	}

	var fileName string

	if artifactsType == "win32" {
		fileName = "artifacts.zip"
	} else {
		fileName = "artifacts.tar.gz"
	}

	file, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading",
	)

	_, err = io.Copy(io.MultiWriter(file, bar), resp.Body)
	if err != nil {
		panic(err)
	}

	if artifactsType == "win32" {
		unzipFile(fileName)
	} else {
		r, err := os.Open(fileName)
		if err != nil {
			fmt.Printf("Error: %s", err.Error())
		}
		extractTar(r, fileName)
	}
}

func unzipFile(fileName string) {
	dst := "output"
	file, err := zip.OpenReader(fileName)
	if err != nil {
		panic(err)
	}

	defer file.Close()
	for _, f := range file.File {
		filePath := filepath.Join(dst, f.Name)
		fmt.Println("Extracting", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dst)+string(os.PathSeparator)) {
			panic("illegal file path")
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		rc, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			panic(err)
		}

	}
	err = os.Remove(fileName)
	if err != nil {
		fmt.Printf("Error removing file %s\n", fileName)
	}

	fmt.Println("Extracting done. Please check the output folder.")
}

// fuck me
// https://stackoverflow.com/questions/57639648/how-to-decompress-tar-gz-file-in-go
func extractTar(gzipStream io.Reader, fileName string) {
	uncompressed, err := xz.NewReader(gzipStream, 0)
	if err != nil {
		fmt.Printf("Error on Extract Tar: Reader failed %s", err.Error())
	}

	reader := tar.NewReader(uncompressed)

	for true {
		header, err := reader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Printf("Next() failed %s", err.Error())
		}

		target := filepath.Join("output", header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					fmt.Printf("MkdirAll Error: %s\n", err.Error())
					break
				}
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				fmt.Printf("OpenFile error: %s\n", err.Error())
				break
			}

			if _, err := io.Copy(f, reader); err != nil {
				fmt.Printf("Error on copy: %s\n", err.Error())
				break
			}

			fmt.Printf("Extracting %s/%s\n", "output", header.Name)
		}
	}
	err = os.Remove(fileName)
	if err != nil {
		fmt.Printf("Error when removing tar file %s\n", err.Error())
	}
	fmt.Println("Extracting done. Please check the output folder.")
}
