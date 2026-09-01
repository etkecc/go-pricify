package pricify

import (
	"slices"
	"testing"
)

func TestParseServerPriceOverrides(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want map[string]int
	}{
		{"single", "plan_b-r3=17", map[string]int{"plan_b-r3": 17}},
		{"comma separated", "plan_b-r3=17,plan_a-r1=5", map[string]int{"plan_b-r3": 17, "plan_a-r1": 5}},
		{"whitespace around equals", "plan_b-r3 = 17", map[string]int{"plan_b-r3": 17}},
		{"float style", "plan_b-r3=15.00", map[string]int{"plan_b-r3": 15}},
		{"duplicate last wins", "plan_b-r3=17,plan_b-r3=20", map[string]int{"plan_b-r3": 20}},
		{"empty", "", map[string]int{}},
		{"no separator", "nosep", map[string]int{}},
		{"missing key", "=17", map[string]int{}},
		{"non numeric value", "plan_b-r3=abc", map[string]int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseServerPriceOverrides(tt.raw)
			if len(got) != len(tt.want) {
				t.Fatalf("parseServerPriceOverrides(%q) = %v, want %v", tt.raw, got, tt.want)
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Fatalf("parseServerPriceOverrides(%q) = %v, want %v", tt.raw, got, tt.want)
				}
			}
		})
	}
}

func TestCalculateVerboseServerPriceOverrideApplies(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":                       "no",
		"etke_service_server_location": "r3",
		"etke_service_server":          "plan_b",
		"etke_service_server_price":    "plan_b-r3=17",
	}

	total, verbose := data.CalculateVerbose(input)
	if total != 17 {
		t.Fatalf("expected total 17, got %d", total)
	}
	item := verbose["etke_service_server"]
	if item == nil {
		t.Fatal("expected verbose to include server item")
	}
	if item.Price != 17 {
		t.Fatalf("expected item price 17, got %d", item.Price)
	}

	for _, it := range data.Items() {
		if it.ID != "plan_b" {
			continue
		}
		if slices.Contains(it.Regions, "r3") && it.Price != 31 {
			t.Fatalf("catalog r3 price mutated to %d, want 31", it.Price)
		}
		if slices.Contains(it.Regions, "r4") && it.Price != 41 {
			t.Fatalf("catalog r4 price mutated to %d, want 41", it.Price)
		}
	}
}

func TestCalculateVerboseServerPriceOverrideRegionSpecificity(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":                       "no",
		"etke_service_server_location": "r4",
		"etke_service_server":          "plan_b",
		"etke_service_server_price":    "plan_b-r3=17",
	}

	total, _ := data.CalculateVerbose(input)
	if total != 41 {
		t.Fatalf("expected catalog total 41, got %d (override leaked across regions)", total)
	}
}

func TestCalculateVerboseServerPriceOverrideSizeSpecificity(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":                       "no",
		"etke_service_server_location": "r1",
		"etke_service_server":          "plan_a",
		"etke_service_server_price":    "plan_b-r3=17",
	}

	total, _ := data.CalculateVerbose(input)
	if total != 21 {
		t.Fatalf("expected catalog total 21, got %d (override leaked across sizes)", total)
	}
}

func TestCalculateVerboseServerPriceOverrideMultipleEntries(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":                       "no",
		"etke_service_server_location": "r3",
		"etke_service_server":          "plan_b",
		"etke_service_server_price":    "plan_b-r3=17,plan_a-r1=5",
	}

	total, _ := data.CalculateVerbose(input)
	if total != 17 {
		t.Fatalf("expected total 17, got %d (plan_a entry leaked into plan_b)", total)
	}
}

func TestCalculateVerboseServerPriceOverrideGarbageFallsBack(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":                       "no",
		"etke_service_server_location": "r3",
		"etke_service_server":          "plan_b",
		"etke_service_server_price":    "plan_b-r3=abc nosep 42",
	}

	total, _ := data.CalculateVerbose(input)
	if total != 31 {
		t.Fatalf("expected catalog total 31, got %d (garbage should fall back)", total)
	}
}

func TestCalculateVerboseServerPriceOverrideWithoutServer(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":                    "no",
		"etke_service_server_price": "plan_b-r3=17",
	}

	total, _ := data.CalculateVerbose(input)
	if total != 0 {
		t.Fatalf("expected total 0 without server selected, got %d", total)
	}
}
