package pricify

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

const fixtureComponentsJSON = `{
  "bases": [
    {
      "id": "matrix",
			"vid": 123,
      "iid": "etke_base_matrix",
      "name": "Matrix base",
      "description": "Matrix base plan",
      "help": "/help/matrix",
      "price": 100
    }
  ],
  "instances": {
    "id": "instances",
    "iid": "etke_instance_size",
    "name": "Instance size",
    "description": "Instance sizing",
    "help": "/help/instance",
    "options": [
      {
        "id": "small",
				"vid": 124,
        "iid": "etke_instance_size",
        "name": "Small",
        "price": 50
      },
      {
        "id": "large",
				"vid": 125,
        "iid": "etke_instance_size",
        "name": "Large",
        "price": 80,
        "price_region": {
          "eu": 70
        }
      }
    ]
  },
  "support": {
    "id": "support",
    "iid": "etke_support_level",
    "name": "Support level",
    "description": "Support plans",
    "help": "/help/support",
    "options": [
      {
        "id": "basic",
				"vid": 126,
        "iid": "etke_support_level",
        "name": "Basic",
        "price": 10
      }
    ]
  },
  "matrixApps": [
    {
      "id": "app",
			"vid": 127,
      "iid": "matrix_app",
      "name": "Matrix app",
      "description": "",
      "help": "",
      "price": 5
    }
  ],
  "matrixBots": [
    {
      "id": "bot",
			"vid": 128,
      "iid": "matrix_bot",
      "name": "Matrix bot",
      "description": "",
      "help": "",
      "price": 6
    }
  ],
  "matrixBridges": [
    {
      "id": "bridge_a",
			"vid": 129,
      "iid": "matrix_bridge_a",
      "name": "Bridge A",
      "description": "",
      "help": "",
      "price": 20
    },
    {
      "id": "bridge_b",
			"vid": 130,
      "iid": "matrix_bridge_b",
      "name": "Bridge B",
      "description": "",
      "help": "",
      "price": 20
    }
  ],
  "matrixBridgesVID": 555,
  "matrixBridgesPrice": 200,
  "additionalMatrixServices": [
    {
      "id": "smtp",
			"vid": 131,
      "iid": "exim_relay_relay_use",
      "name": "SMTP relay",
      "description": "",
      "help": "",
      "price": 30
    }
  ],
  "additionalServices": [
    {
      "id": "email",
			"vid": 132,
      "iid": "etke_service_email",
      "name": "Email service",
      "description": "",
      "help": "",
      "price": 70
    }
  ],
  "archived": "ARCHIVE_URL"
}`

const fixtureArchiveJSON = `{
  "bases": [
    {
      "id": "archive_base",
      "iid": "archive_base",
      "name": "Archived base",
      "description": "",
      "help": "",
      "price": 111
    }
  ],
  "matrixBridgesVID": 777,
  "instances": {
    "id": "instances",
    "iid": "etke_instance_size",
    "name": "Instance size",
    "description": "",
    "help": "",
    "options": [
      {
        "id": "xlarge",
        "iid": "etke_instance_size",
        "name": "XLarge",
        "price": 90
      }
    ]
  }
}`

func mustParseSource(t *testing.T, b []byte) *sourceModel {
	t.Helper()
	source, err := parseSource(b)
	if err != nil {
		t.Fatalf("parseSource error: %v", err)
	}
	return source
}

func mustConvertData(t *testing.T, b []byte) *Data {
	t.Helper()
	return convertToData(mustParseSource(t, b))
}

func TestParseSourceInvalidJSON(t *testing.T) {
	_, err := parseSource([]byte("{"))
	if err == nil {
		t.Fatal("expected parseSource to return error for invalid JSON")
	}
}

func TestSourceModelAppendAndInit(t *testing.T) {
	var nilSource *sourceModel
	nilSource.append(&sourceModel{})

	base := &sourceItem{ID: "base", InventoryID: "base_iid", Name: "Base", Price: 1}
	opt := sourceItem{ID: "opt", InventoryID: "opt_iid", Name: "Option", Price: 2}

	s1 := &sourceModel{}
	s2 := &sourceModel{
		Bases: []*sourceItem{base},
		Instances: &sourceSectionItem{
			ID:          "instances",
			InventoryID: "etke_instance_size",
			Options:     []sourceItem{opt},
		},
		MatrixBridgesVID:   777,
		MatrixBridgesPrice: 200,
	}

	s1.append(s2)

	if len(s1.Bases) != 1 {
		t.Fatalf("expected 1 base, got %d", len(s1.Bases))
	}
	if s1.Instances == nil {
		t.Fatal("expected Instances to be initialized")
	}
	if len(s1.Instances.Options) != 1 {
		t.Fatalf("expected 1 instance option, got %d", len(s1.Instances.Options))
	}
	if s1.MatrixBridgesVID != 777 {
		t.Fatalf("expected MatrixBridgesVID to be copied, got %d", s1.MatrixBridgesVID)
	}
	if s1.MatrixBridgesPrice != 200 {
		t.Fatalf("expected MatrixBridgesPrice to be copied, got %d", s1.MatrixBridgesPrice)
	}

	s1.MatrixBridgesVID = 555
	s1.MatrixBridgesPrice = 150
	s1.append(&sourceModel{
		MatrixBridgesVID:   999,
		MatrixBridgesPrice: 300,
	})
	if s1.MatrixBridgesVID != 555 {
		t.Fatalf("expected MatrixBridgesVID to prefer primary, got %d", s1.MatrixBridgesVID)
	}
	if s1.MatrixBridgesPrice != 150 {
		t.Fatalf("expected MatrixBridgesPrice to prefer primary, got %d", s1.MatrixBridgesPrice)
	}
}

func TestSourceModelInitHandlesNilSections(t *testing.T) {
	source := &sourceModel{}
	source.init()
	if source.Instances == nil || source.Support == nil {
		t.Fatal("expected init to set Instances and Support")
	}
	if len(source.Instances.Options) != 0 || len(source.Support.Options) != 0 {
		t.Fatal("expected empty options for initialized sections")
	}
}

func TestConvertToDataEmptySections(t *testing.T) {
	source := &sourceModel{
		Bases: []*sourceItem{
			{
				ID:          "base",
				InventoryID: "base_iid",
				Name:        "Base",
				Price:       1,
			},
		},
		Instances: &sourceSectionItem{
			ID:          "instances",
			InventoryID: "instances_iid",
		},
		Support: &sourceSectionItem{
			ID:          "support",
			InventoryID: "support_iid",
		},
	}

	data := convertToData(source)
	if data.find("base", "yes") == nil {
		t.Fatal("expected base item to exist")
	}
	if data.find("instances", "anything") != nil {
		t.Fatal("did not expect instance options to exist")
	}
	if data.find("support", "anything") != nil {
		t.Fatal("did not expect support options to exist")
	}
}

func TestConvertToDataNilSections(t *testing.T) {
	source := &sourceModel{
		Bases: []*sourceItem{
			{
				ID:          "base",
				InventoryID: "base_iid",
				Name:        "Base",
				Price:       1,
			},
		},
		Instances: nil,
		Support:   nil,
	}

	data := convertToData(source)
	if data.find("base", "yes") == nil {
		t.Fatal("expected base item to exist")
	}
}

func TestNewHandlesMissingSections(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = origTransport
	}()

	const componentsURL = "http://example.test/missing-sections"
	payload := `{
  "bases": [
    {
      "id": "matrix",
      "iid": "etke_base_matrix",
      "name": "Matrix base",
      "description": "Matrix base plan",
      "help": "/help/matrix",
      "price": 100
    }
  ]
}`

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != componentsURL {
			return nil, http.ErrServerClosed
		}
		return testResponse(payload), nil
	})

	data, err := New(componentsURL)
	if err != nil {
		t.Fatalf("expected New to succeed, got %v", err)
	}
	if data.find("matrix", "yes") == nil {
		t.Fatal("expected base item to be present when sections are missing")
	}
}

func TestFindPrefersIDOverInventoryID(t *testing.T) {
	itemByID := &Item{
		ID:           "dup",
		InventoryID:  "iid_a",
		Value:        "yes",
		Price:        1,
		SectionID:    "section",
		SectionPrice: 0,
	}
	itemByIID := &Item{
		ID:           "other",
		InventoryID:  "dup",
		Value:        "yes",
		Price:        2,
		SectionID:    "section",
		SectionPrice: 0,
	}

	data := &Data{
		items:  []*Item{itemByID, itemByIID},
		idmap:  map[string]*Item{"dup": itemByID},
		iidmap: map[string]*Item{"dup": itemByIID},
	}

	found := data.find("dup", "yes")
	if found != itemByID {
		t.Fatalf("expected ID match to be preferred, got %+v", found)
	}
}

func TestConvertToDataPopulatesMaps(t *testing.T) {
	cached = nil
	data := mustConvertData(t, []byte(fixtureComponentsJSON))

	if data.idmap["matrix"] == nil {
		t.Fatal("expected idmap to include base item")
	}
	if data.iidmap["etke_base_matrix"] == nil {
		t.Fatal("expected iidmap to include base item")
	}

	if data.find("instances", "small") == nil {
		t.Fatal("expected find to resolve section option by id+value")
	}
	if data.find("etke_instance_size", "large") == nil {
		t.Fatal("expected find to resolve section option by inventory id+value")
	}

	if getCache() != data {
		t.Fatal("expected convertToData to update cache")
	}
}

func TestCalculateDefaultBaseMatrix(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{}

	total := data.Calculate(input)
	if total != 100 {
		t.Fatalf("expected total 100, got %d", total)
	}
	if input["etke_base_matrix"] != "yes" {
		t.Fatalf("expected default etke_base_matrix to be set, got %q", input["etke_base_matrix"])
	}
}

func TestCalculateVerboseRegionPrice(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":                       "no",
		"etke_service_server_location": "eu",
		"etke_instance_size":           "large",
	}

	total, verbose := data.CalculateVerbose(input)
	if total != 70 {
		t.Fatalf("expected total 70, got %d", total)
	}
	item := verbose["etke_instance_size"]
	if item == nil {
		t.Fatal("expected verbose to include instance item")
	}
	if item.VID != 125 {
		t.Fatalf("expected instance VID 125, got %d", item.VID)
	}
	if item.Price != 70 {
		t.Fatalf("expected region price 70, got %d", item.Price)
	}
}

func TestCalculateVerboseRegionPriceFallbackToBase(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":                       "no",
		"etke_service_server_location": "us",
		"etke_instance_size":           "large",
	}

	total, verbose := data.CalculateVerbose(input)
	if total != 80 {
		t.Fatalf("expected total 80, got %d", total)
	}
	item := verbose["etke_instance_size"]
	if item == nil {
		t.Fatal("expected verbose to include instance item")
	}
	if item.Price != 80 {
		t.Fatalf("expected base price 80, got %d", item.Price)
	}
}

func TestCalculateVerboseSMTPRelayFreeWithEmail(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":               "no",
		"etke_service_email":   "yes",
		"exim_relay_relay_use": "yes",
	}

	total, verbose := data.CalculateVerbose(input)
	if total != 70 {
		t.Fatalf("expected total 70 (email only), got %d", total)
	}
	relay := verbose["exim_relay_relay_use"]
	if relay == nil {
		t.Fatal("expected verbose to include relay item")
	}
	if relay.Price != 0 {
		t.Fatalf("expected relay to be free, got %d", relay.Price)
	}
	if !strings.Contains(relay.Name, "free with email service") {
		t.Fatalf("expected relay name to note free status, got %q", relay.Name)
	}
	if relay.Value != "with email service" {
		t.Fatalf("expected relay value to be overridden, got %q", relay.Value)
	}
}

func TestCalculateVerboseSectionPriceOnce(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":          "no",
		"matrix_bridge_a": "yes",
		"matrix_bridge_b": "yes",
	}

	total, verbose := data.CalculateVerbose(input)
	if total != 220 {
		t.Fatalf("expected total 220 (section + one bridge), got %d", total)
	}
	if verbose["matrix_bridges"] == nil {
		t.Fatal("expected verbose to include section entry")
	}
	if verbose["matrix_bridges"].VID != 555 {
		t.Fatalf("expected section VID 555, got %d", verbose["matrix_bridges"].VID)
	}
	bridgeCount := 0
	if verbose["matrix_bridge_a"] != nil {
		bridgeCount++
	}
	if verbose["matrix_bridge_b"] != nil {
		bridgeCount++
	}
	if bridgeCount != 1 {
		t.Fatalf("expected exactly 1 bridge entry in verbose, got %d", bridgeCount)
	}
}

func TestCalculateVerboseMixedSections(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix":             "no",
		"matrix_bridge_a":    "yes",
		"matrix_app":         "yes",
		"matrix_bot":         "yes",
		"etke_instance_size": "small",
	}

	total, verbose := data.CalculateVerbose(input)
	if total != 261 {
		t.Fatalf("expected total 261, got %d", total)
	}
	if verbose["matrix_app"] == nil || verbose["matrix_app"].VID != 127 {
		t.Fatalf("expected matrix app VID 127, got %+v", verbose["matrix_app"])
	}
	if verbose["matrix_bot"] == nil || verbose["matrix_bot"].VID != 128 {
		t.Fatalf("expected matrix bot VID 128, got %+v", verbose["matrix_bot"])
	}
	if verbose["etke_instance_size"] == nil || verbose["etke_instance_size"].VID != 124 {
		t.Fatalf("expected instance VID 124, got %+v", verbose["etke_instance_size"])
	}
}

func TestCalculateForbiddenValuesSkipDefault(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"matrix": "NO",
	}

	total := data.Calculate(input)
	if total != 0 {
		t.Fatalf("expected total 0 when matrix is disabled, got %d", total)
	}
	if _, ok := input["etke_base_matrix"]; ok {
		t.Fatal("did not expect default base matrix when matrix key exists")
	}
}

func TestCalculateForbiddenValuesSkipDefaultOnInventoryID(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"etke_base_matrix": " false ",
	}

	total := data.Calculate(input)
	if total != 0 {
		t.Fatalf("expected total 0 when base matrix is disabled, got %d", total)
	}
	if input["etke_base_matrix"] != " false " {
		t.Fatalf("expected input value to remain unchanged, got %q", input["etke_base_matrix"])
	}
}

func TestCalculateVerboseTrimsAndLowercases(t *testing.T) {
	data := mustConvertData(t, []byte(fixtureComponentsJSON))
	input := map[string]string{
		"  MATRIX  ": "  YeS  ",
	}

	total, verbose := data.CalculateVerbose(input)
	if total != 100 {
		t.Fatalf("expected total 100, got %d", total)
	}
	if verbose["etke_base_matrix"] == nil {
		t.Fatal("expected verbose to include base matrix item")
	}
}

func TestNewUsesArchiveAndCacheOnError(t *testing.T) {
	cached = nil
	origTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = origTransport
	}()

	archiveURL := "http://example.test/archive"
	componentsURL := "http://example.test/components"

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case componentsURL:
			body := strings.ReplaceAll(fixtureComponentsJSON, "ARCHIVE_URL", archiveURL)
			body = strings.ReplaceAll(body, "\"matrixBridgesVID\": 555,\n", "")
			return testResponse(body), nil
		case archiveURL:
			return testResponse(fixtureArchiveJSON), nil
		default:
			return nil, http.ErrServerClosed
		}
	})

	data, err := New(componentsURL)
	if err != nil {
		t.Fatalf("expected New to succeed, got %v", err)
	}
	if data.find("instances", "xlarge") == nil {
		t.Fatal("expected archive instance option to be available")
	}
	_, verbose := data.CalculateVerbose(map[string]string{
		"matrix":          "no",
		"matrix_bridge_a": "yes",
	})
	if verbose["matrix_bridges"] == nil {
		t.Fatal("expected verbose to include matrix bridges section")
	}
	if verbose["matrix_bridges"].VID != 777 {
		t.Fatalf("expected archived section VID 777, got %d", verbose["matrix_bridges"].VID)
	}

	setCache(data)
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() == "http://example.test/bad" {
			return testResponse("{"), nil
		}
		return nil, http.ErrServerClosed
	})

	cachedData, err := New("http://example.test/bad")
	if err == nil {
		t.Fatal("expected New to return error for invalid JSON")
	}
	if cachedData != data {
		t.Fatal("expected New to return cached data on error")
	}
}

func TestLoadRejectsHTTPStatus(t *testing.T) {
	origTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = origTransport
	}()

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "http://example.test/status" {
			return nil, http.ErrServerClosed
		}
		resp := testResponse(fixtureArchiveJSON)
		resp.StatusCode = http.StatusInternalServerError
		return resp, nil
	})

	source, err := load("http://example.test/status")
	if err == nil {
		t.Fatal("expected load to fail for non-200 status")
	}
	if source != nil {
		t.Fatal("expected no source to be returned on non-200 status")
	}
}

func TestCloneReturnsCopy(t *testing.T) {
	item := &Item{
		ID:          "id",
		InventoryID: "iid",
		Name:        "Name",
		Price:       10,
		RegionPrice: map[string]int{"eu": 9},
	}
	clone := item.Clone()
	if clone == item {
		t.Fatal("expected Clone to return a different pointer")
	}
	if clone.ID != item.ID || clone.InventoryID != item.InventoryID || clone.Name != item.Name || clone.Price != item.Price {
		t.Fatal("expected Clone to copy item values")
	}
	if len(clone.RegionPrice) != len(item.RegionPrice) {
		t.Fatal("expected Clone to copy region prices")
	}
	if clone.RegionPrice["eu"] != 9 {
		t.Fatalf("expected region price to be preserved, got %d", clone.RegionPrice["eu"])
	}
}

func TestCloneDeepCopiesRegionPriceMap(t *testing.T) {
	item := &Item{
		ID:          "id",
		InventoryID: "iid",
		Name:        "Name",
		Price:       10,
		RegionPrice: map[string]int{"eu": 9},
	}
	clone := item.Clone()
	clone.RegionPrice["eu"] = 11
	if item.RegionPrice["eu"] != 9 {
		t.Fatal("expected Clone to deep copy RegionPrice map")
	}
}

func TestParseSourceRoundTrip(t *testing.T) {
	source := mustParseSource(t, []byte(fixtureComponentsJSON))
	encoded, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	if !bytes.Contains(encoded, []byte(`"bases"`)) {
		t.Fatalf("expected marshaled source to include bases, got %s", encoded)
	}
}

func TestFromSourceItemAndSectionMapping(t *testing.T) {
	source := &sourceModel{
		Bases: []*sourceItem{
			{
				ID:          "base",
				InventoryID: "base_iid",
				Name:        "Base",
				Description: "Base description",
				Help:        "/help/base",
				Price:       1,
			},
		},
		Instances: &sourceSectionItem{
			ID:          "instances",
			InventoryID: "instances_iid",
			Name:        "Instances",
			Description: "Instance description",
			Help:        "/help/instances",
			Options: []sourceItem{
				{
					ID:          "small",
					InventoryID: "instances_iid",
					Name:        "Small",
					Price:       5,
				},
			},
		},
		Support: &sourceSectionItem{},
	}

	data := convertToData(source)
	base := data.find("base", "yes")
	if base == nil {
		t.Fatal("expected base item to exist")
	}
	if base.Description != "Base description" || base.Help != "/help/base" || base.Value != "yes" {
		t.Fatal("expected base item fields to be mapped")
	}

	instance := data.find("instances", "small")
	if instance == nil {
		t.Fatal("expected instance option to exist")
	}
	if !strings.Contains(instance.Name, "Instances") || !strings.Contains(instance.Name, "Small") {
		t.Fatalf("expected instance name to include section and option, got %q", instance.Name)
	}
	if instance.Description != "Instance description" || instance.Help != "/help/instances" {
		t.Fatal("expected instance description/help to be mapped from section")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
