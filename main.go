package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/devfacet/gocmd"
	"github.com/gocolly/colly"
	"github.com/kr/pretty"
)

type sub struct {
	desc string
	link string
}

var subs = []sub{}

var debug = false

//c.Visit("https://subdivx.com/index.php?buscar=smallville+s01e06&accion=5&masdesc=&subtitulos=1&realiza_b=1")
func generateURL(args string) string {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = "subdivx.com"
	u.Path = "index.php"
	q := u.Query()
	q.Set("buscar", args)
	q.Set("accion", "5")
	q.Set("masdesc", "")
	q.Set("subtitulos", "1")
	q.Set("realiza_b", "1")
	u.RawQuery = q.Encode()
	return u.String()
}

func main() {

	flags := struct {
		Help    bool `short:"h" long:"help" description:"Display usage" global:"true"`
		Version bool `short:"v" long:"version" description:"Display version"`
		Debug   bool `short:"b" long:"debub" description:"Enable debugging"`
		Grab    struct {
			Settings bool `settings:"true" allow-unknown-arg:"true"`
		} `command:"grab" description:"Grabs a subtitle based on the provided arguments"`
	}{}

	cmd, err := gocmd.New(gocmd.Options{
		Name:        "subgrab",
		Version:     "0.0.1",
		Description: "A smart (well, not su much right now) subtitle grabber",
		Flags:       &flags,
		AutoHelp:    true,
		AutoVersion: true,
		AnyError:    true,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Set Debug
	if cmd.FlagArgs("Debug") != nil {
		debug = true
		fmt.Println("WARNING: DEBUG ENABLED!")
	}

	// Grab command
	if cmd.FlagArgs("Grab") != nil {
		args := strings.Trim(fmt.Sprintf("%s\n", strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%v",
			cmd.FlagArgs("Grab")[1:]), "["), "]")), "\n")
		grabSub(args)
	}
}

func grabSub(args string) {
	c := colly.NewCollector(
		//Allowed domains
		colly.AllowedDomains("www.subdivx.com", "subdivx.com"),
	)
	log.Println("Starting collector")

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		if debug {
			log.Println(string(r.Body))
		}
	})

	c.OnHTML("div#buscador_detalle", func(e *colly.HTMLElement) {
		text := e.ChildText("#buscador_detalle_sub")
		link := e.ChildAttr("#buscador_detalle_sub_datos > a[href*='bajar']", "href")

		psub := sub{desc: text, link: link}

		subs = append(subs, psub)
	})
	c.Visit(generateURL(args))
	c.Wait()
	fmt.Printf("%# v", pretty.Formatter(subs))
}
