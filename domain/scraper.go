package domain

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/labstack/echo/v4"
	"github.com/lucasarieta/gcs-scraper/model"
)

type ScraperDomain struct{}

func (s *ScraperDomain) SetupRoutes(routes *echo.Group) {
	routes.GET("/scraper", s.Scrape)
}

func (s *ScraperDomain) Scrape(c echo.Context) error {
	url := "https://sarc.pucrs.br/Default/Export.aspx?id=c9472a18-8bbb-4e42-b4fa-6e9c6fcecf99&ano=2024&sem=2"
	resp, err := http.Get(url)
	if err != nil {
		return c.String(500, "Error scraping")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.String(500, fmt.Sprintf("Error scraping: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.String(500, "Error reading body")
	}

	class, err := parseHTML(string(body))
	if err != nil {
		return c.String(500, fmt.Sprintf("Error parsing HTML: %v", err))
	}

	re := regexp.MustCompile(`ano=(\d+)&sem=(\d+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) == 3 {
		class.Year = parseInt(matches[1])
		class.Semester = parseInt(matches[2])
	}

	class.Url = url

	return c.JSON(200, class)
}

func parseHTML(html string) (*model.Class, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("error parsing HTML: %v", err)
	}

	class := &model.Class{}

	titleText := doc.Find("span#lblTitulo").Text()
	re := regexp.MustCompile(`(\w+)-(\d+)\s+(.+)\s+\((\d+)\)\s+-\s+(\d+)/(\d+)`)
	matches := re.FindStringSubmatch(titleText)

	if len(matches) == 7 {
		class.ClassCode = matches[1]
		class.Credits = parseInt(matches[2])
		class.Title = matches[3]
		class.Group = matches[4]
		class.Building = matches[5]
		class.Room = matches[6]
	}

	table := doc.Find("table#dgAulas")
	if table.Length() == 0 {
		return nil, fmt.Errorf("table with id 'dgAulas' not found")
	}

	class.Sessions = make([]map[string]interface{}, 0)

	table.Find("tr").Each(func(i int, row *goquery.Selection) {
		sessionInfo, err := parseRow(row)
		if err == nil {
			class.Sessions = append(class.Sessions, sessionInfo)
		}
	})

	return class, nil
}

func parseRow(row *goquery.Selection) (map[string]interface{}, error) {
	cells := row.Find("td").Map(func(_ int, cell *goquery.Selection) string {
		return strings.TrimSpace(cell.Text())
	})

	if len(cells) < 7 {
		return nil, fmt.Errorf("row doesn't match expected format")
	}

	if cells[2] == "Data" {
		return nil, fmt.Errorf("skipping header")
	}

	startsAt, endsAt, err := extractDateTime(cells[2], strings.TrimSpace(cells[3][2:]))
	if err != nil {
		return nil, fmt.Errorf("error extracting date and time: %v", err)
	}

	return map[string]interface{}{
		"description": getOrDefault(cells[4], "Sem descrição"),
		"startsAt":    startsAt,
		"endsAt":      endsAt,
		"activity":    getOrDefault(cells[5], "Sem atividade"),
		"observation": getOrDefault(cells[6], "Sem observação"),
	}, nil
}

func extractDateTime(dateCell, timeCell string) (startsAt, endsAt time.Time, err error) {
	date, err := time.Parse("02/01/2006", dateCell)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("error parsing date: %v", err)
	}

	timeParts := strings.Split(timeCell, " - ")
	if len(timeParts) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time format")
	}

	startTime, err := time.Parse("15:04", strings.TrimSpace(timeParts[0]))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("error parsing start time: %v", err)
	}

	endTime, err := time.Parse("15:04", strings.TrimSpace(timeParts[1]))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("error parsing end time: %v", err)
	}

	gmt3Location := time.FixedZone("GMT-3", -3*60*60)
	startsAt = time.Date(date.Year(), date.Month(), date.Day(), startTime.Hour(), startTime.Minute(), 0, 0, gmt3Location)
	endsAt = time.Date(date.Year(), date.Month(), date.Day(), endTime.Hour(), endTime.Minute(), 0, 0, gmt3Location)

	return startsAt, endsAt, nil
}

func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
