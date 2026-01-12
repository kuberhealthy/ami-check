package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
)

// createAWSSession builds a new AWS session for EC2 and S3 clients.
func createAWSSession() (*session.Session, error) {
	// Log the session creation for visibility.
	log.Infoln("Building AWS session.")

	// Build a session with verbose credential chain errors.
	cfg := aws.NewConfig()
	cfg = cfg.WithCredentialsChainVerboseErrors(true)

	awsSession, err := session.NewSession(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return awsSession, nil
}
