package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	linkParse "github.com/ap-pauloafonso/html-link-parse"
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

type sitemapxml struct {
	XMLName xml.Name `xml:"urlset"`
	Loc     []urlxml `xml:"url"`
	Xmlns   string   `xml:"xmlns,attr"`
}
type urlxml struct {
	Loc string `xml:"loc"`
}

func main() {
	urlFlag := flag.String("url", "", "valid url")
	depthFlag := flag.Int("depth", 3, "maximum search depth")
	flag.Parse()

	u, err := url.Parse(*urlFlag)
	if err != nil || *urlFlag == "" {
		panic("valid -url flag required")
	}

	var result = bfsNodes(strings.TrimSuffix(u.String(), "/"), *depthFlag)

	v := sitemapxml{Xmlns: xmlns}
	for _, val := range result {
		v.Loc = append(v.Loc, urlxml{Loc: val})
	}

	output, err := xml.MarshalIndent(v, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	os.Stdout.Write([]byte(xml.Header))
	os.Stdout.Write(output)
}

func getUrls(stringURL, domainURL string) []string {
	response, error := http.Get(stringURL)
	defer response.Body.Close()
	ret := []string{}
	if error == nil {
		htmlBytes, _ := ioutil.ReadAll(response.Body)
		childNodes := linkParse.Parse(string(htmlBytes))

		for _, val := range childNodes {
			ret = append(ret, normalize(domainURL, val.Url))
		}

	}

	filteredRet := filter(domainURL, ret)
	return filteredRet
}

func mapContainsKey(m map[string]struct{}, k string) bool {
	_, ok := m[k]
	return ok
}

func bfsNodes(stringURL string, depth int) []string {
	visited := make(map[string]struct{})
	iter := 0
	var f func(u []string)

	f = func(u []string) {
		iter++
		childs := map[string]struct{}{}
		for _, val := range u {
			if _, ok := visited[val]; ok {
				continue
			}
			visited[val] = struct{}{}
			filteredChilds := getUrls(val, stringURL)
			for _, child := range filteredChilds {
				if !mapContainsKey(visited, child) && !mapContainsKey(childs, child) {
					childs[child] = struct{}{}
				}
			}
		}
		ret := []string{}
		for k := range childs {
			ret = append(ret, k)
		}
		if len(ret) > 0 && iter < depth {
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

func filter(baseURL string, stringUrls []string) []string {
	ret := []string{}
	for _, val := range stringUrls {
		if strings.HasPrefix(val, baseURL) {
			ret = append(ret, val)
		}
	}

	return ret
}

func normalize(domain, candidate string) string {
	var result string
	switch {
	case strings.HasPrefix(candidate, "/"):
		result = domain + candidate
	case strings.HasPrefix(candidate, "#"):
		result = domain
	case strings.HasPrefix(candidate, "mailto:"):
		result = candidate
	default:
		result = candidate
	}

	return result
}
