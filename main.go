package main

import (
	"fmt"
	"log"

	"github.com/gocolly/colly"
	"github.com/kr/pretty"
)

type sub struct {
	desc string
	link string
}

var subs = []sub{}

func generateFormData() map[string]string {
	return map[string]string{
		"buscar": "smallville s01e08",
	}
}

func main() {
	c := colly.NewCollector(
		//Allowed domains
		colly.AllowedDomains("www.subdivx.com", "subdivx.com"),
	)
	log.Println("Starting collector")

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	c.OnHTML("div#buscador_detalle", func(e *colly.HTMLElement) {
		text := e.ChildText("#buscador_detalle_sub")
		link := e.ChildAttr("#buscador_detalle_sub_datos > a[href*='bajar']", "href")

		psub := sub{desc: text, link: link}

		subs = append(subs, psub)
	})

	c.Visit("https://subdivx.com/index.php?buscar=smallville+s01e06&accion=5&masdesc=&subtitulos=1&realiza_b=1")
	fmt.Printf("%# v", pretty.Formatter(subs))
}
