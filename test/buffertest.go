package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

func copyToMany(src io.Reader, args ... io.Writer) (err error) {
	//data, err := ioutil.ReadAll(src)
	//if err != nil {
	//	return
	//}
	const ReadLen = 4096
	buf := make([]byte, ReadLen)
	var readLen = 0
	for {
		readLen, err = src.Read(buf)
		if err != nil {
			return
		}
		for _, writer := range args {
			_, err := writer.Write(buf[0:readLen])
			if err != nil {
				return err
			}
		}
		if readLen != ReadLen {
			break
		}
	}
	return
}

func test() {
	mydata := make([]byte, 0)
	buffer := bytes.NewBuffer(mydata)
	_, err := buffer.Write([]byte("123456789"))
	if err != nil {
		fmt.Println(err)
		return
	}

	buffer1 := &bytes.Buffer{}
	io.Copy(buffer1, buffer)
	data, err := ioutil.ReadAll(buffer)
	fmt.Println("buffer : ", data)

	data, err = ioutil.ReadAll(buffer1)
	fmt.Println("buffer1 : ", data)
}

func test2()  {
	buffer := &bytes.Buffer{}
	buffer.Write([]byte("hello world"))
	buffer.Write([]byte("hello world"))
	buffer.Write([]byte("hello world"))
	buffer.Write([]byte("hello world"))

	outBuffer1 := &bytes.Buffer{}
	outBuffer2 := &bytes.Buffer{}
	copyToMany(buffer, outBuffer1, outBuffer2)

	data, err := ioutil.ReadAll(outBuffer1)
	if err != nil {
		fmt.Println(data)
		return
	}
	fmt.Println("outBuffer1", data)

	data2, err := ioutil.ReadAll(outBuffer2)
	if err != nil {
		fmt.Println(data2)
		return
	}
	fmt.Println("outBuffer2", data2)
}
func main() {
	bytesBuffer := []byte{'h','e','l','l','o','w','r','o','l','d'}

	buffer := bytes.NewBuffer(bytesBuffer)
	readBuf:= make([]byte,4)
	buffer.Read(readBuf)
	buffer.Write([]byte("123"))
	buffer.Read(readBuf)
	fmt.Println(readBuf)
	fmt.Println(bytesBuffer)
	fmt.Println(buffer)
}
