package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aiomonitors/godiscord"
)

const thumbsDown = ":thumbsdown:"
const thumbsUp = ":thumbsup:"

var headers = map[string]string{
	"authority":                 "www.supremecommunity.com",
	"cache-control":             "max-age=0",
	"upgrade-insecure-requests": "1",
	"user-agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36",
	"sec-fetch-dest":            "document",
	"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
	"sec-fetch-site":            "none",
	"sec-fetch-mode":            "navigate",
	"sec-fetch-user":            "?1",
	"accept-language":           "en-US,en;q=0.9",
}

var client = &http.Client{}

type Dropitem struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Price       Price  `json:"price"`
	Votes       Votes  `json:"votes"`
	Link        string `json:"link"`
}

type Price struct {
	FullPrice string `json:"full_price"`
}

type Votes struct {
	Upvotes   string `json:"upvotes"`
	Downvotes string `json:"downvotes"`
}

type Droplist []Dropitem

func GetLatestDroplistLink() string {
	req, err := http.NewRequest("GET", "https://www.supremecommunity.com/season/spring-summer2021/droplists/", nil)
	if err != nil {
		fmt.Println(err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	link, _ := doc.Find("div.catalog-inner > div.week-list:first-child > div > a").Attr("href")

	latestdroplist := fmt.Sprintf("https://www.supremecommunity.com%s", link)
	return latestdroplist
}

func ScrapeDropList(link string) Droplist {

	var list Droplist
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		fmt.Println(err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	if err != nil {
		fmt.Println(err)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	doc.Find("div.catalog-item").Each(func(i int, s *goquery.Selection) {
		var Item Dropitem
		Item.Name, _ = s.Attr("data-name")
		if len(Item.Name) < 2 {
			fmt.Println("Products Name not loaded yet!")

		}
		Item.Name = strings.Replace(Item.Name, "®", "", -1)

		// this does not parse the img atm, i don't know why.
		img, imgExists := s.Find("div.catalog-item-top > div.catalog-item__thumb > img").Attr("src")

		if imgExists {
			Item.Image = fmt.Sprintf("https://www.supremecommunity.com%s", img)
		}

		categ, categExists := s.Attr("data-category")

		if categExists {
			Item.Category = strings.Title(categ)
		}
		if len(Item.Category) < 2 {
			Item.Category = "N/A"
		}

		desc, descExists := s.Find("div.catalog-item-top > div.catalog-item__thumb > img").Attr("alt")
		if descExists {
			Item.Description = desc
		}
		fmt.Println(Item.Description)

		productPrice, _ := s.Attr("data-gbpprice")
		Item.Price.FullPrice = fmt.Sprintf("£%v", productPrice)
		Item.Votes.Upvotes, _ = s.Attr("data-upvotes")
		Item.Votes.Downvotes, _ = s.Attr("data-downvotes")
		productLink, _ := s.Attr("href")
		Item.Link = fmt.Sprintf("https://www.supremecommunity.com%s", productLink)

		list = append(list, Item)

	})
	return list
}

func SendToWebHook(items Droplist, webhook string) {
	for _, item := range items {
		e := godiscord.NewEmbed(item.Name, item.Description, item.Link)
		e.SetThumbnail(item.Image)
		e.AddField("Price", item.Price.FullPrice, true)
		e.AddField("Category", item.Category, true)
		e.AddField("Votes", fmt.Sprintf("%v %v / %v %v", thumbsUp, item.Votes.Upvotes, thumbsDown, item.Votes.Downvotes), true)
		e.SetFooter("@RixtiRobotics", "")
		e.SetAuthor("Supreme Community", "https://www.supremecommunity.com/", "")
		e.SendToWebhook(webhook)
		time.Sleep(500 * time.Millisecond)
	}
}

func ConvertToJSON(items Droplist) ([]byte, error) {
	data, err := json.Marshal(items)
	fmt.Println(string(data))
	return data, err
}

func main() {
	lista := ScrapeDropList(GetLatestDroplistLink())
	SendToWebHook(lista, "https://discord.com/api/webhooks/831061966511669289/R6NYpR7utRSKCJrFKRUv5IxHFMLVfmybTdEhIdtQKlCrp8Hj62xUXQtWNMkvk3yJxciG")
}
