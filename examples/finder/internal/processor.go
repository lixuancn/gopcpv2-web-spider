package internal

import (
	"io"
	"fmt"
	"path/filepath"
	"os"
	"errors"
	"gopcpv2-web-spider/module"
)

func genItemProcessors(dirPath string)[]module.ProcessItem{
	savePicture := func(item module.Item)(result module.Item, err error){
		if item == nil {
			return nil, errors.New("无效的条目!")
		}
		// 检查和准备数据。
		var absDirPath string
		if absDirPath, err = checkDirPath(dirPath); err != nil {
			return
		}
		v := item["reader"]
		reader, ok := v.(io.Reader)
		if !ok {
			return nil, fmt.Errorf("无效的读取器类型: %T", v)
		}
		readCloser, ok := reader.(io.ReadCloser)
		if ok {
			defer readCloser.Close()
		}
		v = item["name"]
		name, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("无效的名称类型: %T", v)
		}
		// 创建图片文件。
		fileName := name
		filePath := filepath.Join(absDirPath, fileName)
		f, err := os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("创建文件失败: %s (path: %s)", err, filePath)
		}
		defer f.Close()
		// 写图片文件。
		_, err = io.Copy(f, reader)
		if err != nil {
			return nil, err
		}
		result = make(map[string]interface{})
		for k, v := range item{
			result[k] = v
		}
		result["file_path"] = dirPath
		fileInfo, err := f.Stat()
		if err != nil{
			return nil, err
		}
		result["file_size"] = fileInfo.Size()
		return result, nil
	}

	recordPicture := func(item module.Item)(result module.Item, err error){
		v := item["file_path"]
		path, ok := v.(string)
		if !ok{
			return nil, fmt.Errorf("无效的文件路径的类型：%T", v)
		}
		v = item["file_size"]
		size, ok := v.(int64)
		if !ok{
			return nil, fmt.Errorf("无效的文件名类型：%T", v)
		}
		fmt.Println(fmt.Sprintf("保存了文件, 路径%s，大小%d", path, size))
		return nil, nil
	}
	return []module.ProcessItem{savePicture, recordPicture}
}