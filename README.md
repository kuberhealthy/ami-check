# Kuberhealthy AMI Check

This check validates that every kops instance group image referenced in the kops state store still exists in AWS EC2. It is designed for Kuberhealthy v3 `HealthCheck` resources.

## What It Does

1. Lists kops instance group objects in the configured S3 state store.
2. Parses each instance group and extracts the image reference.
3. Queries EC2 for AMIs owned by well-known kops image publishers.
4. Fails when any instance group image cannot be found.

## Configuration

All configuration is controlled via environment variables.

- `AWS_REGION`: AWS region for EC2 and S3 queries. Default `us-east-1`.
- `AWS_S3_BUCKET_NAME`: kops state store S3 bucket. Default `kops-state-store`.
- `CLUSTER_FQDN`: Cluster FQDN used to filter instance group objects. Default `cluster-fqdn`.
- `DEBUG`: Enable debug logging.

Kuberhealthy injects these variables automatically into the check pod:

- `KH_REPORTING_URL`
- `KH_RUN_UUID`
- `KH_CHECK_RUN_DEADLINE`

## AWS Permissions

The check needs read-only permissions for the state store and AMI listings:

- `s3:ListBucket` on the kops state store bucket
- `s3:GetObject` on the kops state store bucket objects
- `ec2:DescribeImages`

## Build

Use the `Justfile` to build or test the check:

```bash
just build
just test
```

## Example HealthCheck

This example uses an IAM role annotation and custom S3 bucket settings:

```yaml
apiVersion: kuberhealthy.github.io/v2
kind: HealthCheck
metadata:
  name: ami
  namespace: kuberhealthy
spec:
  runInterval: 30m
  timeout: 10m
  extraAnnotations:
    iam.amazonaws.com/role: <role-arn>
  podSpec:
    spec:
      containers:
        - name: ami
          image: kuberhealthy/ami-check:sha-<short-sha>
          imagePullPolicy: IfNotPresent
          env:
            - name: AWS_REGION
              value: "us-east-1"
            - name: AWS_S3_BUCKET_NAME
              value: "s3-bucket-name"
            - name: CLUSTER_FQDN
              value: "cluster.k8s"
          resources:
            requests:
              cpu: 10m
              memory: 10Mi
            limits:
              cpu: 15m
      restartPolicy: Never
      serviceAccountName: ami-sa
```

A full install bundle with a ServiceAccount is available in `healthcheck.yaml`.

## Image Tags

- `sha-<short-sha>` tags are published on every push to `main`.
- `vX.Y.Z` tags are published when a matching Git tag is pushed.
