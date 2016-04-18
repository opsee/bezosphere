package service

import (
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	opsee_aws "github.com/opsee/basic/schema/aws"
	opsee_aws_cloudwatch "github.com/opsee/basic/schema/aws/cloudwatch"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/bezosphere/store"
	"golang.org/x/net/context"
)

func (s *service) CloudwatchListMetrics(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_cloudwatch.ListMetricsOutput, error) {
	var (
		input  = req.GetCloudwatch_ListMetricsInput()
		output = &opsee_aws_cloudwatch.ListMetricsOutput{}
	)

	session, err := s.requestSession(ctx, req, input, output)
	if err != nil {
		return nil, err
	}

	if session.cached {
		return output, nil
	}

	awsRequest, awsOutput := cloudwatch.New(session.session).ListMetricsRequest(nil)
	awsRequest.Params = input

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error fetching cloudwatch metrics")
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

func (s *service) CloudwatchGetMetricStatistics(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_cloudwatch.GetMetricStatisticsOutput, error) {
	var (
		input  = req.GetCloudwatch_GetMetricStatisticsInput()
		output = &opsee_aws_cloudwatch.GetMetricStatisticsOutput{}
	)

	session, err := s.requestSession(ctx, req, input, output)
	if err != nil {
		return nil, err
	}

	if session.cached {
		return output, nil
	}

	awsRequest, awsOutput := cloudwatch.New(session.session).GetMetricStatisticsRequest(nil)
	awsRequest.Params = input

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error fetching cloudwatch metrics")
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
