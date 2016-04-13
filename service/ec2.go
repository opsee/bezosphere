package service

import (
	opsee_aws_ec2 "github.com/opsee/basic/schema/aws/ec2"
	opsee "github.com/opsee/basic/service"
	"golang.org/x/net/context"
)

func (s *service) EC2DescribeInstances(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_ec2.DescribeInstancesOutput, error) {
	_, err := s.requestSession(ctx, req, "EC2DescribeInstances")
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *service) EC2DescribeSecurityGroups(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_ec2.DescribeSecurityGroupsOutput, error) {
	_, err := s.requestSession(ctx, req, "EC2DescribeSecurityGroups")
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *service) EC2DescribeSubnets(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_ec2.DescribeSubnetsOutput, error) {
	_, err := s.requestSession(ctx, req, "EC2DescribeSubnets")
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *service) EC2DescribeVpcs(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_ec2.DescribeVpcsOutput, error) {
	_, err := s.requestSession(ctx, req, "EC2DescribeVpcs")
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *service) EC2DescribeRouteTables(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_ec2.DescribeRouteTablesOutput, error) {
	_, err := s.requestSession(ctx, req, "EC2DescribeRouteTables")
	if err != nil {
		return nil, err
	}

	return nil, nil
}
