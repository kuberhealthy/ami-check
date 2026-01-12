package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const (
	// wellKnownAccountKopeio is the kops account ID.
	wellKnownAccountKopeio = "383156758163"
	// wellKnownAccountRedhat is the Red Hat account ID.
	wellKnownAccountRedhat = "309956199498"
	// wellKnownAccountCoreOS is the CoreOS account ID.
	wellKnownAccountCoreOS = "595879546273"
	// wellKnownAccountAmazonLinux2 is the Amazon Linux 2 account ID.
	wellKnownAccountAmazonLinux2 = "137112412989"
)

// listEC2Images queries EC2 for available AMIs from trusted owners.
func listEC2Images(cfg *CheckConfig, awsSession *session.Session) ([]*ec2.Image, error) {
	// Build the EC2 client for the configured region.
	ec2Client := ec2.New(awsSession, &aws.Config{Region: aws.String(cfg.AWSRegion)})

	// Assemble the trusted owner list.
	kopeioOwner := wellKnownAccountKopeio
	redHatOwner := wellKnownAccountRedhat
	coreOSOwner := wellKnownAccountCoreOS
	awsLinux2Owner := wellKnownAccountAmazonLinux2

	owners := []*string{
		&kopeioOwner,
		&redHatOwner,
		&coreOSOwner,
		&awsLinux2Owner,
	}

	// Request AMIs from the trusted owners.
	result, err := ec2Client.DescribeImages(&ec2.DescribeImagesInput{
		Owners: owners,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list EC2 images: %w", err)
	}

	return result.Images, nil
}
