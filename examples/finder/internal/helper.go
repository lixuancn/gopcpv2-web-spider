package internal

import (
	"fmt"
	"path/filepath"
	"os"
)

func checkDirPath(dirPath string)(absDirPath string, err error){
	if dirPath == ""{
		err = fmt.Errorf("无效的路径", dirPath)
	}
	if filepath.IsAbs(dirPath){
		absDirPath = dirPath
	}else{
		absDirPath, err = filepath.Abs(dirPath)
		if err != nil {
			return
		}
	}
	dir, err := os.Open(absDirPath)
	if err != nil && !os.IsNotExist(err){
		return
	}
	if dir == nil{
		err = os.MkdirAll(absDirPath, 0700)
		if err != nil && !os.IsExist(err){
			return
		}
	}else {
		var fileInfo os.FileInfo
		fileInfo, err = dir.Stat()
		if err != nil{
			return
		}
		if !fileInfo.IsDir(){
			err = fmt.Errorf("必须是一个目录：%s", absDirPath)
			return
		}
	}
	return
}

func Record(level byte, content string){
	if content == ""{
		return
	}
	switch level {
	case 0:
		fmt.Println("[INFO]", content)
	case 1:
		fmt.Println("[WARN]", content)
	case 2:
		fmt.Println("[INFO]", content)
	}
}