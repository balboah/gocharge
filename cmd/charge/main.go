package main

import (
	"context"
	"flag"
	"log"
	"sort"
	"time"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

func main() {
	tibberURL := flag.String("tibberURL", "https://api.tibber.com/v1-beta/gql", "URL for the Tibber API")
	tibberToken := flag.String("tibberToken", "", "access token for the Tibber API")
	minHours := flag.Int("minHours", 4, "guaranteed number of hours charging per day")
	beforeHour := flag.Int("beforeHour", 8, "guaranteed charging before this hour")
	flag.Parse()

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *tibberToken},
	)

	client := graphql.NewClient(
		*tibberURL, oauth2.NewClient(context.Background(), src),
	)
	query := struct {
		Viewer struct {
			Homes []struct {
				CurrentSubscription struct {
					PriceInfo struct {
						Today    []Price
						Tomorrow []Price
					}
				}
			}
		}
	}{}
	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		log.Fatal(err)
	}

	priceInfo := query.Viewer.Homes[0].CurrentSubscription.PriceInfo
	allPrices := append(
		filter(priceInfo.Today, *beforeHour),
		filter(priceInfo.Tomorrow, *beforeHour)...,
	)
	sort.Sort(ByCost(allPrices))
	log.Println(allPrices[:*minHours])
}

type Price struct {
	Total    graphql.Float
	Energy   graphql.Float
	Tax      graphql.Float
	StartsAt time.Time
}

type ByCost []Price

func (c ByCost) Len() int           { return len(c) }
func (c ByCost) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByCost) Less(i, j int) bool { return c[i].Total < c[j].Total }

func filter(prices []Price, deadlineHour int) []Price {
	today := time.Now()
	tomorrow := today.Add(24 * time.Hour)
	targetDate := today
	// If we already passed the deadline, continue until tomorrow.
	if today.Hour() > deadlineHour {
		targetDate = tomorrow
	}
	targetTime := time.Date(
		targetDate.Year(), targetDate.Month(), targetDate.Day(),
		deadlineHour,
		0, 0, 0, targetDate.Location(),
	)

	filtered := []Price{}
	for _, p := range prices {
		if p.StartsAt.After(today) && p.StartsAt.Before(targetTime) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
