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

type staticResolver struct {
	country string
	error   error
}

// NewStaticResolver returns staticResolver which uses statically entered value
func NewStaticResolver(country string) *staticResolver {
	return &staticResolver{
		country: country,
		error:   nil,
	}
}

// NewFailingResolver returns staticResolver with entered error
func NewFailingResolver(err error) *staticResolver {
	return &staticResolver{
		country: "",
		error:   err,
	}
}

// ResolveCountry maps given ip to country
func (d *staticResolver) ResolveCountry(ip string) (string, error) {
	return d.country, d.error
}
