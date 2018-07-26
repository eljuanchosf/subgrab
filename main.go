package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
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

var debug = ""
var dryrun = false

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
		Help    bool   `short:"h" long:"help"                     description:"Display usage" global:"true"`
		Version bool   `short:"v" long:"version"                  description:"Display version"`
		DryRun  bool   `short:"d" long:"dryrun"                   description:"Do not make changes to the filesystem"`
		Infile  string `short:"i" long:"infile"  required:"false" description:"Input video filename for subtitle smart search. Has to be single or double quoted"`
		Outfile string `short:"o" long:"outfile" required:"false" description:"Output filename for the subtitle"`
		Lang    string `short:"l" long:"lang"    required:"false" description:"Preferred language for download. Also sets the output name"`
		Regex   string `short:"r" long:"regex"   required:"false" description:"Regex pattern to apply to the Infile for seaching the subtitle. Must be used between quotes or double quotes"`
		Debug   bool   `long:"b"                                  description:"Enable debugging"`
		DebugEx bool   `long:"bb"                                 description:"Enable extended debugging"`
		Grab    struct {
			Settings bool `settings:"true" allow-unknown-arg:"true"`
		} `command:"grab" description:"Grabs a subtitle based on the provided arguments"`
	}{}

	cmd, err := gocmd.New(gocmd.Options{
		Name:        "subgrab",
		Version:     "0.0.1",
		Description: "A smart (well, not so much right now) subtitle grabber",
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
		debug = "b"
		fmt.Println("WARNING: DEBUG ENABLED!")
	} else if cmd.FlagArgs("DebugEx") != nil {
		debug = "bb"
		fmt.Println("WARNING: **EXTENDED** DEBUG ENABLED!")
	}

	// Set Dryrun
	if flags.DryRun {
		dryrun = true
		fmt.Println("RUNNING IN DRYRUN")
	}

	// Lang parameter
	lang := flags.Lang
	if debug == "b" {
		fmt.Println("Lang:", lang)
	}

	// Infile parameter
	infile := flags.Infile
	if debug == "b" {
		fmt.Println("Infile:", infile)
	}
	regex := flags.Regex
	r := regexp.MustCompile(regex)
	if debug == "b" {
		fmt.Println("Regex:", regex)
	}

	// Grab command
	if cmd.FlagArgs("Grab") != nil {
		args := strings.Trim(fmt.Sprintf("%s\n", strings.TrimRight(strings.TrimLeft(fmt.Sprintf("%v",
			cmd.FlagArgs("Grab")[1:]), "["), "]")), "\n")

		if debug == "b" {
			fmt.Println("Grab:", cmd.FlagArgs("Grab"))
			fmt.Println("Args:", args)
		}

		// Get the files to process
		matches, err := filepath.Glob(infile)
		if err != nil {
			fmt.Println(err)
		}
		if debug == "b" {
			fmt.Println("Directory matches:", matches)
		}
		for _, file := range matches {

			if debug == "b" {
				fmt.Println("File:", file)
			}

			keywords := args + " " + fmt.Sprintf("%v", r.FindString(file))
			// Outfile parameter
			outfile := ""
			filename := ""
			if file != "" {
				filename = file[0 : len(file)-len(filepath.Ext(file))]
			} else {
				outfile = flags.Outfile
				filename = outfile[0 : len(outfile)-len(filepath.Ext(outfile))]
			}

			if lang != "" {
				outfile = filename + "." + lang + ".srt"
			} else {
				outfile = filename + ".srt"
			}

			if debug == "b" {
				fmt.Println("Outfile:", outfile)
			}
			subs = []sub{}
			grabSub(keywords, outfile)
		}
	}
}

func grabSub(args string, outfile string) {
	if debug == "b" {
		fmt.Println("grabSub.args", args)
		fmt.Println("grabSub.outfile", outfile)
	}
	c := colly.NewCollector(
		//Allowed domains
		colly.AllowedDomains("www.subdivx.com", "subdivx.com"),
	)
	log.Println("Starting collector")

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		if debug == "bb" {
			log.Println(string(r.Body))
		}
	})

	c.OnHTML("div#buscador_detalle", func(e *colly.HTMLElement) {
		text := e.ChildText("#buscador_detalle_sub")
		link := e.ChildAttr("#buscador_detalle_sub_datos > a[href*='bajar']", "href")

		psub := sub{desc: text, link: link}

		subs = append(subs, psub)
	})

	url := generateURL(args)
	if debug == "b" {
		fmt.Println("URL:", url)
	}

	if !dryrun {
		c.Visit(url)
		c.Wait()

		if debug == "bb" {
			fmt.Println("Possible subtitles list")
			fmt.Printf("%# v", pretty.Formatter(subs))
		}
		zipFile := "tmp.zip"
		log.Println("Getting sub for", subs[0].desc)
		downloadFile(zipFile, subs[0].link)
		log.Println("Unzipping and deleting file", zipFile)
		files, err := unzip(zipFile, ".", true)
		if err != nil {
			log.Fatal(err)
		}
		if outfile != "" {
			log.Println("Saving to", outfile)
			os.Rename(files[0], outfile)
		}
	}
}

// downloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func unzip(src string, dest string, deleteZipFile bool) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)
		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	if deleteZipFile {
		os.Remove(src)
	}
	return filenames, nil
}
