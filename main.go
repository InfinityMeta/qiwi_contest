package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/html/charset"
)

type Currencies struct {
	XMLName  xml.Name
	Currency []struct {
		Numcode  string `xml:"NumCode"`
		Charcode string `xml:"CharCode"`
		Nominal  string `xml:"Nominal"`
		Name     string `xml:"Name"`
		Value    string `xml:"Value"`
	} `xml:"Valute"`
}

type Options struct {
	Code string
	Date string
}

func ParseFlags() (*Options, error) {
	var opts Options

	flag.StringVar(&opts.Code, "code", "USD", "currency code in format ISO 4217")
	flag.StringVar(&opts.Date, "date", "", "currency rates date in format YYYY-MM-DD")

	flag.Parse()

	return &opts, nil
}

const (
	YYYYMMDD = "2006-01-02"
	DDMMYYYY = "02/01/2006"
	CB_URL   = "http://www.cbr.ru/scripts/XML_daily.asp?date_req="
)

var (
	dateStr string
	code    string
	url     string
)

func main() {

	//Parse flags
	opts, err := ParseFlags()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "can not parse flags:", err)
		os.Exit(1)
	}

	//Convert date to format YYYY-MM-DD
	if opts.Date == "" {
		dateStr = time.Now().UTC().Format(DDMMYYYY)
	} else {
		dateTime, err := time.Parse(YYYYMMDD, opts.Date)
		if err != nil {
			fmt.Printf("error parsing the date:  %s\n", err)
			return
		}
		dateStr = dateTime.Format(DDMMYYYY)
	}

	code = opts.Code
	url = CB_URL + dateStr

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("error forming the request: %s\n", err)
		return
	}

	req.Header.Set("User-Agent", "MyUserAgent/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("error sending the request: %s\n", err)
		return
	}

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("error reading response body: %s\n", err)
		return
	}

	//Decode bytes to xml
	curs := new(Currencies)
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&curs)

	if err != nil {
		fmt.Printf("error decoding: %s\n", err)
		return
	}

	//Search for currency rate
	for i := range curs.Currency {
		if curs.Currency[i].Charcode == code {
			fmt.Println(curs.Currency[i].Charcode + " " + "(" + curs.Currency[i].Name + "):" + " " + curs.Currency[i].Value)
		}
	}

}
