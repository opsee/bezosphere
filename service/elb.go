package service

import (
	"github.com/aws/aws-sdk-go/service/elb"
	opsee_aws "github.com/opsee/basic/schema/aws"
	opsee_aws_elb "github.com/opsee/basic/schema/aws/elb"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/bezosphere/store"
	"golang.org/x/net/context"
)

func (s *service) ELBDescribeLoadBalancers(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_elb.DescribeLoadBalancersOutput, error) {
	var (
		input  = req.GetElb_DescribeLoadBalancersInput()
		output = &opsee_aws_elb.DescribeLoadBalancersOutput{}
	)

	session, err := s.requestSession(ctx, req, input, output)
	if err != nil {
		return nil, err
	}

	if session.cached {
		return output, nil
	}

	awsRequest, awsOutput := elb.New(session.session).DescribeLoadBalancersRequest(nil)
	awsRequest.Params = input

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error describing load balancers")
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
