package utils

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"goblog/core/global"
	"golang.org/x/text/encoding/unicode"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

// Xml字符串压缩为zip包
func XmlCompress(xmlStr string) (string, error) {
	var buffer bytes.Buffer
	enc := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
	u16xmlStr, _ := enc.Bytes([]byte(xmlStr))
	writer := zip.NewWriter(&buffer)
	if zipF, err := writer.Create("update.xml"); err == nil {
		if _, err := zipF.Write([]byte(u16xmlStr)); err == nil {
			writer.Close()
			b64Compress := base64.StdEncoding.EncodeToString(buffer.Bytes())
			return b64Compress, nil
		}
	}
	writer.Close()
	return "", errors.New("压缩失败。")
}

// 解压交互中得到的压缩过的二进制文件为标准xml
func XmlUnCompress(b64Str string) (string, error) {
	if zipF, err := base64.StdEncoding.DecodeString(b64Str); err == nil {
		if reader, err := zip.NewReader(bytes.NewReader(zipF), int64(len(zipF))); err == nil {
			file := reader.File[0]
			xmlStream, _ := file.Open()
			u16xml, _ := ioutil.ReadAll(xmlStream)
			decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
			xml, _ := decoder.Bytes(u16xml[:])
			return string(xml), nil
		} else {
			if err.Error() == "zip: not a valid zip file" {
				// 当该文件格式为非zip包格式，则说明该文件为Ms .cab包格式
				// 写cab文件
				var err error
				cabFile, err := os.Create("meta.cab")
				writer := bufio.NewWriter(cabFile)
				_, err = bytes.NewReader(zipF).WriteTo(writer)
				err = writer.Flush()
				err = cabFile.Close()
				// 调用命令行解压cab
				cmd := exec.Command("cabextract", "meta.cab", "-f")
				stdout, _ := cmd.StdoutPipe()
				if err := cmd.Start(); err != nil {
					fmt.Println(err)
				}
				logScan := bufio.NewScanner(stdout)
				// 原微软所有的包，xml文件均名为blo
				xmlFileName := "blob"
				// 部分我们旧版本在WinServer环境下利用windows make cab工具自己打包的cab文件，xml名称为update.xml_XXX(uuid4)
				// 此处读取命令行的返回文本，从中获取准确的解压后的文件名称以便于读取和删除
				for logScan.Scan() {
					if strings.HasPrefix(strings.Replace(logScan.Text(), " ", "", -1), "extracting") {
						xmlFileName = strings.Replace(strings.Split(logScan.Text(), "extracting")[1], " ", "", -1)
					}
				}
				//err = cmd.Run()
				b, err := ioutil.ReadFile(xmlFileName)
				// utf16解码
				decoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewDecoder()
				bs2, err := decoder.Bytes(b[:])
				if err != nil {
					return "", err
				}
				os.Remove("meta.cab")
				os.Remove(xmlFileName)
				return string(bs2), nil
			}
		}
	}
	return "", errors.New("解压失败。")
}

//解压tar.gz文件
type tarGz struct {
	Xml []byte
	Txt [][]byte
}

func DecompressFiles(filePath string) ([]byte, [][]byte) {
	fd, err := os.Open(filePath)
	defer fd.Close()
	if err != nil {
		global.GLog.Error("archive failed!", zap.Any("error", err))
	}
	gzip_, err := gzip.NewReader(fd)
	defer gzip_.Close()
	if err != nil {
		global.GLog.Error("archive failed!", zap.Any("error", err))
	}
	tar_ := tar.NewReader(gzip_)
	files := &tarGz{}
	for {
		h, err := tar_.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			global.GLog.Error("archive failed!", zap.Any("error", err))
			break
		}
		SuffixArray := strings.Split(h.Name, ".")
		Suffix := SuffixArray[len(SuffixArray)-1]
		reader := bufio.NewReader(tar_)
		for {
			str, err := reader.ReadBytes('\n')
			if Suffix == "txt" {
				files.Txt = append(files.Txt, str)
			}
			if Suffix == "xml" {
				files.Xml = append(files.Xml, str...)
			}

			if err == io.EOF {
				fmt.Print(err)
				break
			}
		}

	}
	return files.Xml, files.Txt
}

//压缩文件夹下所有文件
func CompressFiles(rootDir, fileName string) (err error) {
	// file write
	fw, err := os.Create(rootDir + fileName)
	if err != nil {
		panic(err)
	}
	defer fw.Close()
	// gzip write
	gw := gzip.NewWriter(fw)
	defer gw.Close()
	// tar write
	tw := tar.NewWriter(gw)
	defer tw.Close()
	// 打开文件夹
	dir, err := os.Open(rootDir + "/export")
	if err != nil {
		panic(nil)
	}
	defer dir.Close()
	// 读取文件列表
	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	// 遍历文件列表
	for _, fi := range fis {
		// 跳过文件夹
		if fi.IsDir() {
			continue
		}
		// 打印文件名称
		fmt.Println(fi.Name())
		// 打开文件
		fr, err := os.Open(dir.Name() + "/" + fi.Name())
		if err != nil {
			panic(err)
		}
		defer fr.Close()
		// 信息头
		h := new(tar.Header)
		h.Name = fi.Name()
		h.Size = fi.Size()
		h.Mode = int64(fi.Mode())
		h.ModTime = fi.ModTime()
		// 写信息头
		err = tw.WriteHeader(h)
		if err != nil {
			return err
		}
		// 写文件
		_, err = io.Copy(tw, fr)
		if err != nil {
			return err
		}
	}
	return nil
}
