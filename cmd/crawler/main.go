package main

import (
	"log"

	"github.com/sherwher/crawler/crawler"
)

func main() {

	// 정당정보
	// body, err := crawler.CrawlerParty()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// log.Println(body)
	// return

	addrs, err := crawler.GetCity()
	if err != nil {
		log.Println(err)
		return
	}
	count := 0
	var arr [][]*crawler.Info
	for _, v := range addrs {
		sggaddrs, err := crawler.GetSggCity(v.Code)
		if err != nil {
			log.Println(err)
			return
		}
		for _, sgg := range sggaddrs {
			Infos, err := crawler.CrawlerCandidate(v.Code, sgg.Code)
			if err != nil {
				log.Println(err)
			}
			arr = append(arr, Infos)
			count++
		}
	}

	log.Println("count", count)
}
