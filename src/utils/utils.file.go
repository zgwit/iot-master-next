package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"sort"
)

func ReadFile(filePath string) (content string, err error) {
	var (
		file         *os.File
		file_info    fs.FileInfo
		content_byte []byte
	)

	if file, err = os.Open(filePath); err != nil {
		return
	}
	defer file.Close()

	if file_info, err = file.Stat(); err != nil {
		return
	}

	content_byte = make([]byte, file_info.Size())

	if _, err = file.Read(content_byte); err != nil {
		return
	}

	content = string(content_byte)

	return
}

func CopyFile(originalPath string, targetPath string) error {

	src, err := os.Open(originalPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dest, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	return nil
}

func ReadFileToObject(filePath string, object any) (err error) {

	var (
		content string
	)

	if content, err = ReadFile(filePath); err != nil {
		return
	}

	if err = json.Unmarshal([]byte(content), object); err != nil {
		return
	}

	return
}

func ReadFilekeyToObject(filePath, key string, object any) (err error) {

	var (
		content string
		data    map[string]interface{}
	)

	if content, err = ReadFile(filePath); err != nil {
		return
	}

	if content == "" {
		return
	}

	if err = json.Unmarshal([]byte(content), &data); err != nil {
		return
	}

	if _, exist := data[key]; !exist {
		return errors.New("key " + key + " is not exist")
	}

	data_byte := []byte{}

	if data_byte, err = json.Marshal(data[key]); err != nil {
		return
	}

	if err = json.Unmarshal(data_byte, object); err != nil {
		return
	}

	return
}

func GetDirsNames(dirName string) (names []string, err error) {
	var (
		dirEntries = []fs.DirEntry{}
	)
	names = []string{}

	if dirEntries, err = os.ReadDir(dirName); err != nil {
		return
	}

	for _, entry := range dirEntries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	return
}

func GetDirsNames2(dirName string) (names []string, err error) {

	var (
		folder    *os.File
		fileInfos []fs.FileInfo
	)

	names = []string{}

	if folder, err = os.Open(dirName); err != nil {
		return
	}
	defer folder.Close()

	if fileInfos, err = folder.Readdir(-1); err != nil {
		return
	}

	var files []os.FileInfo

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			files = append(files, fileInfo)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	for _, fileInfo := range files {
		names = append([]string{fileInfo.Name()}, names...)
	}

	return
}

func GetDirFileNames(dirName string) (names []string, err error) {
	var (
		dirEntries = []fs.DirEntry{}
	)
	names = []string{}

	if dirEntries, err = os.ReadDir(dirName); err != nil {
		return
	}

	for _, entry := range dirEntries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}

	return
}

func GetDirFileNames2(dirName string) (names []string, err error) {

	var (
		folder    *os.File
		fileInfos []fs.FileInfo
	)

	names = []string{}

	// 打开文件夹
	if folder, err = os.Open(dirName); err != nil {
		return
	}
	defer folder.Close()

	if fileInfos, err = folder.Readdir(-1); err != nil {
		return
	}

	var files []os.FileInfo

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			files = append(files, fileInfo)
		}
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	for _, fileInfo := range files {
		names = append([]string{fileInfo.Name()}, names...)
	}

	return
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func DirExistOrCreate(dirName string) (err error) {

	_, err = os.Stat(dirName)

	if !os.IsNotExist(err) {
		return
	}

	if err = os.MkdirAll(dirName, 0755); err != nil {
		return
	}

	return
}

func FileExistOrCreate(fileName string) (err error) {

	var (
		file *os.File
	)

	_, err = os.Stat(fileName)

	if !os.IsNotExist(err) {
		return
	}

	if file, err = os.Create(fileName); err != nil {
		return
	}
	defer file.Close()

	return
}

func WriteObjectToFile(filename string, object any) (err error) {

	var (
		file      *os.File
		json_byte []byte
	)

	if json_byte, err = json.Marshal(object); err != nil {
		return
	}

	file, err = os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	if _, err = file.WriteString(string(json_byte)); err != nil {
		return
	}

	return
}

func WriteObjectToFile2(filename string, object any) (err error) {

	var (
		file       *os.File
		json_byte  []byte
		strOutByte bytes.Buffer
	)

	if json_byte, err = json.Marshal(object); err != nil {
		return
	}

	if err = json.Indent(&strOutByte, json_byte, "", "    "); err == nil {
		json_byte = strOutByte.Bytes()
	}

	file, err = os.OpenFile(filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	if _, err = file.WriteString(string(json_byte)); err != nil {
		return
	}

	return
}

func RemoveAll(dirName string) (err error) {

	if err = os.RemoveAll(dirName); err != nil {
		return
	}
	return
}

func ReadFileTail(filename string, n int) ([]string, error) {

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, n)
	var i int
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		i++
		if i > n {
			lines = lines[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
