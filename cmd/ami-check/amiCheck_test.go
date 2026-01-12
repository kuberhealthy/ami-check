package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/pkg/apis/kops"
)

// buildInstanceGroup constructs a minimal instance group for test cases.
func buildInstanceGroup(image string) *kops.InstanceGroup {
	// Assemble the instance group with an image reference.
	group := &kops.InstanceGroup{}
	group.Name = "ig-name"
	group.Spec.Image = image

	return group
}

// buildImage constructs an EC2 image for matching tests.
func buildImage(name string, location string) *ec2.Image {
	// Create the image and fill name/location fields as needed.
	image := &ec2.Image{}
	if len(name) != 0 {
		imageName := name
		image.Name = &imageName
	}
	if len(location) != 0 {
		imageLocation := location
		image.ImageLocation = &imageLocation
	}

	return image
}

// TestExtractInstanceGroupImageNameWithOwner ensures owner/name is trimmed to name.
func TestExtractInstanceGroupImageNameWithOwner(t *testing.T) {
	// Build a kops instance group with owner/name format.
	group := buildInstanceGroup("kope.io/k8s-1.27")

	// Extract the image name.
	name, err := extractInstanceGroupImageName(group)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate the trimmed name.
	if name != "k8s-1.27" {
		t.Fatalf("expected image name k8s-1.27, got %s", name)
	}
}

// TestExtractInstanceGroupImageNameWithoutOwner keeps a plain image name intact.
func TestExtractInstanceGroupImageNameWithoutOwner(t *testing.T) {
	// Build a kops instance group with a plain image name.
	group := buildInstanceGroup("k8s-1.27")

	// Extract the image name.
	name, err := extractInstanceGroupImageName(group)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Validate the unmodified name.
	if name != "k8s-1.27" {
		t.Fatalf("expected image name k8s-1.27, got %s", name)
	}
}

// TestExtractInstanceGroupImageNameEmpty ensures empty image references error.
func TestExtractInstanceGroupImageNameEmpty(t *testing.T) {
	// Build a kops instance group with an empty image.
	group := buildInstanceGroup("")

	// Extract the image name.
	name, err := extractInstanceGroupImageName(group)
	if err == nil {
		t.Fatalf("expected error, got name %s", name)
	}
}

// TestImageMatchesInstanceGroupName validates name-based matching.
func TestImageMatchesInstanceGroupName(t *testing.T) {
	// Build an EC2 image with a matching name.
	image := buildImage("k8s-1.27", "")

	// Match against the image name.
	matched := imageMatchesInstanceGroup(image, "k8s-1.27")
	if !matched {
		t.Fatalf("expected match for image name")
	}
}

// TestImageMatchesInstanceGroupLocation validates location-based matching.
func TestImageMatchesInstanceGroupLocation(t *testing.T) {
	// Build an EC2 image with a matching location string.
	image := buildImage("", "kope.io/k8s-1.27")

	// Match against the image location.
	matched := imageMatchesInstanceGroup(image, "k8s-1.27")
	if !matched {
		t.Fatalf("expected match for image location")
	}
}

// TestImageMatchesInstanceGroupNil ensures nil inputs do not match.
func TestImageMatchesInstanceGroupNil(t *testing.T) {
	// Attempt to match a nil image.
	matched := imageMatchesInstanceGroup(nil, "k8s-1.27")
	if matched {
		t.Fatalf("expected nil image to not match")
	}
}
