package canteen

import (
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/jhachmer/happahappa/pkg/config"
)

// CategoriesToScrape are categories with the main meals
// excluding food and drinks that are offered regularly
// please don't rename your categories stwh
var CategoriesToScrape = []string{
	"PASTA & FRIENDS",
	"FLEISCH & MEER",
	"VEGGIE & VEGAN",
	"QUEERBEET",
	"EVERGREENS",
	"SÜSSE ECKE",
}

type TodaysMenu struct {
	Categories []Category
}

// Body returns a plain text representation of the menu
func (t TodaysMenu) Body() string {
	currentDate := time.Now().Local().Format("02.01.2006")
	var sb strings.Builder
	fmt.Fprintf(&sb, "Mensa-Menu (%s)  \n", currentDate)
	for _, category := range t.Categories {
		fmt.Fprintf(&sb, "  %s\n", category.Name)
		for _, meal := range category.Meals {
			fmt.Fprintf(&sb, "    %s\n", meal)
		}
	}
	return sb.String()
}

// HTML returns an HTML representation of the menu
func (t TodaysMenu) HTML() string {
	currentDate := time.Now().Local().Format("02.01.2006")
	var sb strings.Builder
	fmt.Fprintf(&sb, "<h3>Mensa-Menu (%s)</h3>", currentDate)
	for _, category := range t.Categories {
		fmt.Fprintf(&sb, "<p></p><h4><b>%s</b></h4>", category.Name)
		fmt.Fprintf(&sb, "<ul>")
		for _, meal := range category.Meals {
			fmt.Fprintf(&sb, "<li>%s</li>", meal.HTML())
		}
		fmt.Fprintf(&sb, "</ul><hr>")
	}
	return sb.String()
}

type Category struct {
	Name  string
	Meals []Meal
}

type Meal struct {
	Name  string
	Price string
	Info  MealInfo
}

type MealInfo struct {
	Vegan      bool
	Beef       bool
	Pork       bool
	Grain      bool
	Garlic     bool
	Alcohol    bool
	Chicken    bool
	Vegetarian bool
	Fish       bool
	Milk       bool
}

func (mi MealInfo) String() string {
	emojis := []struct {
		enabled bool
		emoji   string
	}{
		{mi.Vegan, "🥦"},
		{mi.Vegetarian, "🥕"},
		{mi.Beef, "🐄"},
		{mi.Pork, "🐖"},
		{mi.Chicken, "🐓"},
		{mi.Fish, "🐟"},
		{mi.Grain, "🌾"},
		{mi.Garlic, "🧄"},
		{mi.Alcohol, "🍺"},
		{mi.Milk, "🥛"},
	}

	var sb strings.Builder

	for _, item := range emojis {
		if item.enabled {
			sb.WriteString(" ")
			sb.WriteString(item.emoji)
		}
	}

	return sb.String()
}

func (m Meal) String() string {
	return fmt.Sprintf("%s *%s* %s", m.Name, m.Price, m.Info)
}

func (m Meal) HTML() string {
	return fmt.Sprintf("%s <i>%s</i><p>%s</p>", m.Name, m.Price, m.Info)
}

type CanteenScraper struct {
	URL *url.URL
}

func NewCanteenScraper(config *config.Config) (*CanteenScraper, error) {
	requestURL, err := buildCanteenURL(config)
	if err != nil {
		return nil, err
	}
	return &CanteenScraper{
		URL: requestURL,
	}, nil
}

func buildCanteenURL(config *config.Config) (*url.URL, error) {
	u, err := url.Parse(config.Canteen.CanteenURL)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("price", strconv.Itoa(config.Canteen.PriceID))
	params.Add("pay", strconv.Itoa(config.Canteen.PayID))
	params.Add("mensa", strconv.Itoa(config.Canteen.CanteenID))

	u.RawQuery = params.Encode()

	return u, nil
}

func (c CanteenScraper) Scrape() *TodaysMenu {
	categories := make([]Category, 0)
	scraper := colly.NewCollector(
		colly.AllowedDomains("www.stwh-portal.de"),
	)
	scraper.OnHTML("h3 > span.category", func(e *colly.HTMLElement) {
		if !slices.Contains(CategoriesToScrape, e.Text) {
			return
		}
		category := Category{
			Name: e.Text,
		}
		e.DOM.Parent().Next().Find("li .food").Each(func(_ int, s *goquery.Selection) {
			name := s.Find(".food_name").Text()
			price := s.Find(".food_price").Text()

			mealInfo := MealInfo{}
			keywords := map[string]func(){
				"vegan":        func() { mealInfo.Vegan = true },
				"beef":         func() { mealInfo.Beef = true },
				"pork":         func() { mealInfo.Pork = true },
				"Gluten":       func() { mealInfo.Grain = true },
				"garlic":       func() { mealInfo.Garlic = true },
				"Geflügel":     func() { mealInfo.Chicken = true },
				"alcohol":      func() { mealInfo.Alcohol = true },
				"without meat": func() { mealInfo.Vegetarian = true },
				"Fisch":        func() { mealInfo.Fish = true },
				"Milch":        func() { mealInfo.Milk = true },
			}

			s.Find("span.symbol").Each(func(_ int, s *goquery.Selection) {
				info, exists := s.Attr("title")
				if !exists {
					return
				}

				for keyword, setter := range keywords {
					if strings.Contains(info, keyword) {
						setter()
					}
				}
			})
			category.Meals = append(category.Meals, Meal{
				Name:  name,
				Price: price,
				Info:  mealInfo,
			})
		})

		categories = append(categories, category)
	})
	scraper.OnRequest(func(r *colly.Request) {
		slog.Info("visiting URL", "url", r.URL.String())
	})
	scraper.OnError(func(r *colly.Response, err error) {
		slog.Error("error while scraping", "request_url:", r.Request.URL.String(), "response:", r, "error:", err)
	})
	scraper.OnScraped(func(r *colly.Response) {
		slog.Info("finished scraping", "url", r.Request.URL)
	})
	err := scraper.Visit(c.URL.String())
	if err != nil {
		slog.Error("error on page visit", "err", err.Error())
	}
	// Visit is async by default, waiting for it to finish
	scraper.Wait()
	return &TodaysMenu{
		Categories: categories,
	}
}
