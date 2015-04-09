package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

const LCD_COLS = 20

const CHARSET_DEFAULT = 0x18
const CHARSET_KANA = 0x19

var state_weather string

type LCDResponse struct {
	Line      []string
	Luminance int
	Charset   int
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	go update_weather()

	http.HandleFunc("/", HTTPHandler)
	http.ListenAndServe(":1081", nil)
}

func justify(in string) string {
	// If the length of in < LCD_COLS, justify the string
	// within the cols
	if len(in) >= LCD_COLS {
		return in
	}
	return fmt.Sprintf("%s%s", strings.Repeat(" ", (LCD_COLS-len(in))/2), in)
}

func update_weather() {
	for {
		new_weather, err := getWeatherLine()
		if err == nil {
			state_weather = new_weather
			time.Sleep(time.Minute)
		} else {
			log.Println(err)
		}
	}

}

func HTTPHandler(w http.ResponseWriter, r *http.Request) {
	resp := new(LCDResponse)

	// Set luminance
	resp.Luminance = 0x00

	// Set charset
	resp.Charset = CHARSET_KANA

	resp.Line = make([]string, 4)
	// First line date & time
	resp.Line[0] = justify(time.Now().Format("Mon Jan _2 15:04"))

	// Second line email count
	unread := get_num_unread_emails()
	if unread > 0 {
		if unread > 1 {
			resp.Line[1] = fmt.Sprintf("%d new emails", unread)
		} else {
			resp.Line[1] = fmt.Sprintf("%d new email", unread)
		}
	} else {
		resp.Line[1] = "No new emails"
	}
	resp.Line[1] = justify(resp.Line[1])

	// 3rd line BTC ticker, trader profit
	ticker := NewTickerBTC()
	stats, err := ticker.GetPrice()
	profit, err := getTraderProfit()
	if err != nil {
		log.Println(err)
	} else {
		direction := '+'
		resp.Line[2] = justify(
			fmt.Sprintf("BTC %.1f %c$%.1f", stats.Sell,
				direction,
				math.Abs(profit)))
	}

	// Fourth line weather
	resp.Line[3] = justify(state_weather)

	json, err := json.Marshal(resp)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Fprint(w, string(json))
}
