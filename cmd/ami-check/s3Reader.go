package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"k8s.io/kops/pkg/apis/kops"
)

// listKopsInstanceGroups loads instance group data from the kops state store in S3.
func listKopsInstanceGroups(cfg *CheckConfig, awsSession *session.Session) ([]*kops.InstanceGroup, error) {
	// Log the retrieval intent.
	log.Infoln("Listing KOPS instance groups from AWS S3.")

	// Build the S3 client.
	awsS3 := s3.New(awsSession, &aws.Config{Region: aws.String(cfg.AWSRegion)})
	if awsS3 == nil {
		return nil, fmt.Errorf("nil S3 client")
	}

	// List object metadata from the bucket.
	objects, err := listS3Objects(cfg, awsS3)
	if err != nil {
		return nil, err
	}

	// Read and parse instance group objects.
	instanceGroups, err := readInstanceGroupObjects(cfg, awsS3, objects)
	if err != nil {
		return nil, err
	}

	log.Infoln("Found", len(instanceGroups), "instance groups.")
	return instanceGroups, nil
}

// listS3Objects lists objects in the configured S3 bucket.
func listS3Objects(cfg *CheckConfig, awsS3 *s3.S3) ([]*s3.Object, error) {
	// Track pagination and results.
	var marker string
	results := make([]*s3.Object, 0)
	log.Infoln("Querying object keys from S3 bucket.")

	// Fetch the first page of objects.
	objects, err := awsS3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(cfg.AWSS3BucketName),
	})
	if err != nil {
		log.Errorln("failed to list bucket objects:", err.Error())
		return results, err
	}
	results = append(results, objects.Contents...)

	// Iterate remaining pages using marker.
	for objects.Marker != nil {
		marker = *objects.Marker
		if len(marker) == 0 {
			break
		}

		log.Infoln("There are more bucket objects to be queried:", marker)

		objects, err = awsS3.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(cfg.AWSS3BucketName),
			Marker: aws.String(marker),
		})
		if err != nil {
			log.Errorln("failed to list bucket objects:", err.Error())
			return results, err
		}

		results = append(results, objects.Contents...)
	}

	log.Infoln("Found", len(results), "objects in this bucket.")
	return results, nil
}

// readInstanceGroupObjects loads instance group YAML from S3 and parses it.
func readInstanceGroupObjects(cfg *CheckConfig, awsS3 *s3.S3, objects []*s3.Object) ([]*kops.InstanceGroup, error) {
	// Prepare the result slice.
	results := make([]*kops.InstanceGroup, 0)
	log.Infoln("Reading S3 object contents.")

	// Compile the instance group key matcher.
	matcher, err := regexp.Compile(kopsStateStoreInstanceGroupKey)
	if err != nil {
		return results, fmt.Errorf("failed to compile instance group matcher: %w", err)
	}

	// Iterate each object.
	for _, object := range objects {
		// Skip objects without a key.
		if object == nil || object.Key == nil {
			continue
		}

		// Filter to instance group paths.
		if !matcher.MatchString(*object.Key) {
			log.Debugln("Skipping object with key:", *object.Key)
			continue
		}

		// Filter to the target cluster.
		if !strings.Contains(*object.Key, cfg.ClusterName) {
			log.Debugf("Skipping object due to mismatching cluster names. Object for %s, but looking for %s.", *object.Key, cfg.ClusterName)
			continue
		}

		log.Infoln("Information for object with key:", *object.Key)

		// Request the object from S3.
		output, err := awsS3.GetObject(&s3.GetObjectInput{
			Key:    aws.String(*object.Key),
			Bucket: aws.String(cfg.AWSS3BucketName),
		})
		if err != nil {
			log.Errorf("failed to fetch bucket object with key %s: %s", *object.Key, err.Error())
			return results, err
		}
		if output == nil || output.Body == nil {
			log.Errorf("object body was empty for key %s", *object.Key)
			continue
		}

		// Read the object body.
		objectBytes, err := io.ReadAll(output.Body)
		if err != nil {
			log.Errorf("failed to read object body: %s", err.Error())
			continue
		}

		// Parse YAML into instance group struct.
		var ig kops.InstanceGroup
		err = kops.ParseRawYaml(objectBytes, &ig)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal yaml data: %w", err)
			log.Errorln(err)
			continue
		}

		// Append the parsed instance group.
		log.Infoln("Found and unmarshalled data for:", ig.Name)
		results = append(results, &ig)
	}

	return results, nil
}
