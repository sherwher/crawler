package crawler

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/gographics/imagick/imagick"
)

func save() error {

	return nil
}

func GetHtmlCandidate(city int, sgg string) (string, error) {
	rurl := "http://info.nec.go.kr/electioninfo/electionInfo_report.xhtml"

	res, err := http.PostForm(rurl, url.Values{
		"electionId":   {"0020200415"},
		"requestURI":   {"/WEB-INF/jsp/electioninfo/0020200415/pc/pcri03_ex.jsp"},
		"topMenuId":    {"PC"},
		"secondMenuId": {"PCRI03"},
		"menuId":       {"PCRI03"},
		"statementId":  {"PCRI03_#2"},
		"electionCode": {"2"},
		"cityCode":     {strconv.Itoa(city)},
		"sggCityCode":  {sgg},
		"townCode":     {"-1"},
		"sggTownCode":  {"0"},
		"x":            {"30"},
		"y":            {"8"},
	})
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

func CrawlerCandidate(city int, sgg string) ([]*Info, error) {
	html, err := GetHtmlCandidate(city, sgg)
	if err != nil {
		return nil, err
	}
	doc, err := htmlquery.Parse(strings.NewReader(html))
	list := htmlquery.Find(doc, "//tr")
	var Infos []*Info
	for _, val := range list {
		tr := htmlquery.Find(val, "//td")
		var tds = make([]string, len(tr))
		for idx, td := range tr {
			tds[idx] = strings.TrimSpace(htmlquery.InnerText(td))
		}

		var electionId, key string
		a := htmlquery.FindOne(val, "//a")
		if a != nil {
			s := strings.Split(htmlquery.SelectAttr(a, "href"), "'")
			electionId = s[1]
			key = s[3]
		}

		var imageUrl string
		img := htmlquery.FindOne(val, "//input")
		if img != nil {
			// log.Println(img.Attr[1])
			// log.Println(htmlquery.SelectAttr(img, "src"))
			imageUrl = "http://info.nec.go.kr/" + htmlquery.SelectAttr(img, "src")
			filepaths := strings.Split(htmlquery.SelectAttr(img, "src"), "/")
			filename := filepaths[len(filepaths)-1]
			changePath := "./image/people/" + strconv.Itoa(city) + "/" + sgg + "/" + key
			err := os.MkdirAll(changePath, 0777)
			if err != nil {
				log.Println(err)
			}
			if err := DownloadFile(changePath+"/"+filename, imageUrl); err != nil {
				log.Println(err)
				return nil, err
			}

			imageUrl = changePath + "/" + filename
		}

		CriminalrecordURL, err := getCriminalImage(strconv.Itoa(city), sgg, electionId, key)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		if len(tds) != 0 && len(tds) != 1 {
			info := &Info{
				Local:                 tds[0],
				Party:                 tds[1],
				Name:                  tds[3],
				Gender:                tds[4],
				Birth:                 tds[5],
				Address:               tds[6],
				Job:                   tds[7],
				Educationalbackground: tds[8],
				Career:                tds[9],
				Criminalrecord:        tds[10],
				Registdate:            tds[11],
				Image:                 imageUrl,
				City:                  strconv.Itoa(city),
				Sgg:                   sgg,
				ElectionId:            electionId,
				Key:                   key,
				CriminalrecordURL:     CriminalrecordURL,
			}
			log.Println(info)
			Infos = append(Infos, info)
		}
	}
	return Infos, err
}

func GetSggCity(city int) ([]*SubAddressData, error) {
	rurl := "http://info.nec.go.kr/bizcommon/selectbox/selectbox_getSggCityCodeJson.json"
	res, err := http.PostForm(rurl, url.Values{
		"electionId":   {"0020200415"},
		"electionCode": {"2"},
		"cityCode":     {strconv.Itoa(city)},
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var item struct {
		JSONResult struct {
			Body []*SubAddressData `json:"body"`
		} `json:"jsonResult"`
	}
	if err := json.Unmarshal(body, &item); err != nil {
		log.Println(err)
		return nil, err
	}
	return item.JSONResult.Body, nil
}

func GetCity() ([]*AddressData, error) {
	rurl := "http://info.nec.go.kr/bizcommon/selectbox/selectbox_cityCodeBySgJson.json"

	res, err := http.PostForm(rurl, url.Values{
		"electionId":   {"0020200415"},
		"electionCode": {"2"},
	})

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var item struct {
		JSONResult struct {
			Body []*AddressData `json:"body"`
		} `json:"jsonResult"`
	}
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, err
	}
	return item.JSONResult.Body, nil
}

func getCriminalImage(City string, Sgg string, ElectionId string, Key string) ([]string, error) {
	var Pdffile string
	pdfs, err := GetPDF(ElectionId, Key)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var paths []string

	imagick.Initialize()
	defer func() {
		imagick.Terminate()
	}()

	for _, pdf := range pdfs {
		path := pdf.FILEPATH
		filepaths := strings.Split(path, "/")
		filename := strings.Replace(filepaths[len(filepaths)-1], ".tif", ".PDF", 1)
		changePath := "./image/criminal/" + City + "/" + Sgg + "/" + Key
		changePDFPath := changePath + "/" + filename + ".PDF"
		changeImagePath := changePath + "/" + filename + ".jpg"
		err := os.MkdirAll(changePath, 0777)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		Pdffile = "http://info.nec.go.kr/unielec_pdf_file/" + strings.Replace(path, ".tif", ".PDF", 1)
		if err := DownloadFile(changePDFPath, Pdffile); err != nil {
			log.Println(err)
			return nil, err
		}

		mw := imagick.NewMagickWand()
		defer func() {
			mw.Destroy()
		}()

		mw.SetResolution(300, 300)

		if err := mw.ReadImage(changePDFPath); err != nil {
			log.Println(err)
			return nil, err
		}

		mw.SetIteratorIndex(0) // This being the page offset
		if err := mw.SetImageFormat("jpg"); err != nil {
			log.Println("2", err)
			return nil, err
		}

		if err := mw.WriteImage(changeImagePath); err != nil {
			log.Println("3", err)
			return nil, err
		}
		paths = append(paths, changeImagePath)
		if err := os.Remove(changePDFPath); err != nil {
			log.Panicln(err)
			return nil, err
		}
	}
	return paths, nil
}

func GetPDF(electionId string, key string) ([]*PDF, error) {
	rurl := "http://info.nec.go.kr/electioninfo/candidate_detail_scanSearchJson.json"
	res, err := http.PostForm(rurl, url.Values{
		"gubun":       {"5"},
		"electionId":  {electionId},
		"huboId":      {key},
		"statementId": {"PCRI03_candidate_scanSearch"},
	})
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var item struct {
		JSONResult struct {
			Body []*PDF `json:"body"`
		} `json:"jsonResult"`
	}

	if err := json.Unmarshal(body, &item); err != nil {
		log.Println(err)
		return nil, err
	}

	return item.JSONResult.Body, err
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		log.Println(err)
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Println(err)
		return err
	}
	return err
}

type PDF struct {
	SAJINPATH    string `json:"SAJINPATH"`
	SGTYPECODE   int    `json:"SG_TYPECODE"`
	SGGNAME      string `json:"SGGNAME"`
	HBJNAME      string `json:"HBJNAME"`
	FILEGUBUN    string `json:"FILE_GUBUN"`
	SUBSGID      int    `json:"SUB_SG_ID"`
	FILEPATH     string `json:"FILEPATH"`
	HUBOID       int    `json:"HUBOID"`
	DISPSEQ      int    `json:"DISP_SEQ"`
	JDNAME       string `json:"JDNAME"`
	HBJHANJANAME string `json:"HBJHANJANAME"`
	SGNAME       string `json:"SG_NAME"`
}

type AddressData struct {
	Code int    `json:"CODE"`
	Name string `json:"NAME"`
}

type SubAddressData struct {
	Code string `json:"CODE"`
	Name string `json:"NAME"`
}

type Info struct {
	Local                 string
	Party                 string
	Name                  string
	Gender                string
	Birth                 string
	Address               string
	Job                   string
	Educationalbackground string
	Career                string
	Criminalrecord        string
	Registdate            string
	Image                 string
	City                  string
	Sgg                   string
	ElectionId            string
	Key                   string
	CriminalrecordURL     []string
}
