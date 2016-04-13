package service

import (
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	opsee_aws_cloudwatch "github.com/opsee/basic/schema/aws/cloudwatch"
	opsee "github.com/opsee/basic/service"
	"golang.org/x/net/context"
)

func (s *service) CloudwatchListMetrics(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_cloudwatch.ListMetricsOutput, error) {
	session, err := s.requestSession(ctx, req, "CloudwatchListMetrics")
	if err != nil {
		return nil, err
	}

	input := req.GetCloudwatch_ListMetricsInput()
	if input == nil {
		session.log.WithError(ErrNoInput).Error(ErrNoInput.Error())
		return nil, ErrNoInput
	}
	output := &opsee_aws_cloudwatch.ListMetricsOutput{}

	client := cloudwatch.New(session.session)

	awsRequest, _ := client.ListMetricsRequest(nil)
	awsRequest.Params = input
	awsRequest.Data = output

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error listing metrics")
		return nil, err
	}

	return output, nil
}

func (s *service) CloudwatchGetMetricStatistics(ctx context.Context, req *opsee.BezosRequest) (*opsee_aws_cloudwatch.GetMetricStatisticsOutput, error) {
	session, err := s.requestSession(ctx, req, "CloudwatchGetMetricStatistics")
	if err != nil {
		return nil, err
	}

	input := req.GetCloudwatch_GetMetricStatisticsInput()
	if input == nil {
		session.log.WithError(ErrNoInput).Error(ErrNoInput.Error())
		return nil, ErrNoInput
	}

	output := &opsee_aws_cloudwatch.GetMetricStatisticsOutput{}

	client := cloudwatch.New(session.session)

	awsRequest, _ := client.GetMetricStatisticsRequest(nil)
	awsRequest.Params = input
	awsRequest.Data = output

	err = awsRequest.Send()
	if err != nil {
		session.log.WithError(err).Error("error getting metric statistics")
		return nil, err
	}

	return output, nil
}
