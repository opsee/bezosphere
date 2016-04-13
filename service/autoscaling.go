package service

import (
	opsee_aws_autoscaling "github.com/opsee/basic/schema/aws/autoscaling"
	opsee "github.com/opsee/basic/service"
	"golang.org/x/net/context"
)

func (s *service) AutoScalingDescribeAutoScalingGroups(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_autoscaling.DescribeAutoScalingGroupsOutput, error) {
	_, err := s.requestSession(ctx, req, "AutoScalingDescribeAutoScalingGroups")
	if err != nil {
		return nil, err
	}

	return nil, nil
}
