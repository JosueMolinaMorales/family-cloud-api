package aws

import (
	"context"
	"fmt"

	"github.com/JosueMolinaMorales/family-cloud-api/internal/config/log"
	"github.com/JosueMolinaMorales/family-cloud-api/pkg/error"
	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

type CognitoDriver interface {
	ValidateToken(token string) (bool, *error.RequestError)
	DoesUserExist(username string) (bool, *error.RequestError)
	GetCredentials(token string) (*types.Credentials, *error.RequestError)
}

func NewCognitoDriver(logger log.Logger) CognitoDriver {
	cfg, err := aws_config.LoadDefaultConfig(context.Background(), aws_config.WithRegion("us-east-1"), aws_config.WithSharedConfigProfile("personal"))
	if err != nil {
		panic(err)
	}

	return &cognitoDriver{
		cipClient: cognitoidentityprovider.NewFromConfig(cfg),
		ipClient:  cognitoidentity.NewFromConfig(cfg),
		logger:    logger,
	}
}

type cognitoDriver struct {
	cipClient *cognitoidentityprovider.Client
	ipClient  *cognitoidentity.Client
	logger    log.Logger
}

func (c *cognitoDriver) GetCredentials(token string) (*types.Credentials, *error.RequestError) {
	id, err := c.ipClient.GetId(context.TODO(), &cognitoidentity.GetIdInput{
		IdentityPoolId: aws.String("us-east-1:93a17f26-9955-4d0e-85ef-f63ca673ea17"),
		Logins: map[string]string{
			"cognito-idp.us-east-1.amazonaws.com/us-east-1_iS029qCx0": token,
		},
	}, func(o *cognitoidentity.Options) {})
	if err != nil {
		fmt.Println(err)
		return nil, error.NewRequestError(err, error.InternalServerError, "Error getting identity", c.logger)
	}
	fmt.Println(*id.IdentityId)
	creds, err := c.ipClient.GetCredentialsForIdentity(context.TODO(), &cognitoidentity.GetCredentialsForIdentityInput{
		IdentityId: id.IdentityId,
		Logins: map[string]string{
			"cognito-idp.us-east-1.amazonaws.com/us-east-1_iS029qCx0": token,
		},
	}, func(o *cognitoidentity.Options) {})
	if err != nil {
		fmt.Println(err)
		return nil, error.NewRequestError(err, error.InternalServerError, "Error getting credentials", c.logger)
	}

	return creds.Credentials, error.NewRequestError(nil, error.InternalServerError, "Error getting credentials", c.logger)
}

func (c *cognitoDriver) ValidateToken(token string) (bool, *error.RequestError) {
	// TODO
	return true, nil
}

func (c *cognitoDriver) DoesUserExist(username string) (bool, *error.RequestError) {
	res, err := c.cipClient.AdminGetUser(context.TODO(), &cognitoidentityprovider.AdminGetUserInput{
		Username:   &username,
		UserPoolId: aws.String(""),
	})
	if err != nil {
		return false, nil
	}

	if res.Username == nil {
		return false, nil
	}

	return true, nil
}
