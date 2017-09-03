package internal

import (
	"gopcpv2-web-spider/module"
	"net/http"
	"path"
	"net/url"
	"fmt"
	"strings"
	"github.com/PuerkitoBio/goquery"
)

func genResponsePasers()[]module.ParseResponse{
	parseLink := func(httpResp *http.Response, respDepth uint32)([]module.Data, []error){
		dataList := make([]module.Data, 0)
		if httpResp == nil{
			return nil, []error{fmt.Errorf("http响应是nil")}
		}
		httpReq := httpResp.Request
		if httpReq == nil{
			return nil, []error{fmt.Errorf("http请求时nil")}
		}
		if httpResp.StatusCode != 200{
			return nil, []error{fmt.Errorf("状态码不是200， code是%d, url%s", httpResp.StatusCode, httpReq.URL)}
		}
		if httpResp.Body == nil{
			return nil, []error{fmt.Errorf("HTTP响应的body是nil， url：%s", httpReq.URL)}
		}
		var matchedContentType bool
		if httpResp.Header != nil{
			contentTypeList := httpResp.Header["Content-Type"]
			for _, contentType := range contentTypeList{
				if strings.HasPrefix(contentType, "text/html"){
					matchedContentType = true
					break
				}
			}
		}
		if !matchedContentType{
			return dataList, nil
		}
		doc, err := goquery.NewDocumentFromReader(httpResp.Body)
		if err != nil{
			return dataList, []error{err}
		}
		errList := make([]error, 0)
		doc.Find("a").Each(
			func(index int, sel *goquery.Selection){
				href, exists := sel.Attr("href")
				if !exists || href == "" || href == "#" || href == "/"{
					return
				}
				href = strings.TrimSpace(href)
				lowerHref := strings.ToLower(href)
				if href == "" || strings.HasPrefix(lowerHref, "javascript"){
					return
				}
				aUrl, err := url.Parse(href)
				if err != nil{
					return
				}
				if !aUrl.IsAbs(){
					aUrl = httpReq.URL.ResolveReference(aUrl)
				}
				httpReq, err := http.NewRequest("GET", aUrl.String(), nil)
				if err != nil{
					errList = append(errList, err)
				}else{
					req := module.NewRequest(httpReq, respDepth)
					dataList = append(dataList, req)
				}
			},
		)
		// 查找img标签并提取地址。
		doc.Find("img").Each(func(index int, sel *goquery.Selection) {
			// 前期过滤。
			imgSrc, exists := sel.Attr("src")
			if !exists || imgSrc == "" || imgSrc == "#" || imgSrc == "/" {
				return
			}
			imgSrc = strings.TrimSpace(imgSrc)
			imgURL, err := url.Parse(imgSrc)
			if err != nil {
				errList = append(errList, err)
				return
			}
			if !imgURL.IsAbs() {
				imgURL = httpReq.URL.ResolveReference(imgURL)
			}
			httpReq, err := http.NewRequest("GET", imgURL.String(), nil)
			if err != nil {
				errList = append(errList, err)
			} else {
				req := module.NewRequest(httpReq, respDepth)
				dataList = append(dataList, req)
			}
		})
		return dataList, errList
	}

	parseImg := func(httpResp *http.Response, respDepth uint32)([]module.Data, []error){
		// 检查响应。
		if httpResp == nil {
			return nil, []error{fmt.Errorf("HTTP响应是nil")}
		}
		httpReq := httpResp.Request
		if httpReq == nil {
			return nil, []error{fmt.Errorf("HTTP请求是nil")}
		}
		if httpResp.StatusCode != 200 {
			return nil, []error{fmt.Errorf("状态码不是200， code是%d, url%s", httpResp.StatusCode, httpReq.URL)}
		}
		httpRespBody := httpResp.Body
		if httpRespBody == nil {
			return nil, []error{fmt.Errorf("HTTP响应的body是nil， url：%s", httpReq.URL)}
		}
		// 检查HTTP响应头中的内容类型。
		dataList := make([]module.Data, 0)
		var pictureFormat string
		if httpResp.Header != nil {
			contentTypeList := httpResp.Header["Content-Type"]
			var contentType string
			for _, ct := range contentTypeList {
				if strings.HasPrefix(ct, "image") {
					contentType = ct
					break
				}
			}
			index1 := strings.Index(contentType, "/")
			index2 := strings.Index(contentType, ";")
			if index1 > 0 {
				if index2 < 0 {
					pictureFormat = contentType[index1+1:]
				} else if index1 < index2 {
					pictureFormat = contentType[index1+1 : index2]
				}
			}
		}
		if pictureFormat == "" {
			return dataList, nil
		}
		// 生成条目。
		item := make(map[string]interface{})
		item["reader"] = httpRespBody
		item["name"] = path.Base(httpReq.URL.Path)
		item["ext"] = pictureFormat
		dataList = append(dataList, module.Item(item))
		return dataList, nil
	}
	return []module.ParseResponse{parseLink, parseImg}
}