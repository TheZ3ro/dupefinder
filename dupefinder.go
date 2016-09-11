package dupefinder

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

type fileHash struct {
	Hash     string
	Filename string
}

const header = `# This is a dupefinder catalog
#
# See https://github.com/TheZ3ro/dupefinder for more info

`

// Catalog of hash to filename mappings
type DupeCatalog map[string]string

type DeletePair struct {
	Origin   string
	Filename string
}

// Generate a catalog file based on a set of folders
func Generate(folders ...string) (int, DupeCatalog, error) {
	err := validateFolders(folders...)
	if err != nil {
		return 0, nil, err
	}

	errs := make(chan error)
	filenames := make(chan string, 100)
	entries := make(chan fileHash, 100)
	result := DupeCatalog{}

	go walkAllFolders(errs, filenames, folders...)
	go hashFiles(errs, filenames, entries)
	count := 0

	for {
		entry, ok := <-entries
		if !ok {
			break
		}

		count++
		result[entry.Hash] = entry.Filename
	}

	select {
	case err := <-errs:
		if err != nil {
			return 0, nil, err
		}
	default:
	}

	return count, result, nil
}

// Detect duplicates. Set echo to true to print duplicates, rm to delete them.
func Detect(catalog DupeCatalog, uniq bool, folders ...string) (int64, []DeletePair, error) {
	err := validateFolders(folders...)
	if err != nil {
		return 0, nil, err
	}

	errs := make(chan error)
	filenames := make(chan string, 100)
	entries := make(chan fileHash, 100)

	go walkAllFolders(errs, filenames, folders...)
	go hashFiles(errs, filenames, entries)

	delete := []DeletePair(nil)
	deleted := int64(0)
	for {
		entry, ok := <-entries
		if !ok {
			break
		}

		if orig, ok := catalog[entry.Hash]; ok {
			fi, err := os.Stat(entry.Filename)
			if err != nil {
				return 0, nil, err
			}

			del := true
			if uniq {
				if entry.Filename == orig {
					del = false
				}
			}

			if del {
				deleted += fi.Size()
				delPair := DeletePair{orig, entry.Filename}
				delete = append(delete, delPair)
			}
		}
	}

	select {
	case err := <-errs:
		if err != nil {
			return 0, nil, err
		}
	default:
	}

	return deleted, delete, nil
}

// Write catalog map on file
func WriteCatalog(catalog string, dupec DupeCatalog) error {
	out, err := os.Create(catalog)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.WriteString(header)
	if err != nil {
		return err
	}

	for hash, filename := range dupec {
		_, err := out.WriteString(fmt.Sprintf("%s %s\n", hash, filename))
		if err != nil {
			return err
		}
	}
	return nil
}

func validateFolders(folders ...string) error {
	for _, f := range folders {
		isfolder, err := isFolder(f)
		if err != nil {
			return err
		}
		if !isfolder {
			return fmt.Errorf("%s is not a folder", f)
		}
	}

	return nil
}

func isFolder(filename string) (bool, error) {
	f, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false, err
	}

	return fi.IsDir(), nil
}

func walkAllFolders(errs chan error, filenames chan string, folders ...string) {
	defer close(filenames)

	for _, f := range folders {
		err := walkFolder(f, filenames)
		if err != nil {
			errs <- err
			return
		}
	}
}

func walkFolder(filename string, out chan string) error {
	fi, err := ioutil.ReadDir(filename)
	if err != nil {
		return err
	}

	for _, child := range fi {
		fullname := path.Join(filename, child.Name())
		if child.IsDir() {
			err := walkFolder(fullname, out)
			if err != nil {
				return err
			}
		} else if child.Mode().IsRegular() {
			out <- fullname
		}
	}

	return nil
}

func hashFiles(errs chan error, filenames chan string, entries chan fileHash) {
	defer close(entries)

	var wg sync.WaitGroup

	wg.Add(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go func() {
			defer wg.Done()
			for {
				filename, ok := <-filenames
				if !ok {
					return
				}

				hash, err := hashFile(filename)
				if err != nil {
					errs <- err
					return
				}

				entries <- fileHash{
					Hash:     hash,
					Filename: filename,
				}
			}
		}()
	}

	wg.Wait()
}

func hashFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum([]byte{})), nil
}

// Parse the catalog file at filename
func ParseCatalog(filename string) (DupeCatalog, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return ParseCatalogReader(file)
}

// Parse a catalog file using an io.Reader
func ParseCatalogReader(reader io.Reader) (DupeCatalog, error) {
	result := DupeCatalog{}

	bufreader := bufio.NewReader(reader)

	done := false
	for !done {
		line, err := bufreader.ReadString('\n')
		if err == io.EOF {
			done = true
		} else if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("Malformed line: %#v", line)
		}

		result[parts[0]] = parts[1]
	}

	return result, nil
}
