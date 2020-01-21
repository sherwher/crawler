package crawler

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/antchfx/htmlquery"
)

func GetHtmlParty() (string, error) {
	rurl := "http://www.nec.go.kr/portal/bbs/list/B0000350.do?menuNo=200476"

	res, err := http.Get(rurl)
	if err != nil {
		log.Println(err)
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return string(body), nil
}

func CrawlerParty() ([]*Party, error) {
	html, err := GetHtmlParty()
	if err != nil {
		return nil, err
	}
	doc, err := htmlquery.Parse(strings.NewReader(html))
	list := htmlquery.Find(doc, "//tr[@align]")
	var Partys []*Party
	for _, val := range list {
		tr := htmlquery.Find(val, "//td")
		var tds = make([]string, len(tr))
		for idx, td := range tr {
			tds[idx] = strings.TrimSpace(htmlquery.InnerText(td))
		}
		if len(tds) != 0 && len(tds) != 1 {
			party := &Party{
				ParyName:        tds[0],
				RegistDate:      tds[1],
				PartyRepresent:  tds[2],
				PartyAddress:    tds[3],
				PartyCallNumber: tds[4],
			}
			log.Println(party)
			Partys = append(Partys, party)
		}
	}
	return Partys, err
}

type Party struct {
	ParyName        string
	RegistDate      string
	PartyRepresent  string
	PartyAddress    string
	PartyCallNumber string
}
