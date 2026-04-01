package electricityprice

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lepinkainen/hovimestari/internal/store"
)

// SourcePrefix is the prefix used for electricity price memory sources
const SourcePrefix = "electricity"

// XML structures ported from github.com/lepinkainen/entsoe

type publicationMarketDocument struct {
	XMLName    xml.Name     `xml:"Publication_MarketDocument"`
	TimeSeries []timeSeries `xml:"TimeSeries"`
}

type acknowledgementMarketDocument struct {
	XMLName xml.Name `xml:"Acknowledgement_MarketDocument"`
	Reason  struct {
		Code string `xml:"code"`
		Text string `xml:"text"`
	} `xml:"Reason"`
}

type timeSeries struct {
	Period struct {
		TimeInterval struct {
			Start string `xml:"start"`
		} `xml:"timeInterval"`
		Resolution string   `xml:"resolution"`
		Points     []point  `xml:"Point"`
	} `xml:"Period"`
}

type point struct {
	Position    int     `xml:"position"`
	PriceAmount float64 `xml:"price.amount"`
}

type pricePoint struct {
	utcTime time.Time
	price   float64 // c/kWh
}

// Importer handles importing electricity prices from ENTSO-E
type Importer struct {
	store    *store.Store
	apiKey   string
	zone     string
	timezone *time.Location
}

// NewImporter creates a new electricity price importer
func NewImporter(s *store.Store, apiKey, zone string, tz *time.Location) *Importer {
	if tz == nil {
		tz = time.UTC
	}
	return &Importer{
		store:    s,
		apiKey:   apiKey,
		zone:     zone,
		timezone: tz,
	}
}

// Import fetches today's electricity prices and stores them as a memory
func (i *Importer) Import(ctx context.Context) error {
	if i.apiKey == "" {
		return fmt.Errorf("ENTSO-E API key is not configured")
	}

	now := time.Now().In(i.timezone)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, i.timezone)
	tomorrow := today.AddDate(0, 0, 1)

	// ENTSO-E API expects timestamps in YYYYMMDDhhmm UTC format
	startAPI := today.UTC().Format("200601021504")
	endAPI := tomorrow.UTC().Format("200601021504")

	prices, err := i.fetchPrices(ctx, startAPI, endAPI)
	if err != nil {
		return fmt.Errorf("failed to fetch electricity prices: %w", err)
	}

	if len(prices) == 0 {
		return fmt.Errorf("no price data returned for %s", today.Format("2006-01-02"))
	}

	content := formatPriceMemory(prices, today, i.timezone)
	source := fmt.Sprintf("%s:%s", SourcePrefix, i.zone)

	if err = i.store.DeleteMemoriesBySourceAndDate(source, today); err != nil {
		return fmt.Errorf("failed to remove existing electricity price memory: %w", err)
	}

	_, err = i.store.AddMemory(content, &today, source, nil)
	if err != nil {
		return fmt.Errorf("failed to store electricity price memory: %w", err)
	}

	slog.Info("Electricity prices imported", "date", today.Format("2006-01-02"), "count", len(prices))
	return nil
}

func (i *Importer) fetchPrices(ctx context.Context, startAPI, endAPI string) ([]pricePoint, error) {
	url := fmt.Sprintf(
		"https://web-api.tp.entsoe.eu/api?securityToken=%s&documentType=A44&out_Domain=%s&in_Domain=%s&periodStart=%s&periodEnd=%s",
		i.apiKey, i.zone, i.zone, startAPI, endAPI,
	)

	client := &http.Client{Timeout: 30 * time.Second}

	var xmlData []byte
	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
			}
			continue
		}

		xmlData, err = io.ReadAll(resp.Body)
		if closeErr := resp.Body.Close(); closeErr != nil {
			slog.Debug("Failed to close response body", "error", closeErr)
		}
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
			}
			continue
		}
		lastErr = nil
		break
	}

	if lastErr != nil {
		return nil, fmt.Errorf("HTTP request failed after %d attempts: %w", maxRetries, lastErr)
	}

	var doc publicationMarketDocument
	if err := xml.Unmarshal(xmlData, &doc); err != nil {
		var errDoc acknowledgementMarketDocument
		if xmlErr := xml.Unmarshal(xmlData, &errDoc); xmlErr == nil {
			return nil, fmt.Errorf("ENTSO-E API error %s: %s", errDoc.Reason.Code, errDoc.Reason.Text)
		}
		return nil, fmt.Errorf("failed to parse ENTSO-E response: %w", err)
	}

	const layout = "2006-01-02T15:04Z"
	var prices []pricePoint

	for _, ts := range doc.TimeSeries {
		start, err := time.Parse(layout, ts.Period.TimeInterval.Start)
		if err != nil {
			return nil, fmt.Errorf("failed to parse period start %q: %w", ts.Period.TimeInterval.Start, err)
		}

		resolution, err := isoResolutionToDuration(ts.Period.Resolution)
		if err != nil {
			slog.Warn("Unknown resolution, defaulting to 1h", "resolution", ts.Period.Resolution)
			resolution = time.Hour
		}

		for _, p := range ts.Period.Points {
			pointTime := start.Add(time.Duration(p.Position-1) * resolution)
			prices = append(prices, pricePoint{
				utcTime: pointTime,
				price:   p.PriceAmount / 10, // EUR/MWh → c/kWh
			})
		}
	}

	sort.Slice(prices, func(a, b int) bool {
		return prices[a].utcTime.Before(prices[b].utcTime)
	})

	return prices, nil
}

func formatPriceMemory(prices []pricePoint, date time.Time, tz *time.Location) string {
	var total float64
	minPrice, maxPrice := prices[0].price, prices[0].price
	minHour, maxHour := prices[0].utcTime.In(tz).Hour(), prices[0].utcTime.In(tz).Hour()

	for _, p := range prices {
		total += p.price
		h := p.utcTime.In(tz).Hour()
		if p.price < minPrice {
			minPrice = p.price
			minHour = h
		}
		if p.price > maxPrice {
			maxPrice = p.price
			maxHour = h
		}
	}

	avg := total / float64(len(prices))

	var classification string
	switch {
	case avg < 5:
		classification = "CHEAP DAY"
	case avg > 20:
		classification = "EXPENSIVE DAY"
	default:
		classification = "NORMAL DAY"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Electricity prices on %s (%s - avg %.1f c/kWh). Times in %s.\n",
		date.Format("2006-01-02"), classification, avg, tz)
	fmt.Fprintf(&sb, "Min: %.1f c/kWh at %02d:00, Max: %.1f c/kWh at %02d:00\n",
		minPrice, minHour, maxPrice, maxHour)

	// Notable spikes: price > 1.5× average, top 5 by price
	type spike struct {
		hour  int
		price float64
	}
	var spikes []spike
	for _, p := range prices {
		if p.price > avg*1.5 {
			spikes = append(spikes, spike{p.utcTime.In(tz).Hour(), p.price})
		}
	}
	sort.Slice(spikes, func(a, b int) bool { return spikes[a].price > spikes[b].price })
	if len(spikes) > 5 {
		spikes = spikes[:5]
	}
	if len(spikes) > 0 {
		sb.WriteString("Notable spikes (>1.5× avg):")
		for _, s := range spikes {
			fmt.Fprintf(&sb, " %02d:00=%.1f", s.hour, s.price)
		}
		sb.WriteString("\n")
	}

	// All hourly prices
	sb.WriteString("Hourly prices (c/kWh):")
	for _, p := range prices {
		fmt.Fprintf(&sb, " %02d=%.1f", p.utcTime.In(tz).Hour(), p.price)
	}

	return strings.TrimRight(sb.String(), "\n")
}

// isoResolutionToDuration converts ISO 8601 duration strings (PT1H, PT15M, PT30S) to time.Duration.
// Ported from github.com/lepinkainen/entsoe.
func isoResolutionToDuration(resolution string) (time.Duration, error) {
	res := strings.TrimSpace(resolution)
	if res == "" {
		return 0, fmt.Errorf("empty resolution")
	}
	if !strings.HasPrefix(res, "PT") {
		return 0, fmt.Errorf("unsupported ISO-8601 duration: %s", res)
	}
	value := strings.TrimPrefix(res, "PT")
	switch {
	case strings.HasSuffix(value, "H"):
		hours, err := strconv.Atoi(strings.TrimSuffix(value, "H"))
		if err != nil {
			return 0, fmt.Errorf("invalid hour resolution %s: %w", res, err)
		}
		return time.Duration(hours) * time.Hour, nil
	case strings.HasSuffix(value, "M"):
		minutes, err := strconv.Atoi(strings.TrimSuffix(value, "M"))
		if err != nil {
			return 0, fmt.Errorf("invalid minute resolution %s: %w", res, err)
		}
		return time.Duration(minutes) * time.Minute, nil
	case strings.HasSuffix(value, "S"):
		seconds, err := strconv.Atoi(strings.TrimSuffix(value, "S"))
		if err != nil {
			return 0, fmt.Errorf("invalid second resolution %s: %w", res, err)
		}
		return time.Duration(seconds) * time.Second, nil
	}
	return 0, fmt.Errorf("unsupported resolution unit in %s", res)
}
