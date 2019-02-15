package net_utils

import (
	"io"
	"image"
	"image/jpeg"
	"image/png"
	"image/gif"
	"errors"
	"os"
	"strings"
	"fmt"
	"encoding/base64"
)

func CopyToMany(src io.Reader, args ... io.Writer) (err error) {
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

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/*
* 图片裁剪
* 入参:
* 规则:如果精度为0则精度保持不变
*
* 返回:error
 */
func Clip(in io.Reader, out io.Writer, x0, y0, x1, y1, quality int) (ext string, err error) {
	origin, fm, err := image.Decode(in)
	if err != nil {
		return "", err
	}

	if x1 == -1 {
		x1 = origin.Bounds().Size().X
	}
	if y1 == -1 {
		y1 = origin.Bounds().Size().Y
	}

	switch fm {
	case "jpeg":
		img := origin.(*image.YCbCr)

		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.YCbCr)
		return "jpg", jpeg.Encode(out, subImg, &jpeg.Options{quality})
	case "png":
		switch origin.(type) {
		case *image.NRGBA:
			img := origin.(*image.NRGBA)
			subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.NRGBA)
			return "png", png.Encode(out, subImg)
		case *image.RGBA:
			img := origin.(*image.RGBA)
			subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
			return "png", png.Encode(out, subImg)
		}
	case "gif":
		img := origin.(*image.Paletted)
		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.Paletted)
		return "gif", gif.Encode(out, subImg, &gif.Options{})
	default:
		return "", errors.New("ERROR FORMAT")
	}
	return "", nil
}

func IsBase64Image(url string) bool {
	if strings.HasPrefix(url, "data:") {
		return true
	}
	return false
}

func ConvertBase64Image(url string) (imageData []byte, ext string, err error) {
	temp := strings.Split(url, ";base64,")
	if len(temp) < 2 {
		fmt.Println("len is must more than 2")
		return
	}
	ext = strings.Replace(temp[0], "data:image/", "", -1)
	base64data := strings.Replace(temp[1], "\n", "", -1)
	imageData, err = base64.StdEncoding.DecodeString(base64data)
	if err != nil {
		return nil, "", err
	}

	return
}
