package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	linkParse "github.com/ap-pauloafonso/html-link-parse"
)

func main() {
	url := flag.String("url", "default value", "here comes usage")
	flag.Parse()

	var result = bfsNodes(*url, 0)

	arr := []string{}
	v := sitemapxml{}

	for _, val := range result {
		arr = append(arr, val)
		v.Loc = append(v.Loc, urlxml{Loc: val})
	}
	// fmt.Printf("%+v/n", arr)

	output, err := xml.MarshalIndent(v, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	os.Stdout.Write(output)
}

func getUrls(stringURL string) []string {
	response, error := http.Get(stringURL)
	defer response.Body.Close()
	ret := []string{}
	if error == nil {
		htmlBytes, _ := ioutil.ReadAll(response.Body)
		childNodes := linkParse.Parse(string(htmlBytes))

		for _, val := range childNodes {
			ret = append(ret, val.Url)
		}

	}
	return ret
}

// func parse(link []linkParse.Link) map[string]linkParse.Link {
// 	domainURL := link[0].Url
// 	result := map[string]linkParse.Link{}
// 	var f func(link []linkParse.Link)
// 	iter := 0
// 	f = func(link []linkParse.Link) {
// 		iter++
// 		fmt.Println(iter)
// 		childs := map[string]linkParse.Link{}
// 		for _, item := range link {
// 			if _, ok := result[item.Url]; !ok {
// 				result[item.Url] = item
// 				response, error := http.Get(item.Url)
// 				if error == nil {
// 					htmlBytes, _ := ioutil.ReadAll(response.Body)
// 					childNodes := linkParse.Parse(string(htmlBytes))
// 					validChilds := filterUrls(domainURL, childNodes)
// 					for _, child := range validChilds {
// 						if _, a := childs[child.Url]; !a {
// 							if _, b := result[child.Url]; !b {
// 								childs[child.Url] = child
// 							}
// 						}
// 					}
// 				}

// 			}

// 		}
// 		arr := make([]linkParse.Link, 0, len(childs))
// 		for _, val := range childs {
// 			arr = append(arr, val)
// 		}
// 		if len(arr) > 0 {
// 			f(arr)
// 		}
// 	}

// 	f(link)
// 	return result
// }

//Node is struct to help traverse the tree
type Node struct {
	URL       string
	ChildURLS []string
}

func bfsNodes(stringURL string, depth int) []string {
	visited := make(map[string]struct{})
	iter := 0
	var f func(u []string)

	f = func(u []string) {
		iter++
		fmt.Println(iter)
		childs := map[string]struct{}{}
		for _, val := range u {
			if _, ok := visited[val]; ok {
				continue
			}
			visited[val] = struct{}{}
			filteredChilds := filterUrls(stringURL, getUrls(val))
			for _, child := range filteredChilds {
				if _, ok1 := visited[child]; !ok1 {
					if _, ok2 := childs[child]; !ok2 {
						childs[child] = struct{}{}
					}
				}
			}
		}
		ret := []string{}
		for k := range childs {
			ret = append(ret, k)
		}
		if len(ret) > 0 {
			f(ret)
		}

	}

	f([]string{stringURL})

	result := []string{}
	for k := range visited {
		result = append(result, k)
	}
	return result

}

func filterUrls(source string, links []string) []string {
	result := []string{}

	for _, item := range links {
		if strings.HasPrefix(item, "#") || item == "/" || item == "" || strings.HasPrefix(item, "m") {
			continue
		}

		normalizedURL := normalizeURL(item, source)

		if !isDomainURL(normalizedURL, source) {
			continue
		}

		result = append(result, normalizedURL)
	}

	return result

}
func isDomainURL(url, domainURL string) bool {

	return strings.HasPrefix(strings.Split(url, "http")[1], strings.Split(domainURL, "http")[1])
}
func needsCorrection(str, baseURL string) bool {
	if strings.HasPrefix(str, "http") {
		return false
	}
	return !strings.HasPrefix(str, baseURL)
}
func normalizeURL(url, baseURL string) string {
	if !needsCorrection(url, baseURL) {
		return url
	}
	str := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(url, "/")

	return strings.Split(str, "#")[0]
}

type sitemapxml struct {
	XMLName xml.Name `xml:"urlset"`
	Loc     []urlxml `xml:"url"`
}
type urlxml struct {
	Loc string `xml:"loc"`
}
