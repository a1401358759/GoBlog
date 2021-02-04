package cus

import (
	"crypto/sha1"
	"fmt"
	"goblog/core/global"
	"goblog/service/admin"
	"goblog/utils"
	"io"
	"os"
	"path"
	"strings"
)

func ImportMetaData(file string, logFilePath string) {
	var newUpdateCount, newRevisionCount, newLanguageCount, newFileCount, err = admin.UploadFile(file)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed! You have imported {%d} updates with {%d} revisions successfully.", newUpdateCount, newRevisionCount))
		fmt.Println(fmt.Sprintf("Failed! You have imported {%d} languages with {%d} file successfully. ", newLanguageCount, newFileCount))
	}
	fmt.Println(fmt.Sprintf("Congratulation! You have imported {%d} updates with {%d} revisions successfully.", newUpdateCount, newRevisionCount))
	fmt.Println(fmt.Sprintf("Congratulation! You have imported {%d} languages with {%d} file successfully. ", newLanguageCount, newFileCount))
}

func ExportMetaData(file string, logFilePath string) {
	admin.ExportFile(file, logFilePath)
	fmt.Println("export metadata" + " success!")
	fmt.Println(file)
	fmt.Println(logFilePath)
}

func MoveContent(src string) {
	m := global.GConfig.CUS

	f, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	h := sha1.New()
	buf := make([]byte, 65536)
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if 0 == n {
			break
		}
		h.Write(buf)
	}
	fileTypeSlice := strings.Split(src, ".")
	fileType := fileTypeSlice[len(fileTypeSlice)-1]
	fileHashName := h.Sum(nil)
	fileName := fmt.Sprintf("%x", fileHashName)
	fileName = fileName + "." + fileType

	// 生成文件夹目录
	folderName := fileName[len(fileName)-2:]
	dir := path.Join(m.DirPath, folderName)
	if !utils.IsDir(dir) {
		os.MkdirAll(dir, 0777)
	}

	dst := path.Join(m.DirPath, folderName, fileName)

	nBytes, err := utils.Copy(src, dst)
	if err != nil {
		fmt.Printf("The copy operation failed %q\n", err)
	} else {
		fmt.Printf("Copied %d bytes!\n", nBytes)
	}
	return
}
