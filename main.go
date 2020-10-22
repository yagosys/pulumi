// by wandy@fortinet.com
package main

import (
	"fmt"

	"github.com/pulumi/pulumi-alicloud/sdk/v2/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v2/go/alicloud/ecs"
	"github.com/pulumi/pulumi-alicloud/sdk/v2/go/alicloud/vpc"
	"github.com/pulumi/pulumi-alicloud/sdk/v2/go/alicloud/marketplace"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

type infrastructure struct {
	vpcnetwork        *vpc.Network
	vswitchpub        *vpc.Switch
	vswitchpri1       *vpc.Switch
	securitygroup     *ecs.SecurityGroup
	fortigateinstance *ecs.Instance
	sgrule            *ecs.SecurityGroupRule
	eni               *vpc.NetworkInterface
	rtpri1            *vpc.RouteTable
	eniAtt            *vpc.NetworkInterfaceAttachment
}

func createInfrastructure(ctx *pulumi.Context) (*infrastructure, error) {

	description := "asfasdf jalsdfj alsdfj alsf asdf asfjaksfj kasf "
	name := "vpc_name_test"
	cidr := "192.168.0.0/16"
	vswitchpub1_cidr := "192.168.0.0/24"
	vswitchpri1_cidr := "192.168.1.0/24"
	myinstanceType := "ecs.s6-c1m2.small"
	fortigate624trial := "cmjj00040535"
	//fortigate622ImageId := "m-8vb5o2t7kujhxgnm5tyu"
	
	opt0 := true
        currentRegionDs, err := alicloud.GetRegions(ctx, &alicloud.GetRegionsArgs{
        	Current: &opt0,
        })

       	if err != nil {
        	return nil,err
        }

	defaultRegionId:=currentRegionDs.Regions[0].Id


	_default, err := marketplace.GetProduct(ctx, &marketplace.GetProductArgs{
            ProductCode: fortigate624trial,
	    AvailableRegion: &defaultRegionId,
    	})

	if err !=nil {
		return nil,err
	}
	fortigate622ImageId :=_default.Products[0].Skuses[0].Images[0].ImageId
	fmt.Println(fortigate622ImageId)


	vpcArgs := &vpc.NetworkArgs{
		CidrBlock:   pulumi.String(cidr),
		Description: pulumi.String(description),
		Name:        pulumi.String(name),
	}

	myvpc, err := vpc.NewNetwork(ctx, "vpc-1", vpcArgs)
	if err != nil {
		return nil, err
	}

	zonesDs, err := alicloud.GetZones(ctx, &alicloud.GetZonesArgs{
		AvailableInstanceType: &myinstanceType,
	})

	if err != nil {
		return nil, err
	}

	az1 := zonesDs.Zones[0].Id

	if az1 == "" {
		az1 = "cn-beijing-b"
	}

	vswitchpub, err := vpc.NewSwitch(ctx, "switch-1", &vpc.SwitchArgs{
		VpcId:            myvpc.ID(),
		AvailabilityZone: pulumi.String(az1),
		CidrBlock:        pulumi.String(vswitchpub1_cidr),
	})

	if err != nil {
		return nil, err
	}

	vswitchpri1, err := vpc.NewSwitch(ctx, "switch-private-1", &vpc.SwitchArgs{
		VpcId:            myvpc.ID(),
		AvailabilityZone: pulumi.String(az1),
		CidrBlock:        pulumi.String(vswitchpri1_cidr),
	})

	if err != nil {
		return nil, err
	}

	group, err := ecs.NewSecurityGroup(ctx, "web-secgrp", &ecs.SecurityGroupArgs{
		VpcId: myvpc.ID(),
	})

	if err != nil {
		return nil, err
	}

	sgrule1, err := ecs.NewSecurityGroupRule(ctx, "sg-rule1", &ecs.SecurityGroupRuleArgs{
		SecurityGroupId: group.ID(),
		IpProtocol:      pulumi.String("tcp"),
		Type:            pulumi.String("ingress"),
		PortRange:       pulumi.String("22/22"),
		CidrIp:          pulumi.String("0.0.0.0/0"),
	})

	if err != nil {
		return nil, err
	}

	eni1, err := vpc.NewNetworkInterface(ctx, "firstSecondarylyNetworkInterface", &vpc.NetworkInterfaceArgs{
		VswitchId: vswitchpri1.ID(),
		SecurityGroups: pulumi.StringArray{
			group.ID(),
		},
	})

	if err != nil {
		return nil, err
	}

	rtpri1, err := vpc.NewRouteTable(ctx, "routetableforprivatenetwork", &vpc.RouteTableArgs{
		VpcId: myvpc.ID(),
	})
	if err != nil {
		return nil, err
	}
	_, err = vpc.NewRouteTableAttachment(ctx, "routetableattachtovswithpri1", &vpc.RouteTableAttachmentArgs{
		RouteTableId: rtpri1.ID(),
		VswitchId:    vswitchpri1.ID(),
	})
	if err != nil {
		return nil, err
	}

	myinstance, err := ecs.NewInstance(ctx, "instance-1", &ecs.InstanceArgs{
		AvailabilityZone:        pulumi.String(az1),
		ImageId:                 pulumi.String(fortigate622ImageId),
		VswitchId:               vswitchpub.ID(),
		InternetMaxBandwidthOut: pulumi.Int(5),
		InstanceType:            pulumi.String(myinstanceType),
		SecurityGroups:          pulumi.StringArray{group.ID()},
		InstanceName:            pulumi.String("myinstanceCreatedbyPulumi"),
		InstanceChargeType:      pulumi.String("PostPaid"),
		DryRun:                  pulumi.Bool(false),
		Status:			 pulumi.String("Stopped"),
	})
	if err != nil {
		return nil, err
	}

	eniAtt1, err := vpc.NewNetworkInterfaceAttachment(ctx, "firstsecondaryeniattachtovswitchpri1", &vpc.NetworkInterfaceAttachmentArgs{
		InstanceId:         myinstance.ID(),
		NetworkInterfaceId: eni1.ID(),
	})

	if err != nil {
		return nil, err
	}

	_, err = vpc.NewRouteEntry(ctx, "rtpri1defaultroutetofortigateeni1", &vpc.RouteEntryArgs{
		DestinationCidrblock: pulumi.String("0.0.0.0/0"),
		Name:                 pulumi.String("defaultroute"),
		NexthopId:            eni1.ID(),
		NexthopType:          pulumi.String("NetworkInterface"),
		RouteTableId:         rtpri1.ID(),
	})
	if err != nil {
		return nil, err
	}

	return &infrastructure{
		vpcnetwork:        myvpc,
		vswitchpub:        vswitchpub,
		vswitchpri1:       vswitchpri1,
		securitygroup:     group,
		fortigateinstance: myinstance,
		sgrule:            sgrule1,
		eni:               eni1,
		rtpri1:            rtpri1,
		eniAtt:             eniAtt1,
	}, nil
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		infra, err := createInfrastructure(ctx)
		if err != nil {
			return err
		}

		ctx.Export("vpcName", infra.vpcnetwork.ID())
		ctx.Export("vswitchpubName", infra.vswitchpub.ID())
		ctx.Export("vswitchpri1Name", infra.vswitchpri1.ID())
		ctx.Export("sgName", infra.securitygroup.ID())
		ctx.Export("myinstance", infra.fortigateinstance.ID())
		ctx.Export("myinstancepublicIP", infra.fortigateinstance.PublicIp)
		ctx.Export("securtygrouprole", infra.sgrule.ID())
		ctx.Export("eni1", infra.eni.ID())
		ctx.Export("eniattachment1",infra.eniAtt.ID())
		ctx.Export("eniattachment-instanceID",infra.eniAtt.InstanceId)
		ctx.Export("eniattachment-Interface",infra.eniAtt.NetworkInterfaceId)
		return nil
	})
}
