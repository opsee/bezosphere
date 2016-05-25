package main

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/opsee/basic/schema"
	opsee_aws_ec2 "github.com/opsee/basic/schema/aws/ec2"
	opsee_aws_cloudwatch "github.com/opsee/basic/schema/aws/cloudwatch"
	opsee "github.com/opsee/basic/service"
	opsee_types "github.com/opsee/protobuf/opseeproto/types"
)

var (
	DefaultResponseCacheTTL = time.Second * time.Duration(5)
)

type TestReq struct {
	VpcId     string
	Region    string
	User      *schema.User
	Instances []string
}

func main() {
	viper.SetEnvPrefix("bezos_test")
	viper.AutomaticEnv()

	bezosConn, err := grpc.Dial(
		viper.GetString("address"),
		grpc.WithTransportCredentials(
			credentials.NewTLS(&tls.Config{
				InsecureSkipVerify: true,
			}),
		),
	)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	bezosClient := opsee.NewBezosClient(bezosConn)

	testReq := &TestReq{
		VpcId:     "vpc-34610651",
		Region:    "us-west-1",
		Instances: []string{"i-d3b62266"},
		User: &schema.User{
			Id:         1,
			CustomerId: "f2e627a2-d108-11e5-a041-cfa352cc72b9",
			Email:      "mborsuk@gmail.com",
			Verified:   true,
			Active:     true,
		},
	}

	/*err = describeEC2(context.Background(), bezosClient, testReq)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}*/

	err = describeAlarms(context.Background(), bezosClient, testReq)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func describeEC2(ctx context.Context, client opsee.BezosClient, testReq *TestReq) error {
	input := &opsee_aws_ec2.DescribeInstancesInput{
		Filters: []*opsee_aws_ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []string{testReq.VpcId},
			},
		},
		InstanceIds: testReq.Instances,
	}

	timestamp := &opsee_types.Timestamp{}
	timestamp.Scan(time.Now().UTC().Add(DefaultResponseCacheTTL * -1))

	resp, err := client.Get(
		ctx,
		&opsee.BezosRequest{
			User:   testReq.User,
			Region: testReq.Region,
			VpcId:  testReq.VpcId,
			MaxAge: timestamp,
			Input:  &opsee.BezosRequest_Ec2_DescribeInstancesInput{input},
		})
	if err != nil {
		return err
	}

	output := resp.GetEc2_DescribeInstancesOutput()
	if output == nil {
		return fmt.Errorf("error decoding aws response")
	}

	//age := time.Now().Sub(time.Unix(resp.LastModified.Seconds, int64(resp.LastModified.Nanos)))
	//fmt.Printf("age: %ds\n", age.Seconds())
	for _, res := range output.Reservations {
		for _, instance := range res.Instances {
			fmt.Printf("%s:\n", *instance.InstanceId)
			for _, tag := range instance.Tags {
				fmt.Printf("   %s: %s\n", *tag.Key, *tag.Value)
			}
		}
	}

	return nil
}


func describeAlarms(ctx context.Context, client opsee.BezosClient, testReq *TestReq) error {
	input := &opsee_aws_cloudwatch.DescribeAlarmsInput{}

	timestamp := &opsee_types.Timestamp{}
	timestamp.Scan(time.Now().UTC().Add(DefaultResponseCacheTTL * -1))

	resp, err := client.Get(
		ctx,
		&opsee.BezosRequest{
			User:   testReq.User,
			Region: testReq.Region,
			VpcId:  testReq.VpcId,
			MaxAge: timestamp,
			Input:  &opsee.BezosRequest_Cloudwatch_DescribeAlarmsInput{input},
		})
	if err != nil {
		return err
	}

	output := resp.GetCloudwatch_DescribeAlarmsOutput()
	if output == nil {
		return fmt.Errorf("error decoding aws response")
	}

	for _, a := range output.MetricAlarms {
        fmt.Printf("%s: %s\n", *a.AlarmName, *a.StateValue)
	}

	return nil
}
