package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gocolly/colly"
	"github.com/kr/pretty"
)

type sub struct {
	desc string
	link string
}

var subs = []sub{}

//c.Visit("https://subdivx.com/index.php?buscar=smallville+s01e06&accion=5&masdesc=&subtitulos=1&realiza_b=1")
func generateURL() string {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = "subdivx.com"
	u.Path = "index.php"
	q := u.Query()
	q.Set("buscar", "smallville s01e06")
	q.Set("accion", "5")
	q.Set("masdesc", "")
	q.Set("subtitulos", "1")
	q.Set("realiza_b", "1")
	u.RawQuery = q.Encode()
	return u.String()
}

func main() {
	generateURL()
	c := colly.NewCollector(
		//Allowed domains
		colly.AllowedDomains("www.subdivx.com", "subdivx.com"),
	)
	log.Println("Starting collector")

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	/*
		c.OnResponse(func(r *colly.Response) {
			log.Println(string(r.Body))
		})
	*/
	c.OnHTML("div#buscador_detalle", func(e *colly.HTMLElement) {
		text := e.ChildText("#buscador_detalle_sub")
		link := e.ChildAttr("#buscador_detalle_sub_datos > a[href*='bajar']", "href")

		psub := sub{desc: text, link: link}

		subs = append(subs, psub)
	})
	c.Visit(generateURL())
	c.Wait()
	fmt.Printf("%# v", pretty.Formatter(subs))
}
