package location

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"net"
)

//go:generate go run ./generator/generator.go --dbname GeoLite2-Country.mmdb

type staticResolver struct {
	db *geoip2.Reader
}

// NewResolver returns Resolver which uses country database
func StaticResolver() Resolver {

	dbBytes, err := base64.RawStdEncoding.DecodeString(dbData)
	if err != nil {
		fmt.Println(err.Error())
	}

	db, _ := geoip2.FromBytes(dbBytes)
	return &staticResolver{db: db}
}

// ResolveCountry maps given ip to country
func (r *staticResolver) ResolveCountry(ip string) (string, error) {

	ipObject := net.ParseIP(ip)
	if ipObject == nil {
		return "", errors.New("failed to parse IP")
	}

	countryRecord, err := r.db.Country(ipObject)
	if err != nil {
		return "", err
	}

	country := countryRecord.Country.IsoCode
	if country == "" {
		country = countryRecord.RegisteredCountry.IsoCode
		if country == "" {
			return "", errors.New("failed to resolve country")
		}
	}

	return country, nil
}
