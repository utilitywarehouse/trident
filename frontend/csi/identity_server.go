// Copyright 2020 NetApp, Inc. All Rights Reserved.

package csi

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/netapp/trident/utils"
)

func (p *Plugin) Probe(
	ctx context.Context, req *csi.ProbeRequest,
) (*csi.ProbeResponse, error) {

	logc := utils.GetLogWithRequestContext(ctx)

	fields := log.Fields{"Method": "Probe", "Type": "CSI_Identity"}
	logc.WithFields(fields).Debug(">>>> Probe")
	defer logc.WithFields(fields).Debug("<<<< Probe")

	// Ensure Trident bootstrapped OK.  We only return an error if Trident bootstrapping
	// failed (i.e. unrecoverable), not if Trident is still initializing.
	_, err := p.orchestrator.GetVersion(ctx)
	if utils.IsBootstrapError(err) {
		return &csi.ProbeResponse{}, status.Error(codes.FailedPrecondition, err.Error())
	}

	return &csi.ProbeResponse{}, nil
}

func (p *Plugin) GetPluginInfo(
	ctx context.Context, req *csi.GetPluginInfoRequest,
) (*csi.GetPluginInfoResponse, error) {

	logc := utils.GetLogWithRequestContext(ctx)

	fields := log.Fields{"Method": "GetPluginInfo", "Type": "CSI_Identity"}
	logc.WithFields(fields).Debug(">>>> GetPluginInfo")
	defer logc.WithFields(fields).Debug("<<<< GetPluginInfo")

	return &csi.GetPluginInfoResponse{
		Name:          p.name,
		VendorVersion: p.version,
	}, nil
}

func (p *Plugin) GetPluginCapabilities(
	ctx context.Context, req *csi.GetPluginCapabilitiesRequest,
) (*csi.GetPluginCapabilitiesResponse, error) {

	logc := utils.GetLogWithRequestContext(ctx)

	fields := log.Fields{"Method": "GetPluginCapabilities", "Type": "CSI_Identity"}
	logc.WithFields(fields).Debug(">>>> GetPluginCapabilities")
	defer logc.WithFields(fields).Debug("<<<< GetPluginCapabilities")

	return &csi.GetPluginCapabilitiesResponse{
		Capabilities: []*csi.PluginCapability{
			{
				Type: &csi.PluginCapability_Service_{
					Service: &csi.PluginCapability_Service{
						Type: csi.PluginCapability_Service_CONTROLLER_SERVICE,
					},
				},
			},
		},
	}, nil
}
