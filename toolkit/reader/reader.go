package reader

import (
	"io"
	"io/ioutil"
	"fmt"
	"bytes"
)

//代表多重读取器的接口。
type MultipleReader interface {
	Reader() io.ReadCloser
}

type myMultipleReader struct{
	data []byte
}

func NewMultipleReader(reader io.Reader)(MultipleReader, error){
	var data []byte
	var err error
	if reader != nil{
		data, err = ioutil.ReadAll(reader)
		if err != nil{
			return nil, fmt.Errorf("多重读取器： 创建实例失败")
		}

	}else{
		data = []byte{}
	}
	return &myMultipleReader{data: data}, nil
}


func (rr *myMultipleReader)Reader()io.ReadCloser{
	return ioutil.NopCloser(bytes.NewReader(rr.data))
}