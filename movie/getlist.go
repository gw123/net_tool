package main

import (
	"github.com/gw123/net_tool/movie/net_utils"
	"github.com/gw123/net_tool/movie/db/models"
	"github.com/gw123/net_tool/movie/qiniu"
	"github.com/PuerkitoBio/goquery"
	"fmt"
	"net/http"
	"time"
	"strings"
	"bytes"
	"qiniupkg.com/x/errors.v7"

	"github.com/gw123/net_tool/movie/db"
	"io"
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"flag"
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

	//buffer := new(bytes.Buffer)
	//ext, err := net_utils.Clip(rawBuffer, buffer, 0, 0, -1, -1, 100)
	//if err != nil {
	//	fmt.Println("clip error", err.Error())
	//}

	fileMd5Str := fmt.Sprintf("%x", time.Now().UnixNano()/1000)
	filename, filepath, err = getFileName("jpg", fileMd5Str)
	//fmt.Println("filename:", filename, "filepath", filepath)
	if err != nil {
		return filename, filepath, err
	}
	_, err = qiniu.UploadBuffer(rawBuffer, filepath+"/"+filename)
	if err != nil {
		return filename, filepath, err
	}
	return filename, "http://data.xytschool.com/" + filepath, nil
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func getResource(url, cate string) (movie *models.Movie, err error) {
	movie = new(models.Movie)
	movie.FromUrl = url

	res, err := http.Get(url)
	if err != nil {
		return movie, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		errorStr := fmt.Sprintf("返回状态码错误%s", res.StatusCode)
		return movie, errors.New(errorStr)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return movie, errors.New("NewDocumentFromReader: " + err.Error())
	}
	//Find the review items
	doc.Find("body > font > table:nth-child(4) > tbody > tr:nth-child(1) > td:nth-child(2) > table").Each(func(i int, s *goquery.Selection) {
		s.Find("td:nth-child(1)").Each(func(i int, selection *goquery.Selection) {
			//selection.SetAttr("href", "")
			t, err := GbkToUtf8([]byte(selection.Text()))
			if err != nil {
				fmt.Println(err)
				return
			}
			tdstr := string(t)
			tdArr := strings.Split(tdstr, "：")
			if len(tdArr) >= 2 {
				switch tdArr[0] {
				case "影片名称":
					if len(tdArr) == 3 {
						movie.Title = tdArr[1] + "：" + tdArr[2]
					} else if len(tdArr) == 2 {
						movie.Title = tdArr[1]
					}
					fmt.Printf("[%s=>%s]\n", tdArr[0], movie.Title)
					break;
				case "影片备注":
					movie.Note = tdArr[1];
					break;
				case "影片演员":
					movie.Actor = tdArr[1];
					break;
				case "影片导演":
					movie.Direction = tdArr[1];
					break;
				case "影片类型":
					movie.Tppe = tdArr[1];
					break;
				case "影片地区":
					movie.Area = tdArr[1];
					break;
				case "更新时间":
					movie.LastUpdatedTime = tdArr[1];
					break;
				case "影片状态":
					movie.Status = tdArr[1];
					break;
				case "影片语言":
					movie.Language = tdArr[1];
					break;
				case "上映日期":
					movie.PublishedTime = tdArr[1];
					break;
				}
			} else if len(tdArr) == 1 {
				movie.Desc = tdArr[0]
			} else {
				fmt.Println("td 解析错误 长度不对等2")
			}
		})
		//article.Content, err = s.Html()
	})
	//proto = "http://"

	doc.Find(".img img").Each(func(index int, selection *goquery.Selection) {
		src, isExist := selection.Attr("src")
		if !isExist {
			return
		}
		if strings.Index(src, "//") == 0 {
			src = "https:" + src
		}
		src = "http://91zy.cc/" + src
		//fmt.Printf("[%s=>%s]\n", "图片地址", src)
		filename, path, err := downloadImg(src)
		if err != nil {
			fmt.Println(err)
			return
		}
		movie.Img = path + "/" + filename
	})

	doc.Find("body > font > table:nth-child(4) > tbody > tr > td > table > tbody > tr > td > a").
		Each(func(index int, selection *goquery.Selection) {
		t, err := GbkToUtf8([]byte(selection.Text()))
		if err != nil {
			fmt.Println(err)
			return
		}
		tdstr := string(t)
		urltmp := strings.Split(tdstr, "$")
		url = urltmp[1]
		//fmt.Println("########",index)
		if index == 0 {
			movie.Url1 = url
		} else if index == 1 {
			movie.Url2 = url
		}
	})
	//fmt.Println(*movie)
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
	domain := "91zy.cc"
	doc.Find("body > font > table:nth-child(3) > tbody > tr > td:nth-child(1)").Each(func(i int, selection *goquery.Selection) {
		href, isExists := selection.Find("a").Attr("href")
		if !isExists {
			return
		}
		if strings.HasPrefix(href, "//") {
			href = proto + ":" + href
		} else if strings.HasPrefix(href, "/") {
			href = proto + "://" + domain + href
		}
		list = append(list, href)
		fmt.Println(href)
	})
	return
}

func caijiCate(cate int) {
	for index := 1; index <= 60; index++ {
		fmt.Println("caiji the ", index, " page")
		url := fmt.Sprintf("http://91zy.cc/list/?%d-%d.html", cate, index)
		list, err := getResourceList(url)
		if err != nil {
			fmt.Println("getResourceList", err)
			continue
		}

		for index, resourceUrl := range list {
			fmt.Println("caiji resource", index, resourceUrl)
			movie, err := getResource(resourceUrl, "chuantongshougong")
			if err != nil {
				fmt.Println("caiji failed !!", resourceUrl, err)
				continue
			}
			//fmt.Println(resource)
			db.DbInstance.Create(movie)
		}
		time.Sleep(time.Second * 2)
		fmt.Println("caiji the ", index, " page", "over!")
	}
}

func caijiCateUpdate(cate int) {
	for index := 1; index <= 2; index++ {
		fmt.Println("caiji the ", index, " page")
		url := fmt.Sprintf("http://91zy.cc/list/?%d-%d.html", cate, index)
		list, err := getResourceList(url)
		if err != nil {
			fmt.Println("getResourceList", err)
			return
		}

		for index, resourceUrl := range list {
			fmt.Println("采集资源:", index, resourceUrl)
			movie, err := getResource(resourceUrl, "chuantongshougong")
			if err != nil {
				fmt.Println("caiji failed !!", resourceUrl, err)
				continue
			}
			//fmt.Println(resource)
			db.DbInstance.Create(movie)
		}
		time.Sleep(time.Second * 2)
		fmt.Println("caiji the ", index, " page", "over!")
	}
}

func main() {
	method := flag.String("method", "update", "-method update")
	flag.Parse()
	db.Connect()
	if *method == "update" {
		for cateIndex := 1; cateIndex < 20; cateIndex++ {
			fmt.Printf("采集分类 %d \n", cateIndex)
			caijiCateUpdate(cateIndex)
		}
	} else {
		for cateIndex := 1; cateIndex < 20; cateIndex++ {
			caijiCate(cateIndex)
		}
	}
}
