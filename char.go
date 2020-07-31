// topxeq added code, thanks to github.com/spf13/afero

package afero

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// NewMemFS create a new memory file system object
func NewMemFS() *MemMapFs {
	return &MemMapFs{}
}

// IfFileExists if
func (p *MemMapFs) IfFileExists(pathA string) bool {
	rs, errT := Exists(p, pathA)

	if errT != nil {
		return false
	}

	return rs
}

func (p *MemMapFs) IsDir(pathA string) bool {
	rs, errT := IsDir(p, pathA)

	if errT != nil {
		return false
	}

	return rs
}

func (p *MemMapFs) IsFile(fileNameA string) bool {
	f, errT := p.Open(fileNameA)
	if errT != nil {
		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false
	}

	if mode := fi.Mode(); mode.IsRegular() {
		return true
	}

	return false
}

func (p *MemMapFs) RemoveFile(filePathT string) error {
	if p.IsDir(filePathT) {
		return fmt.Errorf("%v is a directory", filePathT)
	}

	errT := p.Remove(filePathT)

	if errT != nil {
		return errT
	}

	if p.IfFileExists(filePathT) {
		return fmt.Errorf("failed to remove file: %v", filePathT)
	}

	return nil
}

// LoadBytesFromFileE LoadBytes, numA omitted or numA[0] < 0 indicates read all
func (p *MemMapFs) LoadBytesFromFile(fileNameA string, numA ...int) ([]byte, error) {
	if !p.IfFileExists(fileNameA) {
		return nil, fmt.Errorf("file not exists")
	}

	fileT, errT := p.Open(fileNameA)
	if errT != nil {
		return nil, errT
	}

	defer fileT.Close()

	if numA == nil || len(numA) < 1 || numA[0] <= 0 {
		fileContentT, errT := ioutil.ReadAll(fileT)
		if errT != nil {
			return nil, errT
		}

		return fileContentT, nil
	}

	bufT := make([]byte, numA[0])

	nnT, errT := fileT.Read(bufT)
	if errT != nil {
		return nil, errT
	}

	if nnT != len(bufT) {
		return nil, fmt.Errorf("read bytes not identical")
	}

	return bufT, nil
}

func (p *MemMapFs) LoadStringFromFile(fileNameA string) (string, error) {
	if !p.IfFileExists(fileNameA) {
		return "", fmt.Errorf("file not exists")
	}

	fileT, err := p.Open(fileNameA)
	if err != nil {
		return "", err
	}

	defer fileT.Close()

	fileContentT, err := ioutil.ReadAll(fileT)
	if err != nil {
		return "", err
	}

	return string(fileContentT), nil
}

func (p *MemMapFs) SaveStringToFile(strA string, fileA string) error {
	file, err := p.Create(fileA)
	if err != nil {
		return err
	}

	defer file.Close()

	wFile := bufio.NewWriter(file)
	wFile.WriteString(strA)
	wFile.Flush()

	return nil
}

func splitLines(strA string) []string {
	if !strings.Contains(strA, "\n") {
		if strings.Contains(strA, "\r") {
			return strings.Split(strA, "\r")
		}
	}
	strT := strings.ReplaceAll(strA, "\r", "")
	return strings.Split(strT, "\n")
}

func (p *MemMapFs) LoadStringListFromFile(fileNameA string) ([]string, error) {
	if !p.IfFileExists(fileNameA) {
		return nil, fmt.Errorf("file not exists")
	}

	fileT, err := p.Open(fileNameA)
	if err != nil {
		return nil, err
	}

	defer fileT.Close()

	fileContentT, err := ioutil.ReadAll(fileT)
	if err != nil {
		return nil, err
	}

	stringList := splitLines(string(fileContentT))

	return stringList, nil
}

func (p *MemMapFs) SaveStringListToFile(strListA []string, fileA string, sepA string) error {
	if strListA == nil {
		return fmt.Errorf("invalid parameter")
	}

	if strListA == nil {
		return fmt.Errorf("empty list")
	}

	lenT := len(strListA)

	var errT error

	file, errT := p.Create(fileA)
	if errT != nil {
		return errT
	}

	defer file.Close()

	wFile := bufio.NewWriter(file)

	for i := 0; i < lenT; i++ {
		_, errT = wFile.WriteString(strListA[i])
		if errT != nil {
			return errT
		}

		if i != (lenT - 1) {
			_, errT = wFile.WriteString(sepA)
			if errT != nil {
				return errT
			}
		}
	}

	wFile.Flush()

	return nil
}

func (p *MemMapFs) AppendStringToFile(strA string, fileA string) error {
	fileT, errT := p.OpenFile(fileA, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if errT != nil {
		return errT
	}

	writerT := bufio.NewWriter(fileT)

	writerT.WriteString(strA)

	writerT.Flush()

	defer fileT.Close()

	return nil
}

func ifSwitchExistsWhole(argsA []string, switchStrA string) bool {
	if argsA == nil {
		return false
	}

	for _, argT := range argsA {
		if argT == switchStrA {
			return true
		}

	}

	return false
}

func getSwitch(argsA []string, switchStrA string, defaultA string) string {
	if argsA == nil {
		return defaultA
	}

	tmpStrT := ""
	for _, argT := range argsA {
		if strings.HasPrefix(argT, switchStrA) {
			tmpStrT = argT[len(switchStrA):]
			if strings.HasPrefix(tmpStrT, "\"") && strings.HasSuffix(tmpStrT, "\"") {
				return tmpStrT[1 : len(tmpStrT)-1]
			}

			return tmpStrT
		}

	}

	return defaultA

}

func strToInt(strA string, defaultA int) int {
	nT, errT := strconv.ParseInt(strA, 10, 0)
	if errT != nil {
		return defaultA
	}

	return int(nT)
}

func getIntSwitch(argsA []string, switchStrA string, defaultA int) int {
	tmpStrT := getSwitch(argsA, switchStrA, string(defaultA))

	return strToInt(tmpStrT, defaultA)
}

// CopyFileFrom usage: CopyFileFrom("a.txt", "b.txt", "-force", "-buffer=100000")
// the last 2 parameters are optional
func (p *MemMapFs) CopyFileFrom(src, dst string, optionsA ...string) error {

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

	if !ifSwitchExistsWhole(optionsA, "-force") {
		_, err = p.Stat(dst)
		if err != nil && os.IsExist(err) {
			return fmt.Errorf("file %s already exists", dst)
		}
	}

	destination, err := p.Create(dst)
	if err != nil {
		return err
	}

	defer destination.Close()

	bufferSizeT := getIntSwitch(optionsA, "-buffer=", 1000000)

	buf := make([]byte, bufferSizeT)
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

// CopyFileTo usage: CopyFileTo("a.txt", "b.txt", "-force", "-buffer=100000")
// the last 2 parameters are optional
func (p *MemMapFs) CopyFileTo(src, dst string, optionsA ...string) error {

	srcFileStat, err := p.Stat(src)
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

	source, err := p.Open(src)
	if err != nil {
		return err
	}

	defer source.Close()

	if !ifSwitchExistsWhole(optionsA, "-force") {
		_, err = os.Stat(dst)
		if err != nil && os.IsExist(err) {
			return fmt.Errorf("file %s already exists", dst)
		}
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer destination.Close()

	bufferSizeT := getIntSwitch(optionsA, "-buffer=", 1000000)

	buf := make([]byte, bufferSizeT)
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

// CopyFile usage: CopyFile("a.txt", "b.txt", "-force", "-buffer=100000")
// the last 2 parameters are optional
func (p *MemMapFs) CopyFile(src, dst string, optionsA ...string) error {

	srcFileStat, err := p.Stat(src)
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

	source, err := p.Open(src)
	if err != nil {
		return err
	}

	defer source.Close()

	if !ifSwitchExistsWhole(optionsA, "-force") {
		_, err = p.Stat(dst)
		if err != nil && os.IsExist(err) {
			return fmt.Errorf("file %s already exists", dst)
		}
	}

	destination, err := p.Create(dst)
	if err != nil {
		return err
	}

	defer destination.Close()

	bufferSizeT := getIntSwitch(optionsA, "-buffer=", 1000000)

	buf := make([]byte, bufferSizeT)
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

func ifOSFileExists(pathA string) bool {
	_, err := os.Stat(pathA)
	return err == nil || os.IsExist(err)
}

func isOSDir(dirNameA string) bool {
	f, err := os.Open(dirNameA)
	if err != nil {
		return false
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false
	}

	if mode := fi.Mode(); mode.IsDir() {
		return true
	}

	return false
}

func ensureMakeOSDirs(pathA string) error {
	if !ifOSFileExists(pathA) {
		os.MkdirAll(pathA, 0777)

		if !isOSDir(pathA) {
			return fmt.Errorf("failed to make directory")
		}
		return nil
	} else {
		if isOSDir(pathA) {
			return nil
		} else {
			return fmt.Errorf("a file with same name exists")
		}
	}
}

func (p *MemMapFs) CopyFilesTo(srcDirA string, patternA string, destDirA string, optionsA ...string) error {
	filesT := p.GenerateFileListRecursivelyWithExclusive(srcDirA, patternA, "", false)

	for _, v := range filesT {
		relPathT, errT := p.Rel(srcDirA, p.Dir(v))
		if errT != nil {
			return errT
		}

		destPathT := filepath.Join(destDirA, relPathT)

		errT = ensureMakeOSDirs(destPathT)

		if errT != nil {
			return errT
		}

		destFilePathT := filepath.Join(destPathT, filepath.Base(v))

		errT = p.CopyFileTo(v, destFilePathT, optionsA...)
		if errT != nil {
			return errT
		}

	}

	return nil

}

func (p *MemMapFs) EnsureMakeDirs(pathA string) error {
	if !p.IfFileExists(pathA) {
		p.MkdirAll(pathA, 0777)

		if !p.IfFileExists(pathA) {
			return fmt.Errorf("failed to make directory")
		}
		return nil
	} else {
		if p.IsDir(pathA) {
			return nil
		} else {
			return fmt.Errorf("a file with same name exists")
		}
	}
}

func (p *MemMapFs) Dir(pathA string) string {
	return filepath.Dir(pathA)
}

func (p *MemMapFs) Rel(baseA, pathA string) (string, error) {
	return filepath.Rel(baseA, pathA)
}

func (p *MemMapFs) Abs(pathA string) string {
	pathT, errT := filepath.Abs(pathA)

	if errT != nil {
		return filepath.ToSlash(pathA)
	}

	volT := filepath.VolumeName(pathT)

	if len(volT) < 1 {
		return filepath.ToSlash(pathT)
	}

	if strings.HasPrefix(pathT, volT) {
		return filepath.ToSlash(pathT[len(volT):])

	}

	return filepath.ToSlash(pathT)

}

func (p *MemMapFs) Join(elem ...string) string {
	return filepath.ToSlash(filepath.Join(elem...))
}

func (p *MemMapFs) Glob(patternA string) (matches []string, err error) {
	return Glob(p, patternA)
}

func (p *MemMapFs) GenerateFileListInDir(dirA string, patternA string, verboseA bool) []string {
	strListT := make([]string, 0, 100)

	pathT := dirA

	errT := Walk(p, pathT, func(path string, f os.FileInfo, err error) error {
		if verboseA {
			fmt.Println(path)
		}

		if f == nil {
			return err
		}

		// fmt.Printf("pathT: %v -> path: %v\n, %v", pathT, path, f.IsDir())

		// if f.IsDir() { // && path != "." && path != pathT {
		if f.IsDir() {
			if path != "." && path != pathT {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		matchedT, errTI := filepath.Match(patternA, filepath.Base(path))
		if errTI == nil {
			if matchedT {
				strListT = append(strListT, filepath.ToSlash(path))
			}
		}

		return nil
	})

	if errT != nil {
		if verboseA {
			fmt.Printf("Search directory failed: %v\n", errT.Error())
		}

		return nil
	}

	return strListT
}

func (p *MemMapFs) GenerateFileListRecursivelyWithExclusive(dirA string, patternA string, exclusivePatternA string, verboseA bool) []string {
	strListT := make([]string, 0, 100)

	errT := Walk(p, dirA, func(path string, f os.FileInfo, err error) error {
		if verboseA {
			fmt.Println(path)
		}

		if f == nil {
			return err
		}

		if f.IsDir() {
			return nil
		}

		matchedT, errTI := filepath.Match(patternA, filepath.Base(path))
		if errTI == nil {
			if matchedT {
				if exclusivePatternA != "" {
					matched2T, err2TI := filepath.Match(exclusivePatternA, filepath.Base(path))
					if err2TI == nil {
						if matched2T {
							return nil
						}
					}
				}

				strListT = append(strListT, filepath.ToSlash(path))
			}
		} else {
			fmt.Printf("matching failed: %v\n", errTI.Error())
		}

		return nil
	})

	if errT != nil {
		fmt.Printf("Search directory failed: %v\n", errT.Error())
		return nil
	}

	return strListT
}

func (p *MemMapFs) Ls(dirA string) []string {
	return p.GenerateFileListInDir(dirA, "*", false)
}

func (p *MemMapFs) Lsr(dirA string) []string {
	return p.GenerateFileListRecursivelyWithExclusive(dirA, "*", "", false)
}

var TimeFormatCompact2 = "2006/01/02 15:04:05"

func (p *MemMapFs) Log(fileNameA string, formatA string, argsA ...interface{}) {
	if strings.HasSuffix(formatA, "\n") {
		p.AppendStringToFile(fmt.Sprintf(fmt.Sprintf("[%v] ", time.Now().Format(TimeFormatCompact2))+formatA, argsA...), fileNameA)
	} else {
		p.AppendStringToFile(fmt.Sprintf(fmt.Sprintf("[%v] ", time.Now().Format(TimeFormatCompact2))+formatA+"\n", argsA...), fileNameA)
	}
}

func (p *MemMapFs) TarwalkFrom(source, target string, tw *tar.Writer) error {
	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	return filepath.Walk(source,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if baseDir != "" {
				header.Name = filepath.ToSlash(filepath.Join(baseDir, strings.TrimPrefix(path, source)))
			}

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tw, file)
			return err
		})
}

func (p *MemMapFs) Tarwalk(source, target string, tw *tar.Writer) error {
	info, err := p.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.ToSlash(filepath.Base(source))
	}

	return Walk(p, source,
		func(path string, info os.FileInfo, err error) error {
			if path == target {
				return nil
			}

			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			if baseDir != "" {
				header.Name = p.Join(baseDir, strings.TrimPrefix(path, source))
			}

			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			file, err := p.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tw, file)
			return err
		})
}

func (p *MemMapFs) TarFrom(pathsA []string, tarPathA string) error {
	file, err := p.Create(tarPathA)
	if err != nil {
		return err
	}

	defer file.Close()

	var fileReader io.WriteCloser = file

	if strings.HasSuffix(tarPathA, ".gz") {
		fileReader = gzip.NewWriter(file)

		defer fileReader.Close()
	}

	tw := tar.NewWriter(fileReader)
	defer tw.Close()

	for _, i := range pathsA {
		if err := p.TarwalkFrom(i, "", tw); err != nil {
			return err
		}
	}

	return nil

}

func (p *MemMapFs) Tar(pathsA []string, tarPathA string) error {
	file, err := p.Create(tarPathA)
	if err != nil {
		return err
	}

	defer file.Close()

	var fileReader io.WriteCloser = file

	if strings.HasSuffix(tarPathA, ".gz") {
		fileReader = gzip.NewWriter(file)

		defer fileReader.Close()
	}

	tw := tar.NewWriter(fileReader)
	defer tw.Close()

	for _, i := range pathsA {
		if err := p.Tarwalk(i, tarPathA, tw); err != nil {
			return err
		}
	}

	return nil

}

func (p *MemMapFs) UntarFrom(sourcefile, extractPath string) error {
	file, err := os.Open(sourcefile)

	if err != nil {
		return err
	}

	defer file.Close()

	var fileReader io.ReadCloser = file

	if strings.HasSuffix(sourcefile, ".gz") {
		if fileReader, err = gzip.NewReader(file); err != nil {
			return err
		}
		defer fileReader.Close()
	}

	tarBallReader := tar.NewReader(fileReader)

	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		filename := p.Join(extractPath, filepath.FromSlash(header.Name))

		switch header.Typeflag {
		case tar.TypeDir:
			err = p.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755 if you prefer

			if err != nil {
				return err
			}

		case tar.TypeReg:
			writer, err := p.Create(filename)

			if err != nil {
				return err
			}

			io.Copy(writer, tarBallReader)

			err = p.Chmod(filename, os.FileMode(header.Mode))

			if err != nil {
				return err
			}

			writer.Close()
		default:
			// log.Printf("Unable to untar type: %c in file %s", header.Typeflag, filename)
		}
	}
	return nil
}

func (p *MemMapFs) Untar(sourcefile, extractPath string) error {
	file, err := p.Open(sourcefile)

	if err != nil {
		return err
	}

	defer file.Close()

	var fileReader io.ReadCloser = file

	if strings.HasSuffix(sourcefile, ".gz") {
		if fileReader, err = gzip.NewReader(file); err != nil {
			return err
		}
		defer fileReader.Close()
	}

	tarBallReader := tar.NewReader(fileReader)

	for {
		header, err := tarBallReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		filename := p.Join(extractPath, filepath.FromSlash(header.Name))

		switch header.Typeflag {
		case tar.TypeDir:
			err = p.MkdirAll(filename, os.FileMode(header.Mode)) // or use 0755 if you prefer

			if err != nil {
				return err
			}

		case tar.TypeReg:
			writer, err := p.Create(filename)

			if err != nil {
				return err
			}

			io.Copy(writer, tarBallReader)

			err = p.Chmod(filename, os.FileMode(header.Mode))

			if err != nil {
				return err
			}

			writer.Close()
		default:
			// log.Printf("Unable to untar type: %c in file %s", header.Typeflag, filename)
		}
	}
	return nil
}
