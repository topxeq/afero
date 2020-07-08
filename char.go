// topxeq added code, thanks to github.com/spf13/afero

package afero

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// IfFileExists if
func (p *MemMapFs) IfFileExists(pathA string) bool {
	rs, errT := Exists(p, pathA)

	if errT != nil {
		return false
	}

	return rs
}

// LoadBytesFromFileE LoadBytes, numA < 0 indicates read all
func (p *MemMapFs) LoadBytesFromFile(fileNameA string, numA int) ([]byte, error) {
	if !p.IfFileExists(fileNameA) {
		return nil, fmt.Errorf("file not exists")
	}

	fileT, errT := p.Open(fileNameA)
	if errT != nil {
		return nil, errT
	}

	defer fileT.Close()

	if numA <= 0 {
		fileContentT, errT := ioutil.ReadAll(fileT)
		if errT != nil {
			return nil, errT
		}

		return fileContentT, nil
	}

	bufT := make([]byte, numA)

	nnT, errT := fileT.Read(bufT)
	if errT != nil {
		return nil, errT
	}

	if nnT != len(bufT) {
		return nil, fmt.Errorf("read bytes not identical")
	}

	return bufT, nil
}

func (p *MemMapFs) CopyFileFrom(src, dst string, forceA bool, bufferSizeA int) error {

	srcFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	mode := srcFileStat.Mode()

	if !mode.IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	if mode.IsDir() {
		return fmt.Errorf("%s is a folder", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}

	defer source.Close()

	if !forceA {
		_, err = p.Stat(dst)
		if err != nil {
			return fmt.Errorf("file %s already exists", dst)
		}
	}

	destination, err := p.Create(dst)
	if err != nil {
		return err
	}

	defer destination.Close()

	if bufferSizeA <= 0 {
		bufferSizeA = 1000000
	}

	buf := make([]byte, bufferSizeA)
	for {
		n, err := source.Read(buf)

		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}

	return err
}
