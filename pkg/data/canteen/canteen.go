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
	"lab.it.hs-hannover.de/8mg-y3w-u2/happahappa/pkg/config"
	"lab.it.hs-hannover.de/8mg-y3w-u2/happahappa/pkg/data/weather"
)

// CATEGORIES_TO_SCRAPE are categories with the main meals
// excluding food and drinks that are offered regularly
// please don't rename your categories stwh
var CATEGORIES_TO_SCRAPE = []string{
	"PASTA & FRIENDS",
	"FLEISCH & MEER",
	"VEGGIE & VEGAN",
	"QUEERBEET",
	"EVERGREENS",
	"SÜSSE ECKE",
}

type Message struct {
	menu    *TodaysMenu
	weather *weather.WeatherResponse
}

type TodaysMenu struct {
	Categories []Category
}

// Body returns a plain text representation of the menu
func (t TodaysMenu) Body() string {
	currentDate := time.Now().Local().Format("02.01.2006")
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Mensa-Menu (%s)  \n", currentDate))
	for _, category := range t.Categories {
		sb.WriteString(fmt.Sprintf("  %s\n", category.Name))
		for _, meal := range category.Meals {
			sb.WriteString(fmt.Sprintf("    %s\n", meal))
		}
	}
	return sb.String()
}

// HTML returns an HTML representation of the menu
func (t TodaysMenu) HTML() string {
	currentDate := time.Now().Local().Format("02.01.2006")
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<h1>Mensa-Menu (%s)</h1>", currentDate))
	for _, category := range t.Categories {
		sb.WriteString(fmt.Sprintf("<h4><b>%s</b></h4>", category.Name))
		sb.WriteString(fmt.Sprintf("<ul>"))
		for _, meal := range category.Meals {
			sb.WriteString(fmt.Sprintf("<li>%s</li>", meal.HTML()))
		}
		sb.WriteString(fmt.Sprintf("</ul>"))
	}
	sb.WriteString(fmt.Sprintf("<br ><br >"))
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
	var sb strings.Builder
	if mi.Vegan {
		sb.WriteString(" &#x1F966 ")
	}
	if mi.Vegetarian {
		sb.WriteString(" &#x1F955 ")
	}
	if mi.Beef {
		sb.WriteString(" &#x1F404 ")
	}
	if mi.Pork {
		sb.WriteString(" &#x1F416 ")
	}
	if mi.Chicken {
		sb.WriteString(" &#x1F414 ")
	}
	if mi.Fish {
		sb.WriteString(" &#x1F41F ")
	}
	if mi.Grain {
		sb.WriteString(" &#x1F33E ")
	}
	if mi.Garlic {
		sb.WriteString(" &#x1F9C4 ")
	}
	if mi.Alcohol {
		sb.WriteString(" &#x1F37A ")
	}
	if mi.Milk {
		sb.WriteString(" &#x1F95B ")
	}
	return sb.String()
}

func (m Meal) String() string {
	return fmt.Sprintf("%s *%s* %s", m.Name, m.Price, m.Info)
}

func (m Meal) HTML() string {
	return fmt.Sprintf("%s <i>%s</i> <br >%s", m.Name, m.Price, m.Info)
}

type CanteenScraper struct {
	URL *url.URL
}

func NewCanteenScraper(config *config.Config) (*CanteenScraper, error) {
	requestUrl, err := buildCanteenURL(config)
	if err != nil {
		return nil, err
	}
	return &CanteenScraper{
		URL: requestUrl,
	}, nil
}

func buildCanteenURL(config *config.Config) (*url.URL, error) {
	u, err := url.Parse(config.Canteen.CanteenUrl)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Add("price", strconv.Itoa(config.Canteen.PriceId))
	params.Add("pay", strconv.Itoa(config.Canteen.PayId))
	params.Add("mensa", strconv.Itoa(config.Canteen.CanteenId))

	u.RawQuery = params.Encode()

	return u, nil
}

func (c CanteenScraper) Scrape() *TodaysMenu {
	categories := make([]Category, 0)
	scraper := colly.NewCollector(
		colly.AllowedDomains("www.stwh-portal.de"))
	scraper.OnHTML("h3 > span.category", func(e *colly.HTMLElement) {
		if !slices.Contains(CATEGORIES_TO_SCRAPE, e.Text) {
			return
		}
		category := Category{
			Name: e.Text,
		}
		e.DOM.Parent().Next().Find("li .food").Each(func(_ int, s *goquery.Selection) {
			name := s.Find(".food_name").Text()
			price := s.Find(".food_price").Text()

			mealInfo := MealInfo{}

			s.Find("span.symbol").Each(func(_ int, s *goquery.Selection) {
				info, _ := s.Attr("title")
				if strings.Contains(info, "vegan") {
					mealInfo.Vegan = true
				}
				if strings.Contains(info, "beef") {
					mealInfo.Beef = true
				}
				if strings.Contains(info, "pork") {
					mealInfo.Pork = true
				}
				if strings.Contains(info, "Gluten") {
					mealInfo.Grain = true
				}
				if strings.Contains(info, "garlic") {
					mealInfo.Garlic = true
				}
				if strings.Contains(info, "Geflügel") {
					mealInfo.Chicken = true
				}
				if strings.Contains(info, "alcohol") {
					mealInfo.Alcohol = true
				}
				if strings.Contains(info, "without meat") {
					mealInfo.Vegetarian = true
				}
				if strings.Contains(info, "Fisch") {
					mealInfo.Fish = true
				}
				if strings.Contains(info, "Milch") {
					mealInfo.Milk = true
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

	}
	//Visit is async by default, waiting for it to finish
	scraper.Wait()
	return &TodaysMenu{
		Categories: categories,
	}
}
