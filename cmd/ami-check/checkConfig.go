package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	log "github.com/sirupsen/logrus"
)

const (
	// awsRegionPattern validates AWS region strings.
	awsRegionPattern = `^[\w]{2}[-][\w]{4,9}[-][\d]$`
	// kopsStateStoreInstanceGroupKey is used to filter S3 object keys.
	kopsStateStoreInstanceGroupKey = `/instancegroup/`

	// defaultAWSRegion is used when AWS_REGION is unset.
	defaultAWSRegion = "us-east-1"
	// defaultAWSS3BucketName is used when AWS_S3_BUCKET_NAME is unset.
	defaultAWSS3BucketName = "kops-state-store"
	// defaultClusterName is used when CLUSTER_FQDN is unset.
	defaultClusterName = "cluster-fqdn"

	// defaultCheckTimeLimit is the fallback time limit for the check run.
	defaultCheckTimeLimit = time.Minute * 1
)

// CheckConfig stores environment-driven configuration for the AMI check.
type CheckConfig struct {
	// AWSRegion selects the region for EC2 and S3 queries.
	AWSRegion string
	// AWSS3BucketName identifies the kops state store bucket.
	AWSS3BucketName string
	// ClusterName filters kops instance group objects in S3.
	ClusterName string
	// Debug enables verbose logging.
	Debug bool
	// CheckTimeLimit sets the allowed runtime for the check.
	CheckTimeLimit time.Duration
}

// parseConfig loads environment variables into a CheckConfig for the run.
func parseConfig() (*CheckConfig, error) {
	// Start with default values.
	cfg := &CheckConfig{}
	cfg.AWSRegion = defaultAWSRegion
	cfg.AWSS3BucketName = defaultAWSS3BucketName
	cfg.ClusterName = defaultClusterName
	cfg.CheckTimeLimit = defaultCheckTimeLimit

	// Parse debug settings first so logs are verbose when needed.
	debugEnv := os.Getenv("DEBUG")
	if len(debugEnv) != 0 {
		debugValue := parseDebugValue(debugEnv)
		cfg.Debug = debugValue
	}

	// Apply debug logging configuration.
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
		log.Infoln("Debug logging enabled.")
	}
	log.Debugln(os.Args)

	// Parse AWS_REGION.
	regionEnv := os.Getenv("AWS_REGION")
	if len(regionEnv) != 0 {
		valid, err := validateAWSRegion(regionEnv)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, fmt.Errorf("AWS_REGION does not match expected format")
		}
		cfg.AWSRegion = regionEnv
	}

	// Parse AWS_S3_BUCKET_NAME.
	bucketEnv := os.Getenv("AWS_S3_BUCKET_NAME")
	if len(bucketEnv) != 0 {
		cfg.AWSS3BucketName = bucketEnv
	}

	// Parse CLUSTER_FQDN.
	clusterEnv := os.Getenv("CLUSTER_FQDN")
	if len(clusterEnv) != 0 {
		cfg.ClusterName = clusterEnv
	}

	// Parse deadline from Kuberhealthy.
	deadline, err := checkclient.GetDeadline()
	if err == nil {
		cfg.CheckTimeLimit = deadline.Sub(time.Now().Add(time.Second * 5))
	}

	// Sync checkclient debug output to config.
	checkclient.Debug = cfg.Debug

	return cfg, nil
}

// validateAWSRegion confirms the AWS region format matches the expected pattern.
func validateAWSRegion(value string) (bool, error) {
	// Compile and evaluate the region regexp.
	ok, err := regexp.MatchString(awsRegionPattern, value)
	if err != nil {
		return false, fmt.Errorf("failed to parse AWS_REGION: %w", err)
	}

	return ok, nil
}

// parseDebugValue interprets DEBUG values without strconv to avoid multi-arch issues.
func parseDebugValue(value string) bool {
	// Normalize the input string.
	normalized := strings.ToLower(strings.TrimSpace(value))

	// Accept common truthy values.
	if normalized == "t" {
		return true
	}
	if normalized == "true" {
		return true
	}
	if normalized == "yes" {
		return true
	}

	return false
}
