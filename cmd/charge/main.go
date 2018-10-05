package main

import (
	"context"
	"errors"
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
	hours := flag.Int("hours", 4, "number of hours charging per day")
	beforeHour := flag.Int("beforeHour", 8, "guaranteed charging before this hour")
	flag.Parse()

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *tibberToken},
	)

	client := graphql.NewClient(
		*tibberURL, oauth2.NewClient(context.Background(), src),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	ok, err := shouldCharge(ctx, client, time.Now(), *hours, *beforeHour)
	if err != nil {
		log.Fatal(err)
	}
	if ok {
		log.Println("Charge!")
	} else {
		log.Println("wait for it")
	}
}

func shouldCharge(ctx context.Context, client *graphql.Client, now time.Time, hours, target int) (bool, error) {
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
	if err := client.Query(ctx, &query, nil); err != nil {
		return false, err
	}

	if len(query.Viewer.Homes) != 1 {
		return false, errors.New("unsupported number of homes")
	}
	priceInfo := query.Viewer.Homes[0].CurrentSubscription.PriceInfo
	allPrices := append(
		filter(priceInfo.Today, target),
		filter(priceInfo.Tomorrow, target)...,
	)
	sort.Sort(ByCost(allPrices))
	if n := len(allPrices); n < hours {
		hours = n
	}
	log.Println("total", len(allPrices), allPrices[:hours])

	for _, goodTimes := range allPrices[:hours] {
		endsAt := goodTimes.StartsAt.Add(time.Hour)
		if goodTimes.StartsAt.Equal(now.Truncate(time.Hour)) && endsAt.After(now) {
			return true, nil
		}
	}
	return false, nil
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
	nextTarget := time.Date(
		targetDate.Year(), targetDate.Month(), targetDate.Day(),
		deadlineHour,
		0, 0, 0, targetDate.Location(),
	)
	previousTarget := nextTarget.Add(-24 * time.Hour)

	filtered := []Price{}
	for _, p := range prices {
		if p.StartsAt.After(previousTarget) && p.StartsAt.Before(nextTarget) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
