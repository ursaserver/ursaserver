package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/ursaserver/ursa"
)

type Conf struct {
	Upstream string
	Routes   []Route
	// TODO, add ability to specify a file to use as logfile
	// Logfile           string
	CustomRates       map[string]CustomRate
	JWTAuthHeaderName string
	JWTAuthGetUserBy  string
}

type Route struct {
	Methods []string
	Pattern string
	Rates   map[string]string
}

type CustomRate struct {
	Header              string
	ValidIfMatchesRegex string
	ValidIfIn           []string
	FailCode            int
	FailMsg             string
}

var MethodCheckerRegex = regexp.MustCompile("^[A-Za-z]+$")

// The rate by field name to use to refer to rate by JWT token
const (
	JSTRateBy = "JWT"
	IPRateBy  = "IP"
)

// Check if the configuration provided is valid, returning nil if valid and an
// error otherwise
func CheckConf(c *Conf) error {
	// Check that the upstream url is valid
	url, err := url.Parse(c.Upstream)
	if err != nil {
		return fmt.Errorf("error parsing upstream url %v. %v", c.Upstream, err)
	}
	if url.String() == "" {
		return fmt.Errorf("url host is empty %q", url)
	}
	// Check that the CustomRates are valid
	validRateNames := []string{JSTRateBy, IPRateBy}
	for k, v := range c.CustomRates {
		validRateNames = append(validRateNames, k)
		// Check if a given custom rate is valid
		if _, err := CustomRateToRateBy(v); err != nil {
			return fmt.Errorf("got error %v in CustomRate %v", err, v)
		}
	}
	// Fail if there are no routes defined
	if len(c.Routes) == 0 {
		return fmt.Errorf("no routes defined in configuration file")
	}
	// Check that all routes are valid
	for _, route := range c.Routes {
		// Check that the pattern regex is valid
		if _, err := regexp.Compile(route.Pattern); err != nil {
			return fmt.Errorf("cannot compile regex Pattern for %v", route)
		}
		// Check if no methods defined
		if len(route.Methods) == 0 {
			return fmt.Errorf("no allowed methods defined for route %q", route)
		}

		// Ensure that the methods names are nothing but alphabetical characters
		for _, method := range route.Methods {
			if !MethodCheckerRegex.MatchString(method) {
				return fmt.Errorf("method name %v is invalid for route %v", method, route)
			}
		}
		// Check that all the rates are valid
		for k, v := range route.Rates {
			if !In(validRateNames, k) {
				return fmt.Errorf("given rate by %v is invalid for route %v. Valid rate keys are %v", k, route, validRateNames)
			}
			if _, err := RateStringToRate(v); err != nil {
				return fmt.Errorf("error reading rate for route %v. error %v", route, err)
			}
		}
	}
	// TODO
	// Ensure that if rate by JWT is being used, the JWT header name, uid field
	// name, and JSWT secret is provided
	return nil
}

// Creates a RateByHeader based on CustomRate given. Returns (RateByHeader, error)
func CustomRateToRateBy(c CustomRate) (ursa.RateBy, error) {
	var rateByHeader ursa.RateBy
	if c.Header == JSTRateBy || c.Header == IPRateBy {
		return rateByHeader, fmt.Errorf("%v is a predefined header name thus cannot be redefined", c.Header)
	}
	if len(c.ValidIfIn) == 0 && len(c.ValidIfMatchesRegex) == 0 {
		return rateByHeader, fmt.Errorf("both ValidIfIn and ValidIfMatchesRegex have zero values for rate %v", c)
	} else if len(c.ValidIfIn) > 0 && len(c.ValidIfMatchesRegex) > 0 {
		return rateByHeader, fmt.Errorf("both ValidIfIn and ValidIfMatchesRegex are defined for rate %v", c)
	}
	if len(c.ValidIfMatchesRegex) > 0 {
		if _, err := regexp.Compile(c.ValidIfMatchesRegex); err != nil {
			return rateByHeader, fmt.Errorf("error compiling ValidIfMatchesRegex %v for  rate %v", c.ValidIfMatchesRegex, c)
		}
	}
	var validFn func(string) bool
	if len(c.ValidIfMatchesRegex) > 0 {
		validFn = func(s string) bool {
			return regexp.MustCompile(c.ValidIfMatchesRegex).MatchString(s)
		}
	} else {
		validFn = func(s string) bool {
			return In(c.ValidIfIn, s)
		}
	}
	rateByHeader = *ursa.NewRateBy(c.Header,
		validFn,
		func(s string) string { return s },
		c.FailCode,
		c.FailMsg)
	return rateByHeader, nil
}

// Creates a Rate based on rate string. Returns (Rate, error)
func RateStringToRate(r string) (ursa.Rate, error) {
	var rate ursa.Rate
	s := SanitizeRateString(r)
	parts := strings.Split(s, "/")
	if len(parts) < 2 {
		return rate, fmt.Errorf("rate not in proper format example: 15/minute")
	}
	capacity, err := strconv.Atoi(parts[0])
	if err != nil {
		return rate, fmt.Errorf("error reading the integer capacity for rate got %v", parts[0])
	}
	switch strings.ToUpper(parts[1]) {
	case "MINUTE":
		return ursa.NewRate(capacity, ursa.Minute), nil
	case "HOUR":
		return ursa.NewRate(capacity, ursa.Hour), nil
	case "DAY":
		return ursa.NewRate(capacity, ursa.Day), nil
	}
	return rate, fmt.Errorf("valid time units are minute/hour/day got %v", parts[1])
}

// Remove any whitespace from the rate string
func SanitizeRateString(r string) string {
	return strings.ReplaceAll(r, " ", "")
}

// Assume that configuration provided as input is valid, where the validity is
// determined by the CheckConf function
func confToUrsaConf(c Conf) ursa.Conf {
	var ursaConf ursa.Conf
	// Setup URL
	url, _ := url.Parse(c.Upstream)
	ursaConf.Upstream = url
	// Create RateBys to make routes
	rateBys := make(map[string]*ursa.RateBy)
	for name, value := range c.CustomRates {
		r, _ := CustomRateToRateBy(value)
		rateBys[name] = &r
	}
	rateBys[IPRateBy] = ursa.RateByIP
	// TODO
	// Rate by JWT Auth Token

	routes := make([]ursa.Route, 0)
	for _, c := range c.Routes {
		var route ursa.Route
		route.Methods = c.Methods
		route.Pattern = regexp.MustCompile(c.Pattern)
		rates := ursa.RouteRates{}
		for rateByName, rateValue := range c.Rates {
			rateBy := rateBys[rateByName]
			rate, _ := RateStringToRate(rateValue)
			rates[rateBy] = rate
		}
		route.Rates = rates
		routes = append(routes, route)
	}
	ursaConf.Routes = routes
	return ursaConf
}
