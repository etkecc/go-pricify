package pricify

import (
	"testing"
)

func TestCalculate(t *testing.T) {
	// Create a sample Data instance for testing
	data := &Data{
		items: []*Item{
			{
				ID:           "item1",
				InventoryID:  "inv1",
				Value:        "yes",
				Price:        100,
				Section:      "section1",
				SectionPrice: 200,
			},
			{
				ID:           "item2",
				InventoryID:  "inv2",
				Value:        "no",
				Price:        150,
				Section:      "section2",
				SectionPrice: 0,
			},
		},
		idmap:  map[string]*Item{},
		iidmap: map[string]*Item{},
	}

	// Add items to the idmap and iidmap
	for _, item := range data.items {
		data.idmap[item.ID] = item
		data.iidmap[item.InventoryID] = item
	}

	// Test Calculate function
	input := map[string]string{
		"item1": "yes",
		"item2": "no",
		"item3": "maybe",
	}

	total := data.Calculate(input)
	expectedTotal := 200 // 200 (item1 section) + 0 (item2, disabled) + 0 (item3, not found)
	if total != expectedTotal {
		t.Errorf("Expected total: %d, but got: %d", expectedTotal, total)
	}
}

func TestCalculateVerbose(t *testing.T) {
	// Create a sample Data instance for testing
	data := &Data{
		items: []*Item{
			{
				ID:           "item1",
				InventoryID:  "inv1",
				Value:        "yes",
				Price:        100,
				Section:      "section1",
				SectionPrice: 200,
			},
			{
				ID:           "item2",
				InventoryID:  "inv2",
				Value:        "no",
				Price:        150,
				Section:      "section2",
				SectionPrice: 0,
			},
		},
		idmap:  map[string]*Item{},
		iidmap: map[string]*Item{},
	}

	// Add items to the idmap and iidmap
	for _, item := range data.items {
		data.idmap[item.ID] = item
		data.iidmap[item.InventoryID] = item
	}

	// Test CalculateVerbose function
	input := map[string]string{
		"item1": "yes",
		"item2": "no",
		"item3": "maybe",
	}

	total, verbose := data.CalculateVerbose(input)
	expectedTotal := 200 // 200 (item1 section) + 0 (item2, disabled) + 0 (item3, not found)
	if total != expectedTotal {
		t.Errorf("Expected total: %d, but got: %d", expectedTotal, total)
	}

	// Check verbose details
	expectedVerbose := map[string]*Item{
		"section1": {
			ID:          "section-section1",
			InventoryID: "section_section1",
			Price:       200,
		},
	}
	for key, expectedItem := range expectedVerbose {
		if item, ok := verbose[key]; !ok {
			t.Errorf("Expected verbose item with key %s, but not found", key)
		} else if item.ID != expectedItem.ID || item.InventoryID != expectedItem.InventoryID {
			t.Errorf("Expected verbose item %+v, but got %+v", expectedItem, item)
		}
	}
}

func TestFind(t *testing.T) {
	// Create a sample Data instance for testing
	data := &Data{
		items: []*Item{
			{
				ID:           "item1",
				InventoryID:  "inv1",
				Value:        "yes",
				Price:        100,
				Section:      "section1",
				SectionPrice: 200,
			},
			{
				ID:           "item2",
				InventoryID:  "inv2",
				Value:        "no",
				Price:        150,
				Section:      "section2",
				SectionPrice: 0,
			},
		},
		idmap:  map[string]*Item{},
		iidmap: map[string]*Item{},
	}

	// Add items to the idmap and iidmap
	for _, item := range data.items {
		data.idmap[item.ID] = item
		data.iidmap[item.InventoryID] = item
	}

	// Test find function
	foundItem := data.find("item1", "yes")
	if foundItem == nil { //nolint:staticcheck // nil check is above
		t.Error("Expected to find item1, but it was not found")
	}
	if foundItem.ID != "item1" { //nolint:staticcheck // nil check is above
		t.Errorf("Expected item1, but found %s", foundItem.ID)
	}

	notFoundItem := data.find("item3", "maybe")
	if notFoundItem != nil {
		t.Errorf("Expected not to find item3, but found %s", notFoundItem.ID)
	}
}

func TestCalculateWithSectionPrice(t *testing.T) {
	// Create a sample Data instance for testing with items having section prices
	data := &Data{
		items: []*Item{
			{
				ID:           "item1",
				InventoryID:  "inv1",
				Value:        "yes",
				Price:        100,
				Section:      "section1",
				SectionPrice: 200,
			},
			{
				ID:           "item2",
				InventoryID:  "inv2",
				Value:        "no",
				Price:        150,
				Section:      "section2",
				SectionPrice: 0,
			},
		},
		idmap:  map[string]*Item{},
		iidmap: map[string]*Item{},
	}

	// Add items to the idmap and iidmap
	for _, item := range data.items {
		data.idmap[item.ID] = item
		data.iidmap[item.InventoryID] = item
	}

	// Test Calculate function with items having section prices
	input := map[string]string{
		"item1": "yes",
		"item2": "no",
		"item3": "maybe",
	}

	total := data.Calculate(input)
	// Since "item1" and "item2" have section prices, their own prices should be ignored.
	// Total should be 200 (section1) + 0 (section2) + 0 (item3, not found)
	expectedTotal := 200
	if total != expectedTotal {
		t.Errorf("Expected total: %d, but got: %d", expectedTotal, total)
	}
}

func TestCalculateVerboseWithSectionPrice(t *testing.T) {
	// Create a sample Data instance for testing with items having section prices
	data := &Data{
		items: []*Item{
			{
				ID:           "item1",
				InventoryID:  "inv1",
				Value:        "yes",
				Price:        100,
				Section:      "section1",
				SectionPrice: 200,
			},
			{
				ID:           "item2",
				InventoryID:  "inv2",
				Value:        "no",
				Price:        150,
				Section:      "section2",
				SectionPrice: 0,
			},
		},
		idmap:  map[string]*Item{},
		iidmap: map[string]*Item{},
	}

	// Add items to the idmap and iidmap
	for _, item := range data.items {
		data.idmap[item.ID] = item
		data.iidmap[item.InventoryID] = item
	}

	// Test CalculateVerbose function with items having section prices
	input := map[string]string{
		"item1": "yes",
		"item2": "no",
		"item3": "maybe",
	}

	total, verbose := data.CalculateVerbose(input)
	// Since "item1" and "item2" have section prices, their own prices should be ignored.
	// Total should be 200 (section1) + 0 (section2) + 0 (item3, not found)
	expectedTotal := 200
	if total != expectedTotal {
		t.Errorf("Expected total: %d, but got: %d", expectedTotal, total)
	}

	// Check verbose details
	expectedVerbose := map[string]*Item{
		"section1": {
			ID:          "section-section1",
			InventoryID: "section_section1",
			Price:       200,
		},
	}
	for key, expectedItem := range expectedVerbose {
		if item, ok := verbose[key]; !ok {
			t.Errorf("Expected verbose item with key %s, but not found", key)
		} else if item.ID != expectedItem.ID || item.InventoryID != expectedItem.InventoryID || item.Price != expectedItem.Price {
			t.Errorf("Expected verbose item %v, but got %v", expectedItem, item)
		}
	}
}
