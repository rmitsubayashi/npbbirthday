package main

import (
    "os"
    "strings"
    "strconv"
    "bytes"
    "errors"
    "unicode"
    "net/http"
    "log"
    "github.com/PuerkitoBio/goquery"
    "github.com/wcharczuk/go-chart"
)

func main() {
	response, httpError := http.Get("https://baseball-freak.com/birthday/")
	if httpError != nil {
	    log.Fatal(httpError)
	}
	defer response.Body.Close()

    dom, domError := goquery.NewDocumentFromReader(response.Body)
    if domError != nil {
        log.Fatal(domError)
    }

    dom.Find("table[class=birthday]").Each(
        func (i int, s *goquery.Selection){
            birthdayData := parseBirthdayTable(i, s)
            showBarChart(birthdayData)
        })

}

func parseBirthdayTable(i int, s *goquery.Selection) []int {
    birthdayCounter := make([]int, 12)
    s.Find("tr").Each(func (_i int, _s *goquery.Selection) {
        birthMonth, count, error := parseBirthdayTableRow(_i, _s)
        if error == nil {
            birthdayCounter[birthMonth-1] += count
        }
    })

    return birthdayCounter
}

func parseBirthdayTableRow(i int, s *goquery.Selection) (int, int, error) {
    birthday := s.Find("th").Text()
    if birthday == "" {
        // 「~月」って書いてあるヘッダー
        return 0, 0, errors.New("not valid row")
    } else {
        nameRow := s.Find("td").Text()
        names := strings.Split(nameRow, ")")
        count := 0
        for _, name := range names {
            if containsKatakana(name) {
                // 外国人（例外いっぱいあるけど許容範囲）
            } else {
                count ++
            }
        }
        birthMonthString := strings.Split(birthday, "月")[0]
        birthMonth, invalidNumStringError := strconv.Atoi(birthMonthString)

        if invalidNumStringError != nil {
            return 0, 0, invalidNumStringError
        } else {
            return birthMonth, count, nil
        }

    }
}

func containsKatakana(s string) bool {
    for _, r := range s {
        if unicode.In(r, unicode.Katakana) {
            return true
        }
    }

    return false
}

func showBarChart(birthdayData []int) {
    barValues := []chart.Value{}
    for index, val := range birthdayData {
        barVal := chart.Value{Value: float64(val), Label: strconv.Itoa(index+1)}
        barValues = append(barValues, barVal)
    }

    chartInfo := chart.BarChart {
        Title: "NPB Birthdays",
        Bars: barValues,
        XAxis:    chart.StyleShow(),
        YAxis: chart.YAxis{
            Style: chart.StyleShow(),
        },
        BaseValue: 0.0,
        UseBaseValue: true,
    }

    buffer := bytes.NewBuffer([]byte{})
    chartError := chartInfo.Render(chart.PNG, buffer)

    if chartError != nil {
        log.Fatal(chartError)
    }

    file, _ := os.Create("result.png")
    file.Write(buffer.Bytes())
}
