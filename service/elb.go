package service

import (
	opsee_aws_elb "github.com/opsee/basic/schema/aws/elb"
	opsee "github.com/opsee/basic/service"
	"golang.org/x/net/context"
)

func (s *service) ELBDescribeLoadBalancers(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_elb.DescribeLoadBalancersOutput, error) {
	_, err := s.requestSession(ctx, req, "ELBDescribeLoadBalancers")
	if err != nil {
		return nil, err
	}

	return nil, nil
}
