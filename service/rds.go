package service

import (
	opsee_aws_rds "github.com/opsee/basic/schema/aws/rds"
	opsee "github.com/opsee/basic/service"
	"golang.org/x/net/context"
)

func (s *service) RDSDescribeDBInstances(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_rds.DescribeDBInstancesOutput, error) {
	_, err := s.requestSession(ctx, req, "RDSDescribeDBInstances")
	if err != nil {
		return nil, err
	}

	return nil, nil
}
