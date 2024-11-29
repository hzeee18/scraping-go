package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly"
)

type Rule struct {
	Title       string
	SubTitle    string
	Description string
	Nomor       string
	Tentang     string
	PostAuthor  string
}

func main() {
	var links []string

	c := colly.NewCollector()

	c.OnHTML(".post-index a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		links = append(links, link)
	})

	// Callback untuk mencari dan mengunjungi halaman "next page"
	c.OnHTML(".blog-pager .prev a", func(e *colly.HTMLElement) {
		nextPageLink := e.Attr("href")
		if nextPageLink != "" {
			c.Visit(e.Request.AbsoluteURL(nextPageLink))
		}
	})

	// Kunjungi halaman awal yang ingin di-scrape
	err := c.Visit("https://www.peraturanpolri.com/")
	if err != nil {
		fmt.Println("Error visiting:", err)
		return
	}

	fmt.Println(len(links))

	// Loop untuk mengunjungi tautan-tautan yang ada dalam slice links
	for _, link := range links {
		// Membuat instance baru dari Collector untuk setiap iterasi loop
		c := colly.NewCollector()

		// Variable untuk menyimpan link PDF
		var pdfLink string

		// Callback untuk scraping link PDF dari halaman
		c.OnHTML(".perkap .download[href]", func(e *colly.HTMLElement) {
			pdfLink = e.Attr("href")
		})

		// Callback untuk scraping isi dari halaman
		c.OnHTML(".hfeed", func(e *colly.HTMLElement) {

			values := &Rule{
				Title:       e.ChildText(".title-post"),
				SubTitle:    e.ChildText(".em"),
				Description: e.ChildText(".post-body p"),
				Nomor:       e.ChildText(".perkap li:nth-child(1)"),
				Tentang:     e.ChildText(".perkap li:nth-child(2)"),
				PostAuthor:  e.ChildText(".post-info .post-author span[itemprop]"),
			}

			file, contentType, err := downloadFile(pdfLink)
			if err != nil {
				fmt.Println(e.ChildText(".title-post"))
				fmt.Println("Failed to get link")
			}
			fmt.Println(pdfLink)

			fmt.Println(contentType)

			saveData(values, link, file, contentType)
		})

		err := c.Visit(link)
		if err != nil {
			fmt.Println("Error visiting:", err)
		}
	}
}

func saveData(rule *Rule, link string, pdfContent []byte, pdfType string) {
	// Mengambil bagian terakhir dari link sebagai nama folder
	folderName := filepath.Base(link)

	// Menghilangkan ekstensi ".html" dari nama folder jika ada
	folderName = strings.TrimSuffix(folderName, ".html")

	// Mengganti karakter '/' pada link dengan karakter '-' untuk digunakan sebagai nama folder
	folderName = strings.ReplaceAll(folderName, "/", "-")

	// Membuat folder baru dengan nama folderName dalam folder "export"
	err := os.MkdirAll(filepath.Join("export", folderName), 0755)
	if err != nil {
		fmt.Println("Error creating folder:", err)
		return
	}

	// Mengambil bagian terakhir dari link sebagai nama file JSON
	filename := filepath.Base(link)

	// Menghilangkan ekstensi ".html" dari nama file JSON jika ada
	filename = strings.TrimSuffix(filename, ".html")

	// Mengganti karakter '/' pada link dengan karakter '-' untuk digunakan sebagai nama file JSON
	filename = strings.ReplaceAll(filename, "/", "-")

	// Marshal data ke dalam format JSON
	jsonData, err := json.Marshal(rule)
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return
	}

	// Tulis data JSON ke dalam file .json
	err = ioutil.WriteFile(filepath.Join("export", folderName, filename+".json"), jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON data to file:", err)
		return
	}

	// Simpan file content (PDF) ke dalam file dengan ekstensi .pdf
	err = ioutil.WriteFile(filepath.Join("export", folderName, filename+pdfType), pdfContent, 0644)
	if err != nil {
		fmt.Println("Error writing PDF data to file:", err)
		return
	}
}

func downloadFile(url string) ([]byte, string, error) {
	// Mendapatkan response dari URL
	response, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	// Membaca konten file dari response body
	fileContent, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}

	// Dapatkan jenis MIME dari header "Content-Type"
	contentType := response.Header.Get("Content-Type")
	if contentType == "application/pdf" {
		contentType = ".pdf"
	} else {
		contentType = ".doc"
	}
	return fileContent, contentType, nil
}
