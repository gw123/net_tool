package main

import (
	"github.com/gw123/net_tool/rouding/net_utils"
	"github.com/gw123/net_tool/rouding/db/models"
	"github.com/gw123/net_tool/rouding/qiniu"
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"net/http"
	"time"
	"strings"
	"bytes"
	"qiniupkg.com/x/errors.v7"

	"github.com/gw123/net_tool/rouding/db"
	"io"
)

func getFileName(ext, md5 string) (filename, filepath string, err error) {
	if ext == "" {
		ext = "jpg"
	}
	//now := time.Now().UnixNano()
	filename = fmt.Sprintf("%s.%s", md5, ext)
	date := time.Now().Format("2006-01-02")
	rootpath := "storage/image"
	relpath := rootpath + "/" + date
	//isOk, err := net_utils.PathExists(relpath)
	//if err != nil {
	//	return
	//}
	//if !isOk {
	//	err = os.MkdirAll(relpath, 0660)
	//	if err != nil {
	//		return
	//	}
	//}
	filepath = relpath
	return
}

func downloadImg(url string) (filename, filepath string, err error) {
	rawBuffer := new(bytes.Buffer)
	if net_utils.IsBase64Image(url) {
		data, _, err := net_utils.ConvertBase64Image(url)
		if err != nil {
			return "", "", err
		}
		rawBuffer.Write(data)
	} else {
		// Request the HTML page.
		res, err := http.Get(url)
		if err != nil {
			return "", "", err
		}

		defer res.Body.Close()
		if res.StatusCode != 200 {
			data := fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status)
			return "", "", errors.New(data)
		}
		io.Copy(rawBuffer, res.Body)
	}

	buffer := new(bytes.Buffer)
	ext, err := net_utils.Clip(rawBuffer, buffer, 0, 42, -1, -1, 100)
	if err != nil {
		fmt.Println("clip error", err.Error())
	}
	fileMd5Str := fmt.Sprintf("%x", time.Now().UnixNano()/1000)
	filename, filepath, err = getFileName(ext, fileMd5Str)
	//fmt.Println("filename:", filename, "filepath", filepath)
	if err != nil {
		return filename, filepath, err
	}
	_, err = qiniu.UploadBuffer(buffer, filepath+"/"+filename)
	if err != nil {
		return filename, filepath, err
	}
	return filename, "http://data.xytschool.com/" + filepath, nil
}

func getResource(url, cate string) (article *models.Caijie, err error) {
	article = new(models.Caijie)
	article.FromUrl = url
	article.Cate = cate
	article.FromSite = "rouding"
	res, err := http.Get(url)
	if err != nil {
		return article, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		errorStr := fmt.Sprintf("返回状态码错误%s", res.StatusCode)
		return article, errors.New(errorStr)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return article, errors.New("NewDocumentFromReader: " + err.Error())
	}
	article.CreatedAt = time.Now()
	article.Title = doc.Find("title").Text()
	article.Title = strings.Replace(article.Title, "╭★肉丁网", "", -1)

	//Find the review items
	doc.Find("#zoom").Each(func(i int, s *goquery.Selection) {
		s.Find("a").Each(func(i int, selection *goquery.Selection) {
			selection.SetAttr("href", "")
		})

		s.Find("img").Each(func(index int, selection *goquery.Selection) {
			src, isExist := selection.Attr("src")
			if !isExist {
				return
			}
			if strings.Index(src, "//") == 0 {
				src = "https:" + src
			}
			filename, path, err := downloadImg(src)
			if err != nil {
				fmt.Println(err)
				return
			}
			selection.SetAttr("src", path+"/"+filename)
		})
		article.Content, err = s.Html()
	})
	return
}

func getResourceList(url string) (list []string, err error) {
	response, err := http.Get(url)

	if err != nil {
		return
	}
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return list, errors.New("NewDocumentFromReader: " + err.Error())
	}
	var proto string
	if strings.HasPrefix(url, "https") {
		proto = "https"
	} else {
		proto = "http"
	}

	doc.Find(".x5").Each(func(i int, selection *goquery.Selection) {
		href, isExists := selection.Find("a").Attr("href")
		if !isExists {
			return
		}
		if strings.HasPrefix(href, "//") {
			href = proto + ":" + href
		}
		list = append(list, href)
		//fmt.Println(href)
	})
	return
}

func main() {
	db.Connect()
	for index := 47; index <= 60; index++ {
		fmt.Println("caiji the ", index, " page")
		url := fmt.Sprintf("http://www.rouding.com/chuantongshougong/%d.html", index)
		list, err := getResourceList(url)
		if err != nil {
			fmt.Println("getResourceList", err)
			return
		}

		for index, resourceUrl := range list {
			fmt.Println("caiji resource", index, resourceUrl)
			resource, err := getResource(resourceUrl, "chuantongshougong")
			if err != nil {
				fmt.Println("caiji failed !!", resourceUrl, err)
				continue
			}
			//fmt.Println(resource)
			db.DbInstance.Create(resource)
		}
		time.Sleep(time.Second * 2)
		fmt.Println("caiji the ", index, " page", "over!")
	}
	//resource, err := getResource("http://www.rouding.com/life-diy/2.html")
	//fmt.Println(resource, err)
	//db.DbInstance.Create(resource)
}
