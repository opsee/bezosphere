package service

import (
	"crypto/tls"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	opsee "github.com/opsee/basic/service"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	grpcauth "google.golang.org/grpc/credentials"
	"net"
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
}

type session struct {
	session *aws_session.Session
	log     *log.Entry
}

type Config struct {
	SpanxAddress string
}

func New(config Config) (*service, error) {
	svc := new(service)

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

func (s *service) requestSession(ctx context.Context, req *opsee.BezosRequest, endpoint string) (*session, error) {
	logger := log.WithField("endpoint", endpoint)

	if req.Input == nil {
		logger.WithError(ErrNoInput).Errorf("invalid input %#v", req.Input)
		return nil, ErrNoInput
	}

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

	logger.Info("valid grpc request")

	creds, err := s.spanxClient.GetCredentials(ctx, &opsee.GetCredentialsRequest{User: req.User})
	if err != nil {
		logger.WithError(err).Error(ErrInvalidCredentials.Error())
		return nil, ErrInvalidCredentials
	}

	sess := aws_session.New(&aws.Config{
		Region: aws.String(req.Region),
		Credentials: credentials.NewStaticCredentials(
			creds.Credentials.GetAccessKeyID(),
			creds.Credentials.GetSecretAccessKey(),
			creds.Credentials.GetSessionToken(),
		),
	})

	return &session{
		session: sess,
		log:     logger,
	}, nil
}
