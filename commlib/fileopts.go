package commlib

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
)

func CreateFolder(abs_path string) bool {
	if _, err := os.Stat(abs_path); os.IsNotExist(err) {
		if c_err := os.MkdirAll(abs_path, 0777); c_err != nil {
			Mtrloggger.Println("Failed to create folder path for ", ":", c_err)
			return false
		}
	}
	return true
}

func CreateFile(abs_file string) (*os.File, error) {
	dir_name, _ := filepath.Split(abs_file)
	for i := 0; i < 3; i++ {
		if _, err := os.Stat(dir_name); os.IsNotExist(err) {
			if CreateFolder(dir_name) {
				goto Lable
			} else {
				continue
			}

		} else {
			goto Lable
		}
	}
	return nil, errors.New("Failed to create folder")
Lable:
	if filep, err := os.Create(abs_file); err == nil {
		return filep, err
	} else {
		return nil, err
	}
}

func ReadFile(abs_file string) ([]string, error) {
	var file_content []string
	if filep, err := os.Open(abs_file); err != nil {
		defer filep.Close()
		Mtrloggger.Printf("Not found File [%s]\n", abs_file)
		return nil, err
	} else {
		scanner := bufio.NewScanner(filep)
		for scanner.Scan() {
			line := scanner.Text()
			Mtrloggger.Println(line)
			file_content = append(file_content, line)
		}
		if err := scanner.Err(); err != nil {
			Mtrloggger.Printf("Read File Error with %v\n", err)
			return file_content, err
		}
	}
	return file_content, nil
}
