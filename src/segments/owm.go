package segments

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/url"

	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Owm struct {
	base

	Weather     string
	URL         string
	units       string
	UnitIcon    string
	Temperature int
}

const (
	// APIKey openweathermap api key
	APIKey properties.Property = "api_key"
	// Location openweathermap location
	Location properties.Property = "location"
	// Units openweathermap units
	Units properties.Property = "units"
	// CacheKeyResponse key used when caching the response
	CacheKeyResponse string = "owm_response"
	// CacheKeyURL key used when caching the url responsible for the response
	CacheKeyURL string = "owm_url"
	// Environmental variable to dynamically set the Open Map API key
	OWMAPIKey string = "POSH_OWM_API_KEY"
	// Environmental variable to dynamically set the location string
	OWMLocationKey string = "POSH_OWM_LOCATION"
)

type weather struct {
	ShortDescription string `json:"main"`
	Description      string `json:"description"`
	TypeID           string `json:"icon"`
}
type temperature struct {
	Value float64 `json:"temp"`
}

type owmDataResponse struct {
	Data        []weather `json:"weather"`
	temperature `json:"main"`
}

func (d *Owm) Enabled() bool {
	err := d.setStatus()

	if err != nil {
		log.Error(err)
		return false
	}

	return true
}

func (d *Owm) Template() string {
	return " {{ .Weather }} ({{ .Temperature }}{{ .UnitIcon }}) "
}

func (d *Owm) getResult() (*owmDataResponse, error) {
	response := new(owmDataResponse)

	apikey := properties.OneOf(d.props, d.env.Getenv(OWMAPIKey), APIKey, "apiKey")
	if apikey == "" {
		apikey = "."
	}

	if apikey == "" {
		return nil, errors.New("no api key found")
	}

	location := d.props.GetString(Location, d.env.Getenv(OWMLocationKey))
	if location == "" {
		return nil, errors.New("no location found")
	}

	location = url.QueryEscape(location)

	units := d.props.GetString(Units, "standard")
	httpTimeout := d.props.GetInt(properties.HTTPTimeout, properties.DefaultHTTPTimeout)

	d.URL = fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&units=%s&appid=%s", location, units, apikey)

	body, err := d.env.HTTPRequest(d.URL, nil, httpTimeout)
	if err != nil {
		return new(owmDataResponse), err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return new(owmDataResponse), err
	}

	return response, nil
}

func (d *Owm) setStatus() error {
	units := d.props.GetString(Units, "standard")

	q, err := d.getResult()
	if err != nil {
		return err
	}

	if len(q.Data) == 0 {
		return errors.New("no data found")
	}

	id := q.Data[0].TypeID

	d.Temperature = int(math.Round(q.Value))
	icon := ""
	switch id {
	case "01n":
		icon = "\ue32b"
	case "01d":
		icon = "\ue30d"
	case "02n":
		icon = "\ue37e"
	case "02d":
		icon = "\ue302"
	case "03n":
		fallthrough
	case "03d":
		icon = "\ue33d"
	case "04n":
		fallthrough
	case "04d":
		icon = "\ue312"
	case "09n":
		fallthrough
	case "09d":
		icon = "\ue319"
	case "10n":
		icon = "\ue325"
	case "10d":
		icon = "\ue308"
	case "11n":
		icon = "\ue32a"
	case "11d":
		icon = "\ue30f"
	case "13n":
		fallthrough
	case "13d":
		icon = "\ue31a"
	case "50n":
		fallthrough
	case "50d":
		icon = "\ue313"
	}
	d.Weather = icon
	d.units = units
	d.UnitIcon = "\ue33e"
	switch d.units {
	case "imperial":
		d.UnitIcon = "°F" // \ue341"
	case "metric":
		d.UnitIcon = "°C" // \ue339"
	case "":
		fallthrough
	case "standard":
		d.UnitIcon = "°K" // <b>K</b>"
	}
	return nil
}
