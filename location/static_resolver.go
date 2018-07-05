/*
 * Copyright (C) 2017 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

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
