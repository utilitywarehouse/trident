// Copyright 2020 NetApp, Inc. All Rights Reserved.

package ontap

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"testing"

	"github.com/google/uuid"

	"github.com/netapp/trident/utils"
	"github.com/prometheus/common/log"
	"github.com/stretchr/testify/assert"

	tridentconfig "github.com/netapp/trident/config"
	drivers "github.com/netapp/trident/storage_drivers"
	"github.com/netapp/trident/storage_drivers/ontap/api"
)

func TestOntapSanStorageDriverConfigString(t *testing.T) {

	vserverAdminHost := ONTAPTEST_LOCALHOST
	vserverAdminPort := strconv.Itoa(rand.Intn(ONTAPTEST_SERVER_MAX_PORT-ONTAPTEST_SERVER_MIN_PORT) +
		ONTAPTEST_SERVER_MIN_PORT)
	vserverAggrName := ONTAPTEST_VSERVER_AGGR_NAME

	var ontapSanDrivers = []SANStorageDriver{

		*newTestOntapSANDriver(vserverAdminHost, vserverAdminPort, vserverAggrName),
	}

	sensitiveIncludeList := map[string]string{
		"username":        "ontap-san-user",
		"password":        "password1!",
		"client username": "client_username",
		"client password": "client_password",
	}

	externalIncludeList := map[string]string{
		"<REDACTED>":                   "<REDACTED>",
		"username":                     "Username:<REDACTED>",
		"password":                     "Password:<REDACTED>",
		"api":                          "API:<REDACTED>",
		"chap username":                "ChapUsername:<REDACTED>",
		"chap initiator secret":        "ChapInitiatorSecret:<REDACTED>",
		"chap target username":         "ChapTargetUsername:<REDACTED>",
		"chap target initiator secret": "ChapTargetInitiatorSecret:<REDACTED>",
		"client private key":           "ClientPrivateKey:<REDACTED>",
	}

	for _, ontapSanDriver := range ontapSanDrivers {
		for key, val := range externalIncludeList {
			assert.Contains(t, ontapSanDriver.String(), val,
				"ontap-san driver does not contain %v", key)
			assert.Contains(t, ontapSanDriver.GoString(), val,
				"ontap-san driver does not contain %v", key)
		}

		for key, val := range sensitiveIncludeList {
			assert.NotContains(t, ontapSanDriver.String(), val,
				"ontap-san driver contains %v", key)
			assert.NotContains(t, ontapSanDriver.GoString(), val,
				"ontap-san driver contains %v", key)
		}
	}
}

func newTestOntapSANDriver(vserverAdminHost, vserverAdminPort, vserverAggrName string) *SANStorageDriver {
	config := &drivers.OntapStorageDriverConfig{}
	sp := func(s string) *string { return &s }

	config.CommonStorageDriverConfig = &drivers.CommonStorageDriverConfig{}
	config.CommonStorageDriverConfig.DebugTraceFlags = make(map[string]bool)
	config.CommonStorageDriverConfig.DebugTraceFlags["method"] = true
	config.CommonStorageDriverConfig.DebugTraceFlags["api"] = true
	// config.Labels = map[string]string{"app": "wordpress"}
	config.ManagementLIF = vserverAdminHost + ":" + vserverAdminPort
	config.SVM = "SVM1"
	config.Aggregate = vserverAggrName
	config.Username = "ontap-san-user"
	config.Password = "password1!"
	config.StorageDriverName = "ontap-san"
	config.StoragePrefix = sp("test_")

	sanDriver := &SANStorageDriver{}
	sanDriver.Config = *config

	// ClientConfig holds the configuration data for Client objects
	clientConfig := api.ClientConfig{
		ManagementLIF:           config.ManagementLIF,
		SVM:                     "SVM1",
		Username:                "client_username",
		Password:                "client_password",
		DriverContext:           tridentconfig.ContextCSI,
		ContextBasedZapiRecords: 100,
		DebugTraceFlags:         config.CommonStorageDriverConfig.DebugTraceFlags,
	}

	sanDriver.API = api.NewClient(clientConfig)
	sanDriver.Telemetry = &Telemetry{
		Plugin:        sanDriver.Name(),
		SVM:           sanDriver.GetConfig().SVM,
		StoragePrefix: *sanDriver.GetConfig().StoragePrefix,
		Driver:        sanDriver,
	}

	return sanDriver
}

func TestOntapSanReconcileNodeAccess(t *testing.T) {
	ctx := context.Background()

	vserverAdminHost := ONTAPTEST_LOCALHOST
	vserverAdminPort := strconv.Itoa(rand.Intn(ONTAPTEST_SERVER_MAX_PORT-ONTAPTEST_SERVER_MIN_PORT) +
		ONTAPTEST_SERVER_MIN_PORT)
	vserverAggrName := ONTAPTEST_VSERVER_AGGR_NAME

	server := newUnstartedVserver(ctx, vserverAdminHost, vserverAdminPort, vserverAggrName)
	server.StartTLS()

	_, port, err := net.SplitHostPort(server.Listener.Addr().String())
	assert.Nil(t, err, "Unable to get Web host port %s", port)

	defer func() {
		if r := recover(); r != nil {
			server.Close()
			log.Error("Panic in fake filer", r)
		}
	}()

	cases := [][]struct {
		igroupName         string
		igroupExistingIQNs []string
		nodes              []*utils.Node
		igroupFinalIQNs    []string
	}{
		// Add a backend
		{
			{
				igroupName:         "igroup1",
				igroupExistingIQNs: []string{},
				nodes: []*utils.Node{
					{
						Name: "node1",
						IQN:  "IQN1",
					},
					{
						Name: "node2",
						IQN:  "IQN2",
					},
				},
				igroupFinalIQNs: []string{"IQN1", "IQN2"},
			},
		},
		// 2 same cluster backends/ nodes unchanged - both current
		{
			{
				igroupName:         "igroup1",
				igroupExistingIQNs: []string{"IQN1", "IQN2"},
				nodes: []*utils.Node{
					{
						Name: "node1",
						IQN:  "IQN1",
					},
					{
						Name: "node2",
						IQN:  "IQN2",
					},
				},
				igroupFinalIQNs: []string{"IQN1", "IQN2"},
			},
			{
				igroupName:         "igroup2",
				igroupExistingIQNs: []string{"IQN3", "IQN4"},
				nodes: []*utils.Node{
					{
						Name: "node3",
						IQN:  "IQN3",
					},
					{
						Name: "node4",
						IQN:  "IQN4",
					},
				},
				igroupFinalIQNs: []string{"IQN3", "IQN4"},
			},
		},
		// 2 different cluster backends - add node
		{
			{
				igroupName:         "igroup1",
				igroupExistingIQNs: []string{"IQN1"},
				nodes: []*utils.Node{
					{
						Name: "node1",
						IQN:  "IQN1",
					},
					{
						Name: "node2",
						IQN:  "IQN2",
					},
				},
				igroupFinalIQNs: []string{"IQN1", "IQN2"},
			},
			{
				igroupName:         "igroup2",
				igroupExistingIQNs: []string{"IQN3", "IQN4"},
				nodes: []*utils.Node{
					{
						Name: "node3",
						IQN:  "IQN3",
					},
					{
						Name: "node4",
						IQN:  "IQN4",
					},
				},
				igroupFinalIQNs: []string{"IQN3", "IQN4"},
			},
		},
		// 2 different cluster backends - remove node
		{
			{
				igroupName:         "igroup1",
				igroupExistingIQNs: []string{"IQN1", "IQN2"},
				nodes: []*utils.Node{
					{
						Name: "node1",
						IQN:  "IQN1",
					},
				},
				igroupFinalIQNs: []string{"IQN1"},
			},
			{
				igroupName:         "igroup2",
				igroupExistingIQNs: []string{"IQN3", "IQN4"},
				nodes: []*utils.Node{
					{
						Name: "node3",
						IQN:  "IQN3",
					},
					{
						Name: "node4",
						IQN:  "IQN4",
					},
				},
				igroupFinalIQNs: []string{"IQN3", "IQN4"},
			},
		},
	}

	for _, testCase := range cases {

		igroups = map[string]map[string]struct{}{}

		var ontapSanDrivers []SANStorageDriver

		for _, driverInfo := range testCase {

			// simulate existing IQNs on the vserver
			igroupsIQNMap := map[string]struct{}{}
			for _, iqn := range driverInfo.igroupExistingIQNs {
				igroupsIQNMap[iqn] = struct{}{}
			}

			igroups[driverInfo.igroupName] = igroupsIQNMap

			sanStorageDriver := newTestOntapSANDriver(vserverAdminHost, vserverAdminPort, vserverAggrName)
			sanStorageDriver.Config.IgroupName = driverInfo.igroupName
			ontapSanDrivers = append(ontapSanDrivers, *sanStorageDriver)
		}

		for driverIndex, driverInfo := range testCase {
			ontapSanDrivers[driverIndex].ReconcileNodeAccess(ctx, driverInfo.nodes,
				uuid.New().String())
		}

		for _, driverInfo := range testCase {

			assert.Equal(t, len(driverInfo.igroupFinalIQNs), len(igroups[driverInfo.igroupName]))

			for _, iqn := range driverInfo.igroupFinalIQNs {
				assert.Contains(t, igroups[driverInfo.igroupName], iqn)
			}
		}
	}
}
func TestOntapSanTerminate(t *testing.T) {
	ctx := context.Background()

	vserverAdminHost := ONTAPTEST_LOCALHOST
	vserverAdminPort := strconv.Itoa(rand.Intn(ONTAPTEST_SERVER_MAX_PORT-ONTAPTEST_SERVER_MIN_PORT) +
		ONTAPTEST_SERVER_MIN_PORT)
	vserverAggrName := ONTAPTEST_VSERVER_AGGR_NAME

	server := newUnstartedVserver(ctx, vserverAdminHost, vserverAdminPort, vserverAggrName)
	server.StartTLS()

	_, port, err := net.SplitHostPort(server.Listener.Addr().String())
	assert.Nil(t, err, "Unable to get Web host port %s", port)

	defer func() {
		if r := recover(); r != nil {
			server.Close()
			log.Error("Panic in fake filer", r)
		}
	}()

	cases := [][]struct {
		igroupName         string
		igroupExistingIQNs []string
	}{
		// 2 different cluster backends - remove backend
		{
			{
				igroupName:         "igroup1",
				igroupExistingIQNs: []string{"IQN1", "IQN2"},
			},
			{
				igroupName:         "igroup2",
				igroupExistingIQNs: []string{"IQN3", "IQN4"},
			},
		},
		{
			{
				igroupName:         "igroup1",
				igroupExistingIQNs: []string{},
			},
		},
	}

	for _, testCase := range cases {

		igroups = map[string]map[string]struct{}{}

		var ontapSanDrivers []SANStorageDriver

		for _, driverInfo := range testCase {

			// simulate existing IQNs on the vserver
			igroupsIQNMap := map[string]struct{}{}
			for _, iqn := range driverInfo.igroupExistingIQNs {
				igroupsIQNMap[iqn] = struct{}{}
			}

			igroups[driverInfo.igroupName] = igroupsIQNMap

			sanStorageDriver := newTestOntapSANDriver(vserverAdminHost, vserverAdminPort, vserverAggrName)
			sanStorageDriver.Config.IgroupName = driverInfo.igroupName
			sanStorageDriver.Telemetry = nil
			ontapSanDrivers = append(ontapSanDrivers, *sanStorageDriver)
		}

		for driverIndex, driverInfo := range testCase {
			ontapSanDrivers[driverIndex].Terminate(ctx, "")
			assert.NotContains(t, igroups, igroups[driverInfo.igroupName])
		}

	}
}
