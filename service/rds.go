package service

import (
	"github.com/aws/aws-sdk-go/service/rds"
	opsee_aws "github.com/opsee/basic/schema/aws"
	opsee_aws_rds "github.com/opsee/basic/schema/aws/rds"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/bezosphere/store"
	"golang.org/x/net/context"
)

func (s *service) RDSDescribeDBInstances(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_rds.DescribeDBInstancesOutput, error) {
	var (
		input  = req.GetRds_DescribeDBInstancesInput()
		output = &opsee_aws_rds.DescribeDBInstancesOutput{}
	)

	session, err := s.requestSession(ctx, req, input, output)
	if err != nil {
		return nil, err
	}

	if session.cached {
		return output, nil
	}

	awsRequest, awsOutput := rds.New(session.session).DescribeDBInstancesRequest(nil)
	awsRequest.Params = input

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error describing db instances")
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
