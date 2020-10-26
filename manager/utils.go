package manager

import (
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	fabApi "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config/lookup"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/test/metadata"
	"github.com/pkg/errors"
)

func simpleFabricSdk(confpath string) *fabsdk.FabricSDK {
	confProvider := config.FromFile(confpath)
	sdk, err := fabsdk.New(confProvider)
	if err != nil {
		panic(err)
	}
	return sdk
}

func simpleLedgerClient(sdk *fabsdk.FabricSDK, channelID, user, org string) *ledger.Client {
	//prepare required contexts
	channelClientCtx := sdk.ChannelContext(channelID, fabsdk.WithUser(user), fabsdk.WithOrg(org))

	// Get a ledger client.
	ledgerClient, err := ledger.New(channelClientCtx)
	if err != nil {
		panic(err)
	}

	return ledgerClient
}

func newLedgerClient1(channelID, user, org, ordererEndpoint, channelConfigTxFile string) {
	sdk, err := fabsdk.New(ConfigBackend)
	if err != nil {
		panic(err)
	}

	// Get signing identity that is used to sign create channel request
	orgMspClient, err := mspclient.New(sdk.Context(), mspclient.WithOrg(org))
	if err != nil {
		panic(err)
	}

	adminIdentity, err := orgMspClient.GetSigningIdentity(user)
	if err != nil {
		panic(err)
	}

	configBackend, err := sdk.Config()
	if err != nil {
		panic(err)
	}

	targets, err := OrgTargetPeers([]string{org}, configBackend)
	if err != nil {
		panic(err)
	}

	req := resmgmt.SaveChannelRequest{
		ChannelID:         channelID,
		ChannelConfigPath: GetChannelConfigTxPath(channelConfigTxFile),
		SigningIdentities: []msp.SigningIdentity{adminIdentity},
	}
	err = InitializeChannel(sdk, org, user, ordererEndpoint, req, targets)
	if err != nil {
		panic(err)
	}
	// sdk, targets
}

// InitializeChannel ...
func InitializeChannel(sdk *fabsdk.FabricSDK, orgID, user, ordererEndpoint string, req resmgmt.SaveChannelRequest, targets []string) error {

	joinedTargets, err := FilterTargetsJoinedChannel(sdk, orgID, req.ChannelID, targets)
	if err != nil {
		return errors.WithMessage(err, "checking for joined targets failed")
	}

	if len(joinedTargets) != len(targets) {
		_, err := SaveChannel(sdk, req, orgID, user, ordererEndpoint)
		if err != nil {
			return errors.Wrapf(err, "create channel failed")
		}

		_, err = JoinChannel(sdk, req.ChannelID, orgID, ordererEndpoint, targets)
		if err != nil {
			return errors.Wrapf(err, "join channel failed")
		}
	}
	return nil
}

// FilterTargetsJoinedChannel filters targets to those that have joined the named channel.
func FilterTargetsJoinedChannel(sdk *fabsdk.FabricSDK, orgID string, channelID string, targets []string) ([]string, error) {
	var joinedTargets []string

	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgID))

	rc, err := resmgmt.New(clientContext)
	if err != nil {
		return nil, errors.WithMessage(err, "failed getting admin user session for org")
	}

	for _, target := range targets {
		// Check if primary peer has joined channel
		alreadyJoined, err := HasPeerJoinedChannel(rc, target, channelID)
		if err != nil {
			return nil, errors.WithMessage(err, "failed while checking if primary peer has already joined channel")
		}
		if alreadyJoined {
			joinedTargets = append(joinedTargets, target)
		}
	}
	return joinedTargets, nil
}

// HasPeerJoinedChannel checks whether the peer has already joined the channel.
// It returns true if it has, false otherwise, or an error
func HasPeerJoinedChannel(client *resmgmt.Client, target string, channel string) (bool, error) {
	foundChannel := false
	response, err := client.QueryChannels(resmgmt.WithTargetEndpoints(target), resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return false, errors.WithMessage(err, "failed to query channel for peer")
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == channel {
			foundChannel = true
		}
	}

	return foundChannel, nil
}

// SaveChannel attempts to save (create or update) the named channel.
func SaveChannel(sdk *fabsdk.FabricSDK, req resmgmt.SaveChannelRequest, org, user, ordererEndpoint string) (bool, error) {

	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser(user), fabsdk.WithOrg(org))

	// Channel management client is responsible for managing channels (create/update)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create new channel management client")
	}

	// Create channel (or update if it already exists)
	if _, err = resMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(ordererEndpoint)); err != nil {
		return false, err
	}

	return true, nil
}

// JoinChannel attempts to save the named channel.
func JoinChannel(sdk *fabsdk.FabricSDK, channelID, orgID, ordererEndpoint string, targets []string) (bool, error) {
	//prepare context
	clientContext := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgID))

	// Resource management client is responsible for managing resources (joining channels, install/instantiate/upgrade chaincodes)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return false, errors.WithMessage(err, "Failed to create new resource management client")
	}

	if err := resMgmtClient.JoinChannel(
		channelID,
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
		resmgmt.WithTargetEndpoints(targets...),
		resmgmt.WithOrdererEndpoint(ordererEndpoint)); err != nil {
		return false, nil
	}
	return true, nil
}

// OrgTargetPeers determines peer endpoints for orgs
func OrgTargetPeers(orgs []string, configBackend ...core.ConfigBackend) ([]string, error) {
	networkConfig := fabApi.NetworkConfig{}
	err := lookup.New(configBackend...).UnmarshalKey("organizations", &networkConfig.Organizations)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get organizations from config ")
	}

	var peers []string
	for _, org := range orgs {
		orgConfig, ok := networkConfig.Organizations[strings.ToLower(org)]
		if !ok {
			continue
		}
		peers = append(peers, orgConfig.Peers...)
	}
	return peers, nil
}

func GetChannelConfigTxPath(filename string) string {
	return filepath.Join(metadata.GetProjectPath(), metadata.ChannelConfigPath, filename)
}
