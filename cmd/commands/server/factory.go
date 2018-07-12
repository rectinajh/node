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

package server

import (
	"github.com/ethereum/go-ethereum/accounts/keystore"
	identity_handler "github.com/mysterium/node/cmd/commands/server/identity"
	"github.com/mysterium/node/communication"
	nats_dialog "github.com/mysterium/node/communication/nats/dialog"
	nats_discovery "github.com/mysterium/node/communication/nats/discovery"
	"github.com/mysterium/node/identity"
	"github.com/mysterium/node/ip"
	"github.com/mysterium/node/location"
	"github.com/mysterium/node/metadata"
	"github.com/mysterium/node/nat"
	"github.com/mysterium/node/openvpn"
	"github.com/mysterium/node/openvpn/middlewares/server/auth"
	"github.com/mysterium/node/openvpn/middlewares/state"
	openvpn_session "github.com/mysterium/node/openvpn/session"
	"github.com/mysterium/node/openvpn/tls"
	"github.com/mysterium/node/server"
	"github.com/mysterium/node/session"
	"path/filepath"
	"sync"
)

// NewCommand function creates new server command by given options
func NewCommand(options CommandOptions) *Command {
	networkDefinition := getNetworkDefinition(options)
	return NewCommandWith(
		options,
		networkDefinition,
		server.NewClient(networkDefinition.DiscoveryAPIAddress),
		ip.NewResolver(options.IpifyUrl),
		nat.NewService(),
	)
}

// NewCommandWith function creates new client command by given options + injects given dependencies
func NewCommandWith(
	options CommandOptions,
	networkDefinition metadata.NetworkDefinition,
	mysteriumClient server.Client,
	ipResolver ip.Resolver,
	natService nat.NATService,
) *Command {

	keystoreDirectory := filepath.Join(options.DirectoryData, "keystore")
	keystoreInstance := keystore.NewKeyStore(keystoreDirectory, keystore.StandardScryptN, keystore.StandardScryptP)
	createSigner := func(id identity.Identity) identity.Signer {
		return identity.NewSigner(keystoreInstance, id)
	}

	identityHandler := identity_handler.NewHandler(
		identity.NewIdentityManager(keystoreInstance),
		mysteriumClient,
		identity.NewIdentityCache(keystoreDirectory, "remember.json"),
		createSigner,
	)

	//TODO this mess is fixed with another PR
	var locationResolver location.Resolver
	if options.LocationCountry != "" {
		locationResolver = location.NewStaticResolver(options.LocationCountry)
	} else if options.LocationDatabaseBuiltIn {
		locationResolver = location.NewBuiltInResolver()
	} else {
		locationResolver = location.NewExternalDbResolver(filepath.Join(options.DirectoryConfig, options.LocationDatabase))
	}

	locationDetector := location.NewDetector(ipResolver, locationResolver)

	return &Command{
		identityLoader: func() (identity.Identity, error) {
			return identity_handler.LoadIdentity(identityHandler, options.Identity, options.Passphrase)
		},
		createSigner:     createSigner,
		locationDetector: locationDetector,
		ipResolver:       ipResolver,
		mysteriumClient:  mysteriumClient,
		natService:       natService,
		dialogWaiterFactory: func(myID identity.Identity) communication.DialogWaiter {
			return nats_dialog.NewDialogWaiter(
				nats_discovery.NewAddressGenerate(networkDefinition.BrokerAddress, myID),
				identity.NewSigner(keystoreInstance, myID),
			)
		},

		sessionManagerFactory: func(primitives *tls.Primitives, vpnServerIP string) session.Manager {
			// TODO: check options for --openvpn-transport option
			clientConfigGenerator := openvpn.NewClientConfigGenerator(
				primitives,
				vpnServerIP,
				options.OpenvpnPort,
				options.Protocol,
			)

			return session.NewManager(
				session.ServiceConfigProvider(clientConfigGenerator),
				&session.UUIDGenerator{},
			)
		},
		vpnServerFactory: func(manager session.Manager, primitives *tls.Primitives, callback state.Callback) openvpn.Process {
			// TODO: check options for --openvpn-transport option
			serverConfigGenerator := openvpn.NewServerConfigGenerator(
				options.DirectoryRuntime,
				options.DirectoryConfig,
				primitives,
				options.OpenvpnPort,
				options.Protocol,
			)

			ovpnSessionManager := openvpn_session.NewManager(manager)
			sessionValidator := openvpn_session.NewValidator(ovpnSessionManager, identity.NewExtractor())

			return openvpn.NewServer(
				options.OpenvpnBinary,
				serverConfigGenerator,
				auth.NewMiddleware(sessionValidator.Validate, sessionValidator.Cleanup),
				state.NewMiddleware(callback),
			)
		},
		checkOpenvpn: func() error {
			return openvpn.CheckOpenvpnBinary(options.OpenvpnBinary)
		},
		protocol:                    options.Protocol,
		proposalAnnouncementStopped: &sync.WaitGroup{},
	}
}

// TODO this function can be aligned with client function when client and server options will merge into
func getNetworkDefinition(options CommandOptions) metadata.NetworkDefinition {
	network := metadata.DefaultNetwork

	switch {
	case options.Localnet:
		network = metadata.LocalnetDefinition
	}

	//override defined values one by one from options
	if options.DiscoveryAPIAddress != metadata.DefaultNetwork.DiscoveryAPIAddress {
		network.DiscoveryAPIAddress = options.DiscoveryAPIAddress
	}

	if options.BrokerAddress != metadata.DefaultNetwork.BrokerAddress {
		network.BrokerAddress = options.BrokerAddress
	}
	return network
}
