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

// Note that because the ursa package doesn't expose structs like ursa.rate
// and ursa.rateBy as public, we are having to create our own here. We also
// note that ursa doesn't expose such structs as a mechanism of defensive
// programming.

// Mimicks ursa.rate
type Rate struct {
	Capacity int
	Duration int
}

// Mimicks ursa.rateBy
type RateByHeader struct {
	Name      string
	Valid     ursa.IsValidHeaderValue
	Signature ursa.SignatureFromHeaderValue
	FailCode  int
	FailMsg   string
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
	_, err := url.Parse(c.Upstream)
	if err != nil {
		return fmt.Errorf("error parsing upstream url %v. %v", c.Upstream, err)
	}
	// Check that the CustomRates are valid
	validRateNames := []string{JSTRateBy}
	for k, v := range c.CustomRates {
		validRateNames = append(validRateNames, k)
		// Check if a given custom rate is valid
		if _, err := CustomRateToRateBy(v); err != nil {
			return fmt.Errorf("got error %v in CustomRate %v", err, v)
		}
	}
	// Check that all routes are valid
	for _, route := range c.Routes {
		// Check that the pattern regex is valid
		if _, err := regexp.Compile(route.Pattern); err != nil {
			return fmt.Errorf("cannot compile regex Pattern for %v", route)
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
	return nil
}

// Creates a RateByHeader based on CustomRate given. Returns (RateByHeader, error)
func CustomRateToRateBy(c CustomRate) (RateByHeader, error) {
	var rateByHeader RateByHeader
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
	rateByHeader.Name = c.Header
	rateByHeader.FailCode = c.FailCode
	rateByHeader.FailMsg = c.FailMsg
	// Use whatever is the header value as the signature
	rateByHeader.Signature = func(value string) string { return value }
	if len(c.ValidIfMatchesRegex) > 0 {
		rateByHeader.Valid = func(s string) bool {
			return regexp.MustCompile(c.ValidIfMatchesRegex).MatchString(s)
		}
	} else {
		rateByHeader.Valid = func(s string) bool {
			return In(c.ValidIfIn, s)
		}
	}
	return rateByHeader, nil
}

// Creates a Rate based on rate string. Returns (Rate, error)
func RateStringToRate(r string) (Rate, error) {
	var rate Rate
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
		return Rate{capacity, int(ursa.Minute)}, nil
	case "HOUR":
		return Rate{capacity, int(ursa.Hour)}, nil
	case "DAY":
		return Rate{capacity, int(ursa.Day)}, nil
	}
	return rate, fmt.Errorf("valid time units are minute/hour/day got %v", parts[1])
}

// Remove any whitespace from the rate string
func SanitizeRateString(r string) string {
	return strings.ReplaceAll(r, " ", "")
}
