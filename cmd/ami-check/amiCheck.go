package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
	"k8s.io/kops/pkg/apis/kops"
)

// runCheck executes the AMI availability validation flow.
func runCheck(cfg *CheckConfig, awsSession *session.Session) error {
	// Log start of check.
	log.Infoln("Running check.")

	// Fetch instance groups from the kops state store.
	instanceGroups, err := listKopsInstanceGroups(cfg, awsSession)
	if err != nil {
		return fmt.Errorf("failed to list kops instance groups: %w", err)
	}
	log.Infoln("Retrieved kops instance groups.")

	// Fetch available AMIs from EC2.
	images, err := listEC2Images(cfg, awsSession)
	if err != nil {
		return fmt.Errorf("failed to list AMIs: %w", err)
	}
	log.Infof("Retrieved AWS AMIs. (Total: %d)", len(images))

	// Check for missing AMIs and collect errors.
	errors := checkImagesAreAvailable(instanceGroups, images)
	if len(errors) != 0 {
		return fmt.Errorf("%s", strings.Join(errors, "; "))
	}

	log.Infoln("kops used images are available.")
	return nil
}

// checkImagesAreAvailable compares instance group images against available AMIs.
func checkImagesAreAvailable(instanceGroups []*kops.InstanceGroup, images []*ec2.Image) []string {
	// Prepare the error list.
	errorMessages := make([]string, 0)

	// Iterate each instance group.
	for _, group := range instanceGroups {
		// Skip nil groups defensively.
		if group == nil {
			continue
		}

		log.Infoln("Looking at instance group:", group.Name)

		// Extract the image name for matching.
		imageName, err := extractInstanceGroupImageName(group)
		if err != nil {
			errorMessages = append(errorMessages, err.Error())
			continue
		}

		// Check whether the AMI is present in the EC2 list.
		found := false
		for _, image := range images {
			if imageMatchesInstanceGroup(image, imageName) {
				found = true
				break
			}
		}

		// Record missing AMIs.
		if !found {
			message := fmt.Sprintf("could not find image matching %s", group.Spec.Image)
			errorMessages = append(errorMessages, message)
		}
	}

	return errorMessages
}

// extractInstanceGroupImageName trims the kops image reference to the AMI name.
func extractInstanceGroupImageName(group *kops.InstanceGroup) (string, error) {
	// Validate the image field.
	if group == nil {
		return "", fmt.Errorf("instance group was nil")
	}
	if len(group.Spec.Image) == 0 {
		return "", fmt.Errorf("instance group %s does not define an image", group.Name)
	}

	// The image is typically in owner/name format.
	parts := strings.Split(group.Spec.Image, "/")
	if len(parts) < 2 {
		return group.Spec.Image, nil
	}

	return parts[1], nil
}

// imageMatchesInstanceGroup checks whether an EC2 image matches the instance group image name.
func imageMatchesInstanceGroup(image *ec2.Image, imageName string) bool {
	// Guard against nil inputs.
	if image == nil {
		return false
	}
	if len(imageName) == 0 {
		return false
	}

	// Check the EC2 image name field.
	if image.Name != nil {
		if strings.Contains(strings.TrimSpace(*image.Name), strings.TrimSpace(imageName)) {
			log.Infoln("Found kops instance group image within list:", *image.Name)
			return true
		}
	}

	// Check the EC2 image location field.
	if image.ImageLocation != nil {
		if strings.Contains(strings.TrimSpace(*image.ImageLocation), strings.TrimSpace(imageName)) {
			log.Infoln("Found kops instance group image within list:", *image.ImageLocation)
			return true
		}
	}

	return false
}
