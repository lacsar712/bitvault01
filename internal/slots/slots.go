package slots

import (
	"fmt"
	"strconv"
	"strings"
)

type Rec struct {
	Title, Body string
	Tags        []string
}

func Sample() Rec {
	return Rec{Title: "gw-rev3-20260818", Body: "slot=a counter=12", Tags: []string{"gw-rev3"}}
}

func Seed() []Rec {
	return []Rec{
		Sample(),
		{Title: "gw-rev3-20260818-b", Body: "slot=b counter=12", Tags: []string{"gw-rev3"}},
	}
}

func AfterWrite(getMin func() (string, error), setMin func(string) error, body string) error {
	c, err := ParseCounter(body)
	if err != nil {
		return err
	}
	cur, err := getMin()
	if err != nil && cur != "" {
		return err
	}
	// An empty stored minimum means no rollback floor has been committed yet,
	// so the new counter is accepted as the floor. Once a floor exists, the new
	// counter must be >= it; anything lower is a rollback and must be rejected.
	if cur != "" {
		min, conv := strconv.Atoi(cur)
		if conv != nil {
			return conv
		}
		if c < min {
			return fmt.Errorf("counter rollback rejected: %d < %d", c, min)
		}
	}
	return setMin(strconv.Itoa(c))
}

func Steps() []string { return []string{"counter-check", "index-images", "export-campaign"} }

func Enforce(title, body string, tags []string) error {
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("image title required")
	}
	slot, counter, err := parse(body)
	if err != nil {
		return err
	}
	if slot != "a" && slot != "b" {
		return fmt.Errorf("slot must be a or b")
	}
	if counter < 0 {
		return fmt.Errorf("counter must be >= 0")
	}
	if len(tags) == 0 {
		return fmt.Errorf("hardware series tag required")
	}
	return nil
}

func ParseCounter(body string) (int, error) {
	_, c, err := parse(body)
	return c, err
}

func parse(body string) (slot string, counter int, err error) {
	gotS, gotC := false, false
	for _, part := range strings.Fields(body) {
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		switch k {
		case "slot":
			slot, gotS = v, true
		case "counter":
			n, conv := strconv.Atoi(v)
			if conv != nil {
				return "", 0, conv
			}
			counter, gotC = n, true
		}
	}
	if !gotS || !gotC {
		return "", 0, fmt.Errorf("body must contain slot= and counter=")
	}
	return slot, counter, nil
}
