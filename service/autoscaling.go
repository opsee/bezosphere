package service

import (
	"github.com/aws/aws-sdk-go/service/autoscaling"
	opsee_aws "github.com/opsee/basic/schema/aws"
	opsee_aws_autoscaling "github.com/opsee/basic/schema/aws/autoscaling"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/bezosphere/store"
	"golang.org/x/net/context"
)

func (s *service) AutoScalingDescribeAutoScalingGroups(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_autoscaling.DescribeAutoScalingGroupsOutput, error) {
	var (
		input  = req.GetAutoscaling_DescribeAutoScalingGroupsInput()
		output = &opsee_aws_autoscaling.DescribeAutoScalingGroupsOutput{}
	)

	session, err := s.requestSession(ctx, req, input, output)
	if err != nil {
		return nil, err
	}

	if session.cached {
		return output, nil
	}

	awsRequest, awsOutput := autoscaling.New(session.session).DescribeAutoScalingGroupsRequest(nil)
	awsRequest.Params = input

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error fetching autoscaling groups")
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
