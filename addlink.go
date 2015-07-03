package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var fileList []string

var chart = map[string]string{
	"instances":       "instance",
	"disks":           "disk",
	"snapshots":       "snapshot",
	"images":          "image",
	"networks":        "network",
	"security_groups": "securitygroup",
	"vpcs":            "vpc",
	"vrouters":        "vrouter",
	"vswitches":       "vswitch",
	"route_tables":    "routertable",
	"regions":         "region",
	"monitoring":      "monitor",
	"instance_types":  "other",
}

const URL_PREFIX = "http://docs.aliyun.com/#/pub/"

func WalkRule(path string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() {
		return nil
	}

	haveFile := strings.HasSuffix(path, ".go")
	if haveFile {
		testFile := strings.HasSuffix(path, "[*]test.go")
		if testFile == false {
			fileList = append(fileList, path)
		} else {
			fmt.Println(path)
		}
	}

	return nil
}

func GetFilelist(path string) {
	err := filepath.Walk(path, WalkRule)

	if err != nil {
		fmt.Errorf("filepath.Walk() returned %v\n", err)
	}
}

func GetPackageAndFileName(path string) (string, string, string) {
	fmt.Println("Getting filename from: ", path)

	directory := strings.Split(path, "/")

	pkgPath := GetPackagePath(path, directory[len(directory)-1])
	pkgName := directory[len(directory)-2]
	fileName := directory[len(directory)-1]

	return pkgPath, pkgName, fileName
}

func GetPackagePath(path string, filename string) string {
	fmt.Println("Getting package path from: ", path)

	result := strings.Split(path, filename)

	return result[0]
}

func DealFile(path string) {
	fmt.Println("Dealing file: ", path)

	pkgPath, pkgName, fileName := GetPackageAndFileName(path)

	docExist, ok := chart[strings.Split(fileName, ".")[0]]
	if ok == false {
		return
	}

	url := "// You can read doc at:" + URL_PREFIX + pkgName + "/" + "open-api/" + docExist + "&"

	fread, err := os.Open(path)
	defer fread.Close()

	if err != nil {
		fmt.Println(path, err)
		return
	} else {
		inbuff := bufio.NewReader(fread)

		newPath := NewFilePath(pkgPath, fileName)

		fwrite, err := os.Create(newPath)
		defer fwrite.Close()
		if err != nil {
			fmt.Println(newPath, err)
			return
		} else {
			fmt.Println(newPath)
		}

		for {
			line, err := inbuff.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}

			//^func
			isMatch, err := regexp.MatchString("^func", line)
			if isMatch {
				urlDescribe := url + GetFuncName(line)
				fwrite.WriteString("//\n")
				fwrite.WriteString(urlDescribe)
				fwrite.WriteString("\n")
				fwrite.WriteString(line)

				fmt.Println(urlDescribe)
			} else {
				fwrite.WriteString(line)
			}
		}

		WriteBackAndRemove(newPath, path)

		fmt.Println("finish reading", pkgName, url)
	}
}

func NewFilePath(pkgPath string, fileName string) string {
	name := strings.Split(fileName, ".")
	newName := name[0] + "_temp.go"
	newFilePath := pkgPath + newName

	return newFilePath
}

func GetFuncName(line string) string {
	fmt.Println(line)

	buff := []byte(line)
	var nameBytes []byte
	//跳过“func ”

	pos := 4
	for buff[pos] == ' ' {
		pos++
	}

	if buff[pos] != '(' {
		for i := pos; buff[i] != '('; i++ {
			nameBytes = append(nameBytes, buff[i])
		}
	} else {
		i := pos
		for buff[i] != ')' {
			i++
		}

		i = i + 2 // ") " 跳过空格

		for buff[i] != '(' {
			nameBytes = append(nameBytes, buff[i])
			i++
		}
	}

	return strings.ToLower(string(nameBytes))
}

func WriteBackAndRemove(temp string, src string) {
	fread, err := os.Open(temp)
	defer fread.Close()

	if err != nil {
		fmt.Println(temp, err)
		return
	} else {
		inbuff := bufio.NewReader(fread)

		fwrite, err := os.Create(src)
		defer fwrite.Close()
		if err != nil {
			fmt.Println(src, err)
			return
		}

		for {
			line, err := inbuff.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			} else {
				fwrite.WriteString(line)
			}
		}

		os.Remove(temp)
	}
}

func main() {
	GetFilelist("/Users/zhangyf/Documents/GitHub/aliyungo/ecs")

	for i := 0; i < len(fileList); i++ {
		DealFile(fileList[i])
	}
}
