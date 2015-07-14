package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/bitly/go-simplejson"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//docs link
const DOCS_URL = "http://docs.aliyun.com"

type Funclist map[string]bool

var allDocs = make(map[string]Funclist)

var ECSAPI = map[string]bool{
	"instance":      true,
	"disk":          true,
	"snapshot":      true,
	"image":         true,
	"network":       true,
	"securitygroup": true,
	"vpc":           true,
	"vrouter":       true,
	"vswitch":       true,
	"routertable":   true,
	"region":        true,
	"monitor":       true,
	"other":         true,
	"datatype":      true,
	"appendix":      true,
}

var OSSAPI = map[string]bool{
	"service":          true,
	"bucket":           true,
	"object":           true,
	"multipart-upload": true,
	"cors":             true,
}

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

var ossChart = map[string]string{
	"PutBucket":        "PutBucket",
	"ACL":              "GetBucketAcl",
	"PutBucketWebsite": "PutBucketWebsite",
	"Location":         "GetBucketLoaction",
	"DelBucket":        "DeleteBucket",
	"List":             "GetBucket",
	"Put":              "PutObject",
	"PutCopy":          "CopyObject",
	"Get":              "GetObject",
	"Del":              "DeleteObject",
	"DelMulti":         "DeleteMultipleObjects",
	"Head":             "HeadObject",
	"InitMulti":        "InitiateMultipartUpload",
	"PutPart":          "UploadPart",
	"PutPartCopy":      "UploadPartCopy",
	"Complete":         "Complete MultipartUpload",
	"Abort":            "AbortMultipartUpload",
	"ListParts":        "ListPartsFull",
}

const URL_PREFIX = "http://docs.aliyun.com/#/pub/"
const URL_OSS_PREFIX = "http://docs.aliyun.com/#/pub/oss/api-reference/"

var notexist = make([]string, 5)

func main() {
	var module string

	doc, err := goquery.NewDocument("http://docs.aliyun.com")
	if err != nil {
		fmt.Print(err)
	}

	doc.Find("script").Each(func(i int, contentSelection *goquery.Selection) {
		content := contentSelection.Text()

		cons := strings.Split(content, ";")

		for i := 0; i < len(cons); i++ {
			if strings.Contains(cons[i], "window.docModule") {
				module = cons[i]
			}
		}
	})

	module = strings.TrimPrefix(module, "\nwindow.docModule=JSON.parse('")
	module = strings.TrimSuffix(module, "')")

	//GetEcsDocsApi(module)
	GetOssDocsApi(module)
	//GetFilelist("/Users/zhangyf/Documents/GitHub/aliyungo/ecs")
	//GetFilelist("/Users/zhangyf/Documents/GoWork/src/github.com/denverdino/aliyungo/ecs")
	//GetFilelist("/home/ubuntu/Documents/GoWork/src/github.com/denverdino/aliyungo/oss")
	GetFilelist("/home/ubuntu/Documents/GitHub/aliyungo/oss")
	// GetFilelist("/home/ubuntu/Documents/GitHub/aliyungo/ecs")

	for i := 0; i < len(fileList); i++ {
		DealOssFile(fileList[i])
	}

	fmt.Println(notexist)
}

func GetEcsDocsApi(module string) {
	jsonstring, _ := simplejson.NewJson([]byte(module))
	version, _ := jsonstring.Get("ecs").Get("version").String()
	ecsDoc, _ := jsonstring.Get("ecs").Get("list").Array()
	fmt.Println("ecs version: " + version)

	//Get open-api
	var openAPI map[string]interface{}

	for _, d := range ecsDoc {
		element := d.(map[string]interface{})
		if element["name_en"] == "open-api" {
			openAPI = element
		}
	}
	oaFolder := openAPI["isFolder"].([]interface{})

	var docs map[string]interface{}
	for _, d := range oaFolder {
		docs = d.(map[string]interface{}) //取出指定类型api信息，如instance，disk。。。
		name := docs["name_en"].(string)  //取出api名字

		if ECSAPI[name] == true {
			docsFolder := docs["isFolder"].([]interface{})
			docsList := Funclist{}

			for _, d := range docsFolder {
				element := d.(map[string]interface{})
				funcname := element["name_en"].(string)
				docsList[funcname] = true
			}
			allDocs[name] = docsList
		}
	}

	fmt.Println(allDocs)
}

func GetOssDocsApi(module string) {
	jsonstring, _ := simplejson.NewJson([]byte(module))
	version, _ := jsonstring.Get("oss").Get("version").String()
	ecsDoc, _ := jsonstring.Get("oss").Get("list").Array()
	fmt.Println("oss version: " + version)

	//Get open-api
	var openAPI map[string]interface{}

	for _, d := range ecsDoc {
		element := d.(map[string]interface{})
		if element["name_en"] == "api-reference" {
			openAPI = element
		}
	}
	oaFolder := openAPI["isFolder"].([]interface{})

	var docs map[string]interface{}
	for _, d := range oaFolder {
		docs = d.(map[string]interface{}) //取出指定类型api信息，如instance，disk。。。
		name := docs["name_en"].(string)  //取出api名字

		if OSSAPI[name] == true {
			docsFolder := docs["isFolder"].([]interface{})
			docsList := Funclist{}

			for _, d := range docsFolder {
				element := d.(map[string]interface{})
				funcname := element["name_en"].(string)
				docsList[funcname] = true
			}
			allDocs[name] = docsList
		}
	}

	fmt.Println(allDocs)
}

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

func DealOssFile(path string) {
	fmt.Println("Dealing Oss file: ", path)

	pkgPath, pkgName, fileName := GetPackageAndFileName(path)

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
				fname := GetFuncName(line)
				apiName, isExist := isExistApiDocs(fname)
				if isExist {
					urlDescribe := URL_OSS_PREFIX + apiName + "&" + ossChart[fname]
					fwrite.WriteString("//\n")
					fwrite.WriteString("// You can read doc at ")
					fwrite.WriteString(urlDescribe)
					fwrite.WriteString("\n")
					fwrite.WriteString(line)

					fmt.Println(urlDescribe)
				} else {
					fname = fname
					notexist = append(notexist, fname)
					fwrite.WriteString(line)
				}

			} else {
				fwrite.WriteString(line)
			}
		}

		WriteBackAndRemove(newPath, path)

		fmt.Println("finish reading", pkgName)
	}
}

func DealFile(path string) {
	fmt.Println("Dealing file: ", path)

	pkgPath, pkgName, fileName := GetPackageAndFileName(path)

	docExist, ok := chart[strings.Split(fileName, ".")[0]]
	if ok == false {
		return
	}

	url := URL_PREFIX + pkgName + "/" + "open-api/" + docExist + "&"

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
				fname := strings.ToLower(GetFuncName(line))
				if allDocs[docExist][fname] {
					urlDescribe := url + fname
					fwrite.WriteString("//\n")
					fwrite.WriteString("// You can read doc at ")
					fwrite.WriteString(urlDescribe)
					fwrite.WriteString("\n")
					fwrite.WriteString(line)

					fmt.Println(urlDescribe)
				} else {
					fname = docExist + "&" + fname
					notexist = append(notexist, fname)
					fwrite.WriteString(line)
				}

			} else {
				fwrite.WriteString(line)
			}
		}

		WriteBackAndRemove(newPath, path)

		fmt.Println("finish reading", pkgName, url)
	}
}

func isExistApiDocs(fname string) (string, bool) {
	var apiName []string
	for key := range allDocs {
		apiName = append(apiName, key)
	}

	for k := range apiName {
		api := apiName[k]
		if allDocs[api][ossChart[fname]] == true {
			return api, true
		}
	}
	return "", false
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

	return string(nameBytes)
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

