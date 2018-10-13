package gocharge

import (
	"context"
	"errors"
	"log"
	"sort"
	"time"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

const tibberAPI = "https://api.tibber.com/v1-beta/gql"

type TibberClient struct {
	*graphql.Client
}

func NewTibberClient(token string) *TibberClient {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return &TibberClient{
		graphql.NewClient(
			tibberAPI, oauth2.NewClient(context.Background(), src),
		),
	}
}

func ShouldCharge(ctx context.Context, tibber *TibberClient, now time.Time, hours, target int) (bool, error) {
	query := struct {
		Viewer struct {
			Homes []struct {
				CurrentSubscription struct {
					PriceInfo struct {
						Today    []TibberPrice
						Tomorrow []TibberPrice
					}
				}
			}
		}
	}{}
	if err := tibber.Query(ctx, &query, nil); err != nil {
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

type TibberPrice struct {
	Total    graphql.Float
	Energy   graphql.Float
	Tax      graphql.Float
	StartsAt time.Time
}

type ByCost []TibberPrice

func (c ByCost) Len() int           { return len(c) }
func (c ByCost) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByCost) Less(i, j int) bool { return c[i].Total < c[j].Total }

func filter(prices []TibberPrice, deadlineHour int) []TibberPrice {
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

	filtered := []TibberPrice{}
	for _, p := range prices {
		if p.StartsAt.After(previousTarget) && p.StartsAt.Before(nextTarget) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
