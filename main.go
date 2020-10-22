// by wandy@fortinet.com
package main

import (
	"github.com/pulumi/pulumi-alicloud/sdk/v2/go/alicloud"
	"github.com/pulumi/pulumi-alicloud/sdk/v2/go/alicloud/ecs"
	"github.com/pulumi/pulumi-alicloud/sdk/v2/go/alicloud/vpc"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

type infrastructure struct {
	vpcnetwork        *vpc.Network
	vswitch           *vpc.Switch
	securitygroup     *ecs.SecurityGroup
	fortigateinstance *ecs.Instance
	sgrule            *ecs.SecurityGroupRule
}

func createInfrastructure(ctx *pulumi.Context) (*infrastructure, error) {

	description := "asfasdf jalsdfj alsdfj alsf asdf asfjaksfj kasf "
	name := "vpc_name_test"
	cidr := "192.168.0.0/16"
	vswitch1_cidr := "192.168.0.0/24"
	myinstanceType := "ecs.s6-c1m2.small"
	fortigate622ImageId := "m-2zeikw0lpc9k4fubwplw"

	vpcArgs := &vpc.NetworkArgs{
		CidrBlock:   pulumi.String(cidr),
		Description: pulumi.String(description),
		Name:        pulumi.String(name),
	}

	myvpc, err := vpc.NewNetwork(ctx, "vpc-1", vpcArgs)
	if err != nil {
		return nil,err
	}

	zonesDs, err := alicloud.GetZones(ctx, &alicloud.GetZonesArgs{
		AvailableInstanceType: &myinstanceType,
	})

	if err != nil {
		return nil,err
	}

	az1 := zonesDs.Zones[0].Id

	if az1 == "" {
		az1 = "cn-beijing-b"
	}

	vswitch, err := vpc.NewSwitch(ctx, "switch-1", &vpc.SwitchArgs{
		VpcId:            myvpc.ID(),
		AvailabilityZone: pulumi.String(az1),
		CidrBlock:        pulumi.String(vswitch1_cidr),
	})

	if err != nil {
		return nil,err
	}

	group, err := ecs.NewSecurityGroup(ctx, "web-secgrp", &ecs.SecurityGroupArgs{
		VpcId: myvpc.ID(),
	})

	if err != nil {
		return nil,err
	}

	sgrule1, err := ecs.NewSecurityGroupRule(ctx, "sg-rule1", &ecs.SecurityGroupRuleArgs{
		SecurityGroupId: group.ID(),
		IpProtocol:      pulumi.String("tcp"),
		Type:            pulumi.String("ingress"),
		PortRange:       pulumi.String("22/22"),
		CidrIp:          pulumi.String("0.0.0.0/0"),
	})

	if err != nil {
		return nil,err
	}

	myinstance, err := ecs.NewInstance(ctx, "instance-1", &ecs.InstanceArgs{
		AvailabilityZone:        pulumi.String(az1),
		ImageId:                 pulumi.String(fortigate622ImageId),
		VswitchId:               vswitch.ID(),
		InternetMaxBandwidthOut: pulumi.Int(5),
		InstanceType:            pulumi.String(myinstanceType),
		SecurityGroups:          pulumi.StringArray{group.ID()},
		InstanceName:            pulumi.String("myinstanceCreatedbyPulumi"),
		InstanceChargeType:      pulumi.String("PostPaid"),
		DryRun:                  pulumi.Bool(true),
	})

	if err != nil {
		return nil,err
	}

	return &infrastructure{
		vpcnetwork:        myvpc,
		vswitch:           vswitch,
		securitygroup:     group,
		fortigateinstance: myinstance,
		sgrule: sgrule1,
	}, nil
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		infra, err := createInfrastructure(ctx)
		if err != nil {
			return err
		}

		ctx.Export("vpcName", infra.vpcnetwork.ID())
		ctx.Export("vswitchName", infra.vswitch.ID())
		ctx.Export("sgName", infra.securitygroup.ID())
		ctx.Export("myinstance", infra.fortigateinstance.ID())
		ctx.Export("myinstancepublicIP", infra.fortigateinstance.PublicIp)
		ctx.Export("securtygrouprole",infra.sgrule.ID())
		return nil
	})
}
