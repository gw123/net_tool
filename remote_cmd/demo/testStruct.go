package main

import (
	"github.com/gw123/remote_cmd/types"
	"unsafe"
	"fmt"
	"reflect"
)

type ByteSlice struct {
	addr uintptr
	len  int
	cap  int
}

func LoginServer2ByteSlice(data types.LoginServer) []byte {
	size := unsafe.Sizeof(data)
	bytes := &ByteSlice{
		addr: uintptr(unsafe.Pointer(&data)),
		cap:  int(size),
		len:  int(size),
	}
	byteSliceStructPointer := unsafe.Pointer(bytes)
	byteSlicePointer := (*[]byte)(byteSliceStructPointer)
	return *byteSlicePointer
}

func ByteSlice2LoginServer(data []byte) *types.LoginServer {
	byteSlicePointer := unsafe.Pointer(&data)
	byteSliceStructPointer := *(**types.LoginServer)(byteSlicePointer)
	return byteSliceStructPointer
}

func GetType(data interface{}) {
	fmt.Println("type: ", reflect.TypeOf(data))
	fmt.Println("value: ", reflect.ValueOf(data))
}

func main() {
	loginServer := &types.LoginServer{
		ClientId:    10,
		Timestamp:   2,
		ConnectType: 1,
	}
	//size := unsafe.Sizeof(*loginServer)
	//fmt.Println(size)
	data := LoginServer2ByteSlice(*loginServer)
	fmt.Println(data)

	reLoginServer := ByteSlice2LoginServer(data)
	fmt.Println((*reLoginServer).ClientId)
	fmt.Println((*reLoginServer).Timestamp)

	GetType(*loginServer)
	//data := *(*[]byte)(unsafe.Pointer(testBytes))
}
