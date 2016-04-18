package service

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	opsee_aws "github.com/opsee/basic/schema/aws"
	opsee_aws_ec2 "github.com/opsee/basic/schema/aws/ec2"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/bezosphere/store"
	"golang.org/x/net/context"
)

func (s *service) EC2DescribeInstances(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_ec2.DescribeInstancesOutput, error) {
	var (
		input  = req.GetEc2_DescribeInstancesInput()
		output = &opsee_aws_ec2.DescribeInstancesOutput{}
	)

	session, err := s.requestSession(ctx, req, input, output)
	if err != nil {
		return nil, err
	}

	if session.cached {
		return output, nil
	}

	awsRequest, awsOutput := ec2.New(session.session).DescribeInstancesRequest(nil)
	awsRequest.Params = input

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error describing ec2 instances")
		return nil, err
	}

	opsee_aws.CopyInto(output, awsOutput)
	return output, nil
}

func (s *service) EC2DescribeSecurityGroups(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_ec2.DescribeSecurityGroupsOutput, error) {
	var (
		input  = req.GetEc2_DescribeSecurityGroupsInput()
		output = &opsee_aws_ec2.DescribeSecurityGroupsOutput{}
	)

	session, err := s.requestSession(ctx, req, input, output)
	if err != nil {
		return nil, err
	}

	if session.cached {
		return output, nil
	}

	awsRequest, awsOutput := ec2.New(session.session).DescribeSecurityGroupsRequest(nil)
	awsRequest.Params = input

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error describing security groups")
		return nil, err
	}

	opsee_aws.CopyInto(output, awsOutput)

	err = s.db.Put(store.Request{
		CustomerId: req.User.CustomerId,
		Input:      input,
		Output:     output,
	})

	if err != nil {
		session.log.WithError(err).Error("error saving to cache")
		// continue on
	}

	return output, nil
}

func (s *service) EC2DescribeSubnets(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_ec2.DescribeSubnetsOutput, error) {
	var (
		input  = req.GetEc2_DescribeSubnetsInput()
		output = &opsee_aws_ec2.DescribeSubnetsOutput{}
	)

	session, err := s.requestSession(ctx, req, input, output)
	if err != nil {
		return nil, err
	}

	if session.cached {
		return output, nil
	}

	awsRequest, awsOutput := ec2.New(session.session).DescribeSubnetsRequest(nil)
	awsRequest.Params = input

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error describing subnets")
		return nil, err
	}

	opsee_aws.CopyInto(output, awsOutput)

	err = s.db.Put(store.Request{
		CustomerId: req.User.CustomerId,
		Input:      input,
		Output:     output,
	})

	if err != nil {
		session.log.WithError(err).Error("error saving to cache")
		// continue on
	}

	return output, nil
}

func (s *service) EC2DescribeVpcs(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_ec2.DescribeVpcsOutput, error) {
	var (
		input  = req.GetEc2_DescribeVpcsInput()
		output = &opsee_aws_ec2.DescribeVpcsOutput{}
	)

	session, err := s.requestSession(ctx, req, input, output)
	if err != nil {
		return nil, err
	}

	if session.cached {
		return output, nil
	}

	awsRequest, awsOutput := ec2.New(session.session).DescribeVpcsRequest(nil)
	awsRequest.Params = input

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error describing vpcs")
		return nil, err
	}

	opsee_aws.CopyInto(output, awsOutput)

	err = s.db.Put(store.Request{
		CustomerId: req.User.CustomerId,
		Input:      input,
		Output:     output,
	})

	if err != nil {
		session.log.WithError(err).Error("error saving to cache")
		// continue on
	}

	return output, nil
}

func (s *service) EC2DescribeRouteTables(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_ec2.DescribeRouteTablesOutput, error) {
	var (
		input  = req.GetEc2_DescribeRouteTablesInput()
		output = &opsee_aws_ec2.DescribeRouteTablesOutput{}
	)

	session, err := s.requestSession(ctx, req, input, output)
	if err != nil {
		return nil, err
	}

	if session.cached {
		return output, nil
	}

	awsRequest, awsOutput := ec2.New(session.session).DescribeRouteTablesRequest(nil)
	awsRequest.Params = input

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error describing route tables")
		return nil, err
	}

	opsee_aws.CopyInto(output, awsOutput)

	err = s.db.Put(store.Request{
		CustomerId: req.User.CustomerId,
		Input:      input,
		Output:     output,
	})

	if err != nil {
		session.log.WithError(err).Error("error saving to cache")
		// continue on
	}

	return output, nil
}
