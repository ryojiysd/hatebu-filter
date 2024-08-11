package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type HatenaFeed struct {
    HatenaBookmarks []struct {
        Title       string `xml:"title"`
        Link        string `xml:"link"`
        Description string `xml:"description"`
        Date        string `xml:"date"`
        Count       int    `xml:"bookmarkcount"`
    } `xml:"item"`
}

type Item struct {
    Title string `xml:"title"`
    Link  string `xml:"link"`
    Desc  string `xml:"description"`
    Date  string `xml:"pubDate"`
}

type RSS2 struct {
    XMLName     xml.Name `xml:"rss"`
    Version     string   `xml:"version,attr"`
    Title       string   `xml:"channel>title"`
    Link        string   `xml:"channel>link"`
    Description string   `xml:"channel>description"`
    ItemList    []Item   `xml:"channel>item"`
}

type DenyList struct {
	Domains []string `json:"deny_domains"`
	Keywords []string `json:"deny_keywords"`
}

func main() {
	http.HandleFunc("/", handler)
    http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Header)

	rss := getRSS("http://b.hatena.ne.jp/hotentry/all.rss")

	feed := HatenaFeed{}
	err := xml.Unmarshal([]byte(rss), &feed)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	denyList := readDenyList()

	items := make([]Item, 0)
	for _, item := range feed.HatenaBookmarks {
		var deny = false
		for _, domain := range denyList.Domains {
			if strings.Contains(item.Link, domain) {
				deny = true
				break
			}
		}
		for _, keyword := range denyList.Keywords {
			if strings.Contains(item.Title, keyword) {
				deny = true
				break
			}
		}
		if deny {
			continue
		}
		items = append(items, Item{item.Title, item.Link, item.Description, item.Date})
	}

	// RSSの内容を設定 / 取得した記事追加
    newFeed := RSS2{
        Version:     "2.0",
        Title:       "はてなブックマーク - 人気エントリー - 総合 w/o anond",
        Link:        "https://b.hatena.ne.jp/hotentry/all",
        Description: "最近の人気エントリー w/o anond",
    }
    newFeed.ItemList = make([]Item, len(items))
    for i, item := range items {
        newFeed.ItemList[i].Title = item.Title
        newFeed.ItemList[i].Link = item.Link
        newFeed.ItemList[i].Desc = item.Desc
        newFeed.ItemList[i].Date = item.Date
    }

    // XMLに変換
    result, err := xml.MarshalIndent(newFeed, "  ", "    ")
    if err != nil {
        fmt.Printf("error: %v\n", err)
        return
    }

	// Webページに出力
    fmt.Fprint(w, "<?xml version='1.0' encoding='UTF-8'?>")
    fmt.Fprint(w, string(result))
}

func getRSS(url string) string {
    resp, err := http.Get("http://b.hatena.ne.jp/hotentry/all.rss")
    if err != nil {
        // エラーハンドリングを書く
    }
    defer resp.Body.Close()

    // _を使うことでエラーを無視できる
    body, _ := io.ReadAll(resp.Body)

    return string(body)
}

func readDenyList() DenyList {
	env := os.Getenv("DENY_LIST")
	if env == "" {
		return DenyList{}
	}
	var denyList DenyList
	err := json.Unmarshal([]byte(env), &denyList)
	if err != nil {
		return DenyList{}
	}
	return denyList
}

