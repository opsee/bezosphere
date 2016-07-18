package service

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/rds"
	opsee_aws "github.com/opsee/basic/schema/aws"
	opsee_aws_autoscaling "github.com/opsee/basic/schema/aws/autoscaling"
	opsee_aws_cloudwatch "github.com/opsee/basic/schema/aws/cloudwatch"
	opsee_aws_ec2 "github.com/opsee/basic/schema/aws/ec2"
	opsee_aws_ecs "github.com/opsee/basic/schema/aws/ecs"
	opsee_aws_elb "github.com/opsee/basic/schema/aws/elb"
	opsee_aws_rds "github.com/opsee/basic/schema/aws/rds"
	opsee "github.com/opsee/basic/service"
	"github.com/opsee/bezosphere/store"
	log "github.com/opsee/logrus"
	"github.com/opsee/spanx/spanxcreds"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	grpcauth "google.golang.org/grpc/credentials"
)

var (
	ErrNoInput            = errors.New("request requires input, but none was given.")
	ErrNoUser             = errors.New("request requires a user, but none was given.")
	ErrNoRegion           = errors.New("request requires a region, but none was given.")
	ErrNoVpcId            = errors.New("request requires a vpc id, but none was given.")
	ErrInvalidUser        = errors.New("user is invalid.")
	ErrInvalidCredentials = errors.New("invalid AWS credentials.")
)

type service struct {
	spanxClient opsee.SpanxClient
	db          store.Store
}

type Config struct {
	SpanxAddress string
	Db           store.Store
}

func New(config Config) (*service, error) {
	svc := &service{
		db: config.Db,
	}

	spanxconn, err := grpc.Dial(
		config.SpanxAddress,
		grpc.WithTransportCredentials(grpcauth.NewTLS(&tls.Config{})),
	)

	if err != nil {
		return nil, err
	}

	svc.spanxClient = opsee.NewSpanxClient(spanxconn)

	return svc, nil
}

func (s *service) Start(listenAddr, cert, certkey string) error {
	auth, err := grpcauth.NewServerTLSFromFile(cert, certkey)
	if err != nil {
		return err
	}

	server := grpc.NewServer(grpc.Creds(auth))
	opsee.RegisterBezosServer(server, s)

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}

	log.Infof("starting grpc server at %s", listenAddr)
	return server.Serve(lis)
}

func (s *service) Get(ctx context.Context, req *opsee.BezosRequest) (*opsee.BezosResponse, error) {
	if req.Input == nil {
		log.WithError(ErrNoInput).Errorf("invalid input %#v", req.Input)
		return nil, ErrNoInput
	}

	logger := log.WithField("input", reflect.TypeOf(req.Input).Elem().Name())

	if req.User == nil {
		logger.WithError(ErrNoUser).Error(ErrNoUser.Error())
		return nil, ErrNoUser
	}

	if err := req.User.Validate(); err != nil {
		logger.WithError(err).Error(ErrInvalidUser.Error())
		return nil, ErrInvalidUser
	}

	logger = logger.WithFields(log.Fields{
		"customer_id": req.User.CustomerId,
		"user_id":     req.User.Id,
	})

	if req.Region == "" {
		logger.WithError(ErrNoRegion).Error(ErrNoRegion.Error())
		return nil, ErrNoRegion
	}

	if req.VpcId == "" {
		logger.WithError(ErrNoVpcId).Error(ErrNoVpcId.Error())
		return nil, ErrNoVpcId
	}

	logger.Debug("valid grpc request")
	bites, err := json.Marshal(req.Input)
	if err != nil {
		logger.WithError(err).Error("can't marshal request input")
		return nil, err
	}
	logger.Debug("received request: ", string(bites))

	input, output, err := inputOutput(req.Input)
	if err != nil {
		logger.WithError(err).Error("error finding output")
		return nil, err
	}

	if shouldSkipCache(req.Input) {
		err = errors.New("input type not cached")
	} else {
		err = s.db.Get(store.Request{
			CustomerId: req.User.CustomerId,
			Input:      input,
			Output:     output,
			MaxAge:     req.MaxAge,
		})
	}

	var response *opsee.BezosResponse

	if err != nil {
		logger.WithError(err).Error("cache miss")
	} else {
		logger.Debug("cache hit")

		response, err = buildResponse(output)
		if err != nil {
			logger.WithError(err).Error("no response found")
			return nil, err
		}

		return response, nil
	}

	session := session.New(&aws.Config{
		Region:      aws.String(req.Region),
		Credentials: spanxcreds.NewSpanxCredentials(req.User, s.spanxClient),
	})

	err = dispatchRequest(ctx, logger, session, input, output)
	if err != nil {
		return nil, err
	}

	err = s.db.Put(store.Request{
		CustomerId: req.User.CustomerId,
		Input:      input,
		Output:     output,
	})

	if err != nil {
		logger.WithError(err).Error("error saving to cache")
		// just continue on
	}

	response, err = buildResponse(output)
	if err != nil {
		logger.WithError(err).Error("no response found")
		return nil, err
	}

	return response, nil
}

func shouldSkipCache(input interface{}) bool {
	switch input.(type) {
	case *opsee.BezosRequest_Cloudwatch_GetMetricStatisticsInput:
		return true
	}
	return false
}

func dispatchRequest(ctx context.Context, logger *log.Entry, session *session.Session, input interface{}, output interface{}) error {
	var (
		err       error
		awsOutput interface{}
	)

	switch input.(type) {
	case *opsee_aws_cloudwatch.ListMetricsInput:
		ipt := &cloudwatch.ListMetricsInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = cloudwatch.New(session).ListMetrics(ipt)

	case *opsee_aws_cloudwatch.GetMetricStatisticsInput:
		ipt := &cloudwatch.GetMetricStatisticsInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = cloudwatch.New(session).GetMetricStatistics(ipt)

	case *opsee_aws_ec2.DescribeInstancesInput:
		ipt := &ec2.DescribeInstancesInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ec2.New(session).DescribeInstances(ipt)

	case *opsee_aws_ec2.DescribeSecurityGroupsInput:
		ipt := &ec2.DescribeSecurityGroupsInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ec2.New(session).DescribeSecurityGroups(ipt)

	case *opsee_aws_ec2.DescribeSubnetsInput:
		ipt := &ec2.DescribeSubnetsInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ec2.New(session).DescribeSubnets(ipt)

	case *opsee_aws_ec2.DescribeVpcsInput:
		ipt := &ec2.DescribeVpcsInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ec2.New(session).DescribeVpcs(ipt)

	case *opsee_aws_ec2.DescribeRouteTablesInput:
		ipt := &ec2.DescribeRouteTablesInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ec2.New(session).DescribeRouteTables(ipt)

	case *opsee_aws_elb.DescribeLoadBalancersInput:
		ipt := &elb.DescribeLoadBalancersInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = elb.New(session).DescribeLoadBalancers(ipt)

	case *opsee_aws_autoscaling.DescribeAutoScalingGroupsInput:
		ipt := &autoscaling.DescribeAutoScalingGroupsInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = autoscaling.New(session).DescribeAutoScalingGroups(ipt)

	case *opsee_aws_rds.DescribeDBInstancesInput:
		ipt := &rds.DescribeDBInstancesInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = rds.New(session).DescribeDBInstances(ipt)

	case *opsee_aws_ecs.ListTasksInput:
		ipt := &ecs.ListTasksInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ecs.New(session).ListTasks(ipt)

	case *opsee_aws_ecs.DescribeTasksInput:
		ipt := &ecs.DescribeTasksInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ecs.New(session).DescribeTasks(ipt)

	case *opsee_aws_ecs.DescribeContainerInstancesInput:
		ipt := &ecs.DescribeContainerInstancesInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ecs.New(session).DescribeContainerInstances(ipt)

	case *opsee_aws_ecs.ListContainerInstancesInput:
		ipt := &ecs.ListContainerInstancesInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ecs.New(session).ListContainerInstances(ipt)

	case *opsee_aws_ecs.ListClustersInput:
		ipt := &ecs.ListClustersInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ecs.New(session).ListClusters(ipt)

	case *opsee_aws_ecs.ListServicesInput:
		ipt := &ecs.ListServicesInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ecs.New(session).ListServices(ipt)

	case *opsee_aws_ecs.DescribeServicesInput:
		ipt := &ecs.DescribeServicesInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ecs.New(session).DescribeServices(ipt)

	case *opsee_aws_ecs.DescribeTaskDefinitionInput:
		ipt := &ecs.DescribeTaskDefinitionInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = ecs.New(session).DescribeTaskDefinition(ipt)

	case *opsee_aws_cloudwatch.DescribeAlarmsInput:
		ipt := &cloudwatch.DescribeAlarmsInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = cloudwatch.New(session).DescribeAlarms(ipt)

	case *opsee_aws_cloudwatch.DescribeAlarmsForMetricInput:
		ipt := &cloudwatch.DescribeAlarmsForMetricInput{}
		opsee_aws.CopyInto(ipt, input)
		awsOutput, err = cloudwatch.New(session).DescribeAlarmsForMetric(ipt)

	default:
		return fmt.Errorf("input type not found: %#v", input)
	}

	if err != nil {
		logger.WithError(err).Error("aws request error")
		return err
	}

	opsee_aws.CopyInto(output, awsOutput)
	return nil
}

func inputOutput(ipt interface{}) (interface{}, interface{}, error) {
	var (
		input  interface{}
		output interface{}
	)

	switch t := ipt.(type) {
	case *opsee.BezosRequest_Ecs_ListTasksInput:
		input = t.Ecs_ListTasksInput
		output = &opsee_aws_ecs.ListTasksOutput{}

	case *opsee.BezosRequest_Ecs_DescribeTasksInput:
		input = t.Ecs_DescribeTasksInput
		output = &opsee_aws_ecs.DescribeTasksOutput{}

	case *opsee.BezosRequest_Ecs_DescribeContainerInstancesInput:
		input = t.Ecs_DescribeContainerInstancesInput
		output = &opsee_aws_ecs.DescribeContainerInstancesOutput{}

	case *opsee.BezosRequest_Ecs_ListContainerInstancesInput:
		input = t.Ecs_ListContainerInstancesInput
		output = &opsee_aws_ecs.ListContainerInstancesOutput{}

	case *opsee.BezosRequest_Ecs_ListClustersInput:
		input = t.Ecs_ListClustersInput
		output = &opsee_aws_ecs.ListClustersOutput{}

	case *opsee.BezosRequest_Ecs_ListServicesInput:
		input = t.Ecs_ListServicesInput
		output = &opsee_aws_ecs.ListServicesOutput{}

	case *opsee.BezosRequest_Ecs_DescribeServicesInput:
		input = t.Ecs_DescribeServicesInput
		output = &opsee_aws_ecs.DescribeServicesOutput{}

	case *opsee.BezosRequest_Ecs_DescribeTaskDefinitionInput:
		input = t.Ecs_DescribeTaskDefinitionInput
		output = &opsee_aws_ecs.DescribeTaskDefinitionOutput{}

	case *opsee.BezosRequest_Cloudwatch_ListMetricsInput:
		input = t.Cloudwatch_ListMetricsInput
		output = &opsee_aws_cloudwatch.ListMetricsOutput{}

	case *opsee.BezosRequest_Cloudwatch_GetMetricStatisticsInput:
		input = t.Cloudwatch_GetMetricStatisticsInput
		output = &opsee_aws_cloudwatch.GetMetricStatisticsOutput{}

	case *opsee.BezosRequest_Ec2_DescribeInstancesInput:
		input = t.Ec2_DescribeInstancesInput
		output = &opsee_aws_ec2.DescribeInstancesOutput{}

	case *opsee.BezosRequest_Ec2_DescribeSecurityGroupsInput:
		input = t.Ec2_DescribeSecurityGroupsInput
		output = &opsee_aws_ec2.DescribeSecurityGroupsOutput{}

	case *opsee.BezosRequest_Ec2_DescribeSubnetsInput:
		input = t.Ec2_DescribeSubnetsInput
		output = &opsee_aws_ec2.DescribeSubnetsOutput{}

	case *opsee.BezosRequest_Ec2_DescribeVpcsInput:
		input = t.Ec2_DescribeVpcsInput
		output = &opsee_aws_ec2.DescribeVpcsOutput{}

	case *opsee.BezosRequest_Ec2_DescribeRouteTablesInput:
		input = t.Ec2_DescribeRouteTablesInput
		output = &opsee_aws_ec2.DescribeRouteTablesOutput{}

	case *opsee.BezosRequest_Elb_DescribeLoadBalancersInput:
		input = t.Elb_DescribeLoadBalancersInput
		output = &opsee_aws_elb.DescribeLoadBalancersOutput{}

	case *opsee.BezosRequest_Autoscaling_DescribeAutoScalingGroupsInput:
		input = t.Autoscaling_DescribeAutoScalingGroupsInput
		output = &opsee_aws_autoscaling.DescribeAutoScalingGroupsOutput{}

	case *opsee.BezosRequest_Rds_DescribeDBInstancesInput:
		input = t.Rds_DescribeDBInstancesInput
		output = &opsee_aws_rds.DescribeDBInstancesOutput{}

	case *opsee.BezosRequest_Cloudwatch_DescribeAlarmsInput:
		input = t.Cloudwatch_DescribeAlarmsInput
		output = &opsee_aws_cloudwatch.DescribeAlarmsOutput{}

	case *opsee.BezosRequest_Cloudwatch_DescribeAlarmsForMetricInput:
		input = t.Cloudwatch_DescribeAlarmsForMetricInput
		output = &opsee_aws_cloudwatch.DescribeAlarmsForMetricOutput{}

	default:
		return nil, nil, fmt.Errorf("input type not found: %#v", ipt)
	}

	return input, output, nil
}

func buildResponse(opt interface{}) (*opsee.BezosResponse, error) {
	var (
		response *opsee.BezosResponse
	)

	switch t := opt.(type) {
	case *opsee_aws_ecs.ListTasksOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ecs_ListTasksOutput{t}}

	case *opsee_aws_ecs.DescribeTasksOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ecs_DescribeTasksOutput{t}}

	case *opsee_aws_ecs.DescribeContainerInstancesOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ecs_DescribeContainerInstancesOutput{t}}

	case *opsee_aws_ecs.ListContainerInstancesOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ecs_ListContainerInstancesOutput{t}}

	case *opsee_aws_ecs.ListClustersOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ecs_ListClustersOutput{t}}

	case *opsee_aws_ecs.ListServicesOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ecs_ListServicesOutput{t}}

	case *opsee_aws_ecs.DescribeServicesOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ecs_DescribeServicesOutput{t}}

	case *opsee_aws_ecs.DescribeTaskDefinitionOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ecs_DescribeTaskDefinitionOutput{t}}

	case *opsee_aws_cloudwatch.ListMetricsOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Cloudwatch_ListMetricsOutput{t}}

	case *opsee_aws_cloudwatch.GetMetricStatisticsOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Cloudwatch_GetMetricStatisticsOutput{t}}

	case *opsee_aws_ec2.DescribeInstancesOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ec2_DescribeInstancesOutput{t}}

	case *opsee_aws_ec2.DescribeSecurityGroupsOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ec2_DescribeSecurityGroupsOutput{t}}

	case *opsee_aws_ec2.DescribeSubnetsOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ec2_DescribeSubnetsOutput{t}}

	case *opsee_aws_ec2.DescribeVpcsOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ec2_DescribeVpcsOutput{t}}

	case *opsee_aws_ec2.DescribeRouteTablesOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Ec2_DescribeRouteTablesOutput{t}}

	case *opsee_aws_elb.DescribeLoadBalancersOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Elb_DescribeLoadBalancersOutput{t}}

	case *opsee_aws_autoscaling.DescribeAutoScalingGroupsOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Autoscaling_DescribeAutoScalingGroupsOutput{t}}

	case *opsee_aws_rds.DescribeDBInstancesOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Rds_DescribeDBInstancesOutput{t}}

	case *opsee_aws_cloudwatch.DescribeAlarmsOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Cloudwatch_DescribeAlarmsOutput{t}}

	case *opsee_aws_cloudwatch.DescribeAlarmsForMetricOutput:
		response = &opsee.BezosResponse{Output: &opsee.BezosResponse_Cloudwatch_DescribeAlarmsForMetricOutput{t}}

	default:
		return nil, fmt.Errorf("output type not found: %#v", t)
	}

	return response, nil
}
