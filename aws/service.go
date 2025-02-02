package aws

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/codecommit"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/turbot/go-kit/helpers"
	"github.com/turbot/steampipe-plugin-sdk/v4/plugin"
)

func BackupService(ctx context.Context, d *plugin.QueryData) (*backup.Backup, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return backup.New(sess), nil
}

func CodeCommitService(ctx context.Context, d *plugin.QueryData) (*codecommit.CodeCommit, error) {
	sess, err := getSessionForQuerySupportedRegion(ctx, d, codecommit.EndpointsID)
	if err != nil {
		return nil, err
	}
	return codecommit.New(sess), nil
}

func CloudFrontService(ctx context.Context, d *plugin.QueryData) (*cloudfront.CloudFront, error) {
	sess, err := getSession(ctx, d, GetDefaultAwsRegion(d))
	if err != nil {
		return nil, err
	}
	return cloudfront.New(sess), nil
}

func CloudWatchLogsService(ctx context.Context, d *plugin.QueryData) (*cloudwatchlogs.CloudWatchLogs, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return cloudwatchlogs.New(sess), nil
}

func DynamoDbService(ctx context.Context, d *plugin.QueryData) (*dynamodb.DynamoDB, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return dynamodb.New(sess), nil
}

func Ec2Service(ctx context.Context, d *plugin.QueryData, region string) (*ec2.EC2, error) {
	sess, err := getSessionForRegion(ctx, d, region)
	if err != nil {
		return nil, err
	}
	return ec2.New(sess), nil
}

func Ec2RegionsService(ctx context.Context, d *plugin.QueryData, region string) (*ec2.EC2, error) {
	// We can query EC2 for the list of supported regions. But, if credentials
	// are insufficient this query will retry many times, so we create a special
	// client with a small number of retries to prevent hangs.
	// Note - This is not cached, but usually the result of using this service will be.
	sess, err := getSessionWithMaxRetries(ctx, d, region, 4, 25*time.Millisecond)
	if err != nil {
		return nil, err
	}
	svc := ec2.New(sess)
	return svc, nil
}

func EcsService(ctx context.Context, d *plugin.QueryData) (*ecs.ECS, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return ecs.New(sess), nil
}

func EksService(ctx context.Context, d *plugin.QueryData) (*eks.EKS, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return eks.New(sess), nil
}

func EmrService(ctx context.Context, d *plugin.QueryData) (*emr.EMR, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return emr.New(sess), nil
}

func GlobalAcceleratorService(ctx context.Context, d *plugin.QueryData) (*globalaccelerator.GlobalAccelerator, error) {
	// Global Accelerator is a global service that supports endpoints in multiple AWS Regions but you must specify
	// the us-west-2 (Oregon) Region to create or update accelerators.
	sess, err := getSession(ctx, d, "us-west-2")
	if err != nil {
		return nil, err
	}
	return globalaccelerator.New(sess), nil
}

func NetworkFirewallService(ctx context.Context, d *plugin.QueryData) (*networkfirewall.NetworkFirewall, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return networkfirewall.New(sess), nil
}

func PinpointService(ctx context.Context, d *plugin.QueryData) (*pinpoint.Pinpoint, error) {
	sess, err := getSessionForQuerySupportedRegion(ctx, d, pinpoint.EndpointsID)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, nil
	}
	return pinpoint.New(sess), nil
}

func Route53Service(ctx context.Context, d *plugin.QueryData) (*route53.Route53, error) {
	sess, err := getSession(ctx, d, GetDefaultAwsRegion(d))
	if err != nil {
		return nil, err
	}
	return route53.New(sess), nil
}

func SageMakerService(ctx context.Context, d *plugin.QueryData) (*sagemaker.SageMaker, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return sagemaker.New(sess), nil
}

func SecurityHubService(ctx context.Context, d *plugin.QueryData) (*securityhub.SecurityHub, error) {
	sess, err := getSessionForQuerySupportedRegion(ctx, d, securityhub.EndpointsID)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, nil
	}
	return securityhub.New(sess), nil
}

// ServiceQuotasService returns the service connection for AWS ServiceQuotas service
func ServiceQuotasService(ctx context.Context, d *plugin.QueryData) (*servicequotas.ServiceQuotas, error) {
	// have we already created and cached the service?
	serviceCacheKey := fmt.Sprintf("servicequotas-%s", "region")
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(*servicequotas.ServiceQuotas), nil
	}
	// so it was not in cache - create service
	sess, err := getSession(ctx, d, "")
	if err != nil {
		return nil, err
	}
	svc := servicequotas.New(sess)
	d.ConnectionManager.Cache.Set(serviceCacheKey, svc)
	return svc, nil
}

func ServiceQuotasRegionalService(ctx context.Context, d *plugin.QueryData) (*servicequotas.ServiceQuotas, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return servicequotas.New(sess), nil
}

func SESService(ctx context.Context, d *plugin.QueryData) (*ses.SES, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return ses.New(sess), nil
}

func SSOAdminService(ctx context.Context, d *plugin.QueryData) (*ssoadmin.SSOAdmin, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return ssoadmin.New(sess), nil
}

func StepFunctionsService(ctx context.Context, d *plugin.QueryData) (*sfn.SFN, error) {
	sess, err := getSessionForQueryRegion(ctx, d)
	if err != nil {
		return nil, err
	}
	return sfn.New(sess), nil
}

func StsService(ctx context.Context, d *plugin.QueryData) (*sts.STS, error) {
	// TODO - Should STS be regional instead?
	sess, err := getSession(ctx, d, GetDefaultAwsRegion(d))
	if err != nil {
		return nil, err
	}
	return sts.New(sess), nil
}

func getSession(ctx context.Context, d *plugin.QueryData, region string) (*session.Session, error) {

	sessionCacheKey := fmt.Sprintf("session-%s", region)
	if cachedData, ok := d.ConnectionManager.Cache.Get(sessionCacheKey); ok {
		return cachedData.(*session.Session), nil
	}

	awsConfig := GetConfig(d.Connection)

	// As per the logic used in retryRules of NewConnectionErrRetryer, default to minimum delay of 25ms and maximum
	// number of retries as 9 (our default). The default maximum delay will not be more than approximately 3 minutes to avoid Steampipe
	// waiting too long to return results
	maxRetries := 9
	var minRetryDelay time.Duration = 25 * time.Millisecond // Default minimum delay

	// Set max retry count from config file or env variable (config file has precedence)
	if awsConfig.MaxErrorRetryAttempts != nil {
		maxRetries = *awsConfig.MaxErrorRetryAttempts
	} else if os.Getenv("AWS_MAX_ATTEMPTS") != "" {
		maxRetriesEnvVar, err := strconv.Atoi(os.Getenv("AWS_MAX_ATTEMPTS"))
		if err != nil || maxRetriesEnvVar < 1 {
			panic("invalid value for environment variable \"AWS_MAX_ATTEMPTS\". It should be an integer value greater than or equal to 1")
		}
		maxRetries = maxRetriesEnvVar
	}

	// Set min delay time from config file
	if awsConfig.MinErrorRetryDelay != nil {
		minRetryDelay = time.Duration(*awsConfig.MinErrorRetryDelay) * time.Millisecond
	}

	if maxRetries < 1 {
		panic("\nconnection config has invalid value for \"max_error_retry_attempts\", it must be greater than or equal to 1. Edit your connection configuration file and then restart Steampipe.")
	}
	if minRetryDelay < 1 {
		panic("\nconnection config has invalid value for \"min_error_retry_delay\", it must be greater than or equal to 1. Edit your connection configuration file and then restart Steampipe.")
	}

	sess, err := getSessionWithMaxRetries(ctx, d, region, maxRetries, minRetryDelay)
	if err != nil {
		plugin.Logger(ctx).Error("getClient.getSessionWithMaxRetries", "region", region, "err", err)
	} else {
		// Caching sessions saves about 10ms, which is significant when there are
		// multiple instantiations (per account region) and when doing queries that
		// often take <100ms total. But, it's not that important compared to having
		// fresh credentials all the time. So, set a short cache length to ensure
		// we don't get tripped up by credential rotation on short lived roles etc.
		// The minimum assume role time is 15 minutes, so 5 minutes feels like a
		// reasonable balance - I certainly wouldn't do longer.
		d.ConnectionManager.Cache.SetWithTTL(sessionCacheKey, sess, 5*time.Minute)
	}

	return sess, err
}

func getSessionWithMaxRetries(ctx context.Context, d *plugin.QueryData, region string, maxRetries int, minRetryDelay time.Duration) (*session.Session, error) {

	// get aws config info
	awsConfig := GetConfig(d.Connection)

	// session default configuration
	sessionOptions := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region:     &region,
			MaxRetries: aws.Int(maxRetries),
			Retryer:    NewConnectionErrRetryer(maxRetries, minRetryDelay, ctx),
		},
	}

	// handle custom endpoint URL, if any
	var awsEndpointUrl string

	awsEndpointUrl = os.Getenv("AWS_ENDPOINT_URL")

	if awsConfig.EndpointUrl != nil {
		awsEndpointUrl = *awsConfig.EndpointUrl
	}

	if awsEndpointUrl != "" {
		sessionOptions.Config.Endpoint = aws.String(awsEndpointUrl)
	}

	if awsConfig.S3ForcePathStyle != nil {
		sessionOptions.Config.S3ForcePathStyle = awsConfig.S3ForcePathStyle
	}

	if awsConfig.Profile != nil {
		sessionOptions.Profile = *awsConfig.Profile
	}

	if awsConfig.AccessKey != nil && awsConfig.SecretKey == nil {
		return nil, fmt.Errorf("Partial credentials found in connection config, missing: secret_key")
	} else if awsConfig.SecretKey != nil && awsConfig.AccessKey == nil {
		return nil, fmt.Errorf("Partial credentials found in connection config, missing: access_key")
	} else if awsConfig.AccessKey != nil && awsConfig.SecretKey != nil {

		sessionOptions.Config.Credentials = credentials.NewStaticCredentials(
			*awsConfig.AccessKey, *awsConfig.SecretKey, "",
		)

		if awsConfig.SessionToken != nil {
			sessionOptions.Config.Credentials = credentials.NewStaticCredentials(
				*awsConfig.AccessKey, *awsConfig.SecretKey, *awsConfig.SessionToken,
			)
		}
	}

	sess, err := session.NewSessionWithOptions(sessionOptions)
	if err != nil {
		plugin.Logger(ctx).Error("getSessionWithMaxRetries.NewSessionWithOptions", "sessionOptions", sessionOptions, "err", err)
		return nil, err
	}

	return sess, nil
}

// Get a session for the region defined in query data, but only after checking it's
// a supported region for the given serviceID.
func getSessionForQuerySupportedRegion(ctx context.Context, d *plugin.QueryData, serviceID string) (*session.Session, error) {
	region := d.KeyColumnQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("getSessionForQueryRegion called without a region in QueryData")
	}
	validRegions, err := SupportedRegionsForService(ctx, d, serviceID)
	if err != nil {
		return nil, err
	}

	if !helpers.StringSliceContains(validRegions, region) {
		// We choose to ignore unsupported regions rather than returning an error
		// for them - it's a better user experience. So, return a nil session rather
		// than an error. The caller must handle this case.
		return nil, nil
	}
	// Supported region, so get and return the session
	return getSession(ctx, d, region)
}

// Helper function to get the session for a region set in query data
func getSessionForQueryRegion(ctx context.Context, d *plugin.QueryData) (*session.Session, error) {
	region := d.KeyColumnQualString(matrixKeyRegion)
	if region == "" {
		return nil, fmt.Errorf("getSessionForQueryRegion called without a region in QueryData")
	}
	return getSession(ctx, d, region)
}

// Helper function to get the session for a specific region
func getSessionForRegion(ctx context.Context, d *plugin.QueryData, region string) (*session.Session, error) {
	if region == "" {
		return nil, fmt.Errorf("getSessionForRegion called with an empty region")
	}
	return getSession(ctx, d, region)
}

// GetDefaultAwsRegion returns the default region for AWS partiton
// if not set by Env variable or in aws profile
func GetDefaultAwsRegion(d *plugin.QueryData) string {
	allAwsRegions := getAllAwsRegions()

	// have we already created and cached the service?
	serviceCacheKey := "GetDefaultAwsRegion"
	if cachedData, ok := d.ConnectionManager.Cache.Get(serviceCacheKey); ok {
		return cachedData.(string)
	}

	// get aws config info
	awsConfig := GetConfig(d.Connection)

	var regions []string
	var region string

	if awsConfig.Regions != nil {
		regions = awsConfig.Regions
		region = regions[0]
	} else {
		session, err := session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		})
		if err != nil {
			panic(err)
		}
		if session != nil && session.Config != nil {
			region = *session.Config.Region
		}

		if region != "" {
			regions = []string{region}
		}
	}

	invalidPatterns := []string{}
	for _, namePattern := range regions {
		validRegions := []string{}
		for _, validRegion := range allAwsRegions {
			if ok, _ := path.Match(namePattern, validRegion); ok {
				validRegions = append(validRegions, validRegion)
			}
		}

		// Region items with wildcards that match on 0 regions should not be
		// considered invalid
		if len(validRegions) == 0 && !strings.ContainsAny(namePattern, "?*") {
			invalidPatterns = append(invalidPatterns, namePattern)
		}
	}

	if len(invalidPatterns) > 0 {
		panic("\nconnection config has invalid \"regions\": " + strings.Join(invalidPatterns, ", ") + ". Edit your connection configuration file and then restart Steampipe.")
	}

	// most of the global services like IAM, S3, Route 53, etc. in all cloud types target these regions
	if strings.HasPrefix(region, "us-gov") && !helpers.StringSliceContains(allAwsRegions, region) {
		region = "us-gov-west-1"
	} else if strings.HasPrefix(region, "cn") && !helpers.StringSliceContains(allAwsRegions, region) {
		region = "cn-northwest-1"
	} else if strings.HasPrefix(region, "us-isob") && !helpers.StringSliceContains(allAwsRegions, region) {
		region = "us-isob-east-1"
	} else if strings.HasPrefix(region, "us-iso") && !helpers.StringSliceContains(allAwsRegions, region) {
		region = "us-iso-east-1"
	} else if !helpers.StringSliceContains(allAwsRegions, region) {
		region = "us-east-1"
	}

	d.ConnectionManager.Cache.Set(serviceCacheKey, region)
	return region
}

// Function from https://github.com/panther-labs/panther/blob/v1.16.0/pkg/awsretry/connection_retryer.go
func NewConnectionErrRetryer(maxRetries int, minRetryDelay time.Duration, ctx context.Context) *ConnectionErrRetryer {
	rand.Seed(time.Now().UnixNano()) // reseting state of rand to generate different random values
	return &ConnectionErrRetryer{
		ctx: ctx,
		DefaultRetryer: client.DefaultRetryer{
			NumMaxRetries: maxRetries,    // MUST be set or all retrying is skipped!
			MinRetryDelay: minRetryDelay, // Set minimum retry delay
		},
	}
}

// ConnectionErrRetryer wraps the SDK's built in DefaultRetryer adding customization
// to retry `connection reset by peer` errors.
// Note: This retryer should be used for either idempotent operations or for operations
// where performing duplicate requests to AWS is acceptable.
// See also: https://github.com/aws/aws-sdk-go/issues/3027#issuecomment-567269161
type ConnectionErrRetryer struct {
	client.DefaultRetryer
	ctx context.Context
}

func (r ConnectionErrRetryer) ShouldRetry(req *request.Request) bool {
	if req.Error != nil {
		if strings.Contains(req.Error.Error(), "connection reset by peer") {
			return true
		}

		var awsErr awserr.Error
		if errors.As(req.Error, &awsErr) {
			/*
				If no credentials are set or an invalid profile is provided, the AWS SDK
				will attempt to authenticate using all known methods. This takes a while
				since it will attempt to reach the EC2 metadata service and will continue
				to retry on connection errors, e.g.,
				awsErr.OrigErr()="Put "http://169.254.169.254/latest/api/token": context deadline exceeded (Client.Timeout exceeded while awaiting headers)
				awsErr.OrigErr()="Get "http://169.254.169.254/latest/meta-data/iam/security-credentials/": dial tcp 169.254.169.254:80: connect: no route to host"
				To reduce the time to fail, limit the number of retries for these errors specifically.
			*/
			if awsErr.OrigErr() != nil {
				if strings.Contains(awsErr.OrigErr().Error(), "http://169.254.169.254/latest") && req.RetryCount > 3 {
					return false
				}
			}
		}
	}

	// Fallback to SDK's built in retry rules
	return r.DefaultRetryer.ShouldRetry(req)
}

// Customize the RetryRules to implement exponential backoff retry
func (d ConnectionErrRetryer) RetryRules(r *request.Request) time.Duration {
	retryCount := r.RetryCount
	minDelay := d.MinRetryDelay

	// If errors are caused by load, retries can be ineffective if all API request retry at the same time.
	// To avoid this problem added a jitter of "+/-20%" with delay time.
	// For example, if the delay is 25ms, the final delay could be between 20 and 30ms.
	var jitter = float64(rand.Intn(120-80)+80) / 100

	// Creates a new exponential backoff using the starting value of
	// minDelay and (minDelay * 3^retrycount) * jitter on each failure
	// For example, with a min delay time of 25ms: 23.25ms, 63ms, 238.5ms, 607.4ms, 2s, 5.22s, 20.31s..., up to max.
	retryTime := time.Duration(int(float64(int(minDelay.Nanoseconds())*int(math.Pow(3, float64(retryCount)))) * jitter))

	// Cap retry time at 5 minutes to avoid too long a wait
	if retryTime > time.Duration(5*time.Minute) {
		retryTime = time.Duration(5 * time.Minute)
	}

	return retryTime
}
