package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var (
	regexPattern                               *regexp.Regexp
	regexCapturedGroupNotContainsPattern       *regexp.Regexp
	regexNonNamedCapturedGroupContainsPattern  *regexp.Regexp
	regexPatterns                              []*regexp.Regexp
	regexCapturedGroupNotContainsPatterns      []*regexp.Regexp
	regexNonNamedCapturedGroupContainsPatterns []*regexp.Regexp

	regexAllMatchInput                     string
	regexAllMatchData                      []string
	regexAllMatchMetadata                  *Metadata
	regexAllMatchMetadataSerialized        string
	regexContainsUnmatchInput              string
	regexContainsUnmatchData               []string
	regexContainsUnmatchMetadata           *Metadata
	regexContainsUnmatchMetadataSerialized string
	regexContainsSkipLines                 []int
	regexContainsSkipData                  []string
	regexContainsSkipMetadata              *Metadata
	regexContainsSkipMetadataSerialized    string
	regexAllUnmatchInput                   string
	regexAllUnmatchMetadata                *Metadata
	regexAllUnmatchMetadataSerialized      string
	regexAllSkipLines                      []int
	regexAllSkipMetadata                   *Metadata
	regexAllSkipMetadataSerialized         string
	regexEmptyMetadata                     *Metadata
	regexEmptyMetadataSerialized           string
	regexMixedSkipLines                    []int
	regexMixedData                         []string
	regexMixedMetadata                     *Metadata
	regexMixedMetadataSerialized           string

	ltsvAllMatchInput                     string
	ltsvAllMatchData                      []string
	ltsvAllMatchMetadata                  *Metadata
	ltsvAllMatchMetadataSerialized        string
	ltsvContainsUnmatchInput              string
	ltsvContainsUnmatchData               []string
	ltsvContainsUnmatchMetadata           *Metadata
	ltsvContainsUnmatchMetadataSerialized string
	ltsvContainsSkipLines                 []int
	ltsvContainsSkipData                  []string
	ltsvContainsSkipMetadata              *Metadata
	ltsvContainsSkipMetadataSerialized    string
	ltsvAllUnmatchInput                   string
	ltsvAllUnmatchMetadata                *Metadata
	ltsvAllUnmatchMetadataSerialized      string
	ltsvAllSkipLines                      []int
	ltsvAllSkipMetadata                   *Metadata
	ltsvAllSkipMetadataSerialized         string
	ltsvEmptyMetadata                     *Metadata
	ltsvEmptyMetadataSerialized           string
	ltsvMixedSkipLines                    []int
	ltsvMixedData                         []string
	ltsvMixedMetadata                     *Metadata
	ltsvMixedMetadataSerialized           string

	fileNotFoundMessage        string
	fileNotFoundMessageWindows string
	isDirMessage               string
	isDirMessageWindows        string
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	regexPattern = regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[ -~]+ [0-9+]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) "(?P<request_uri>[ -~]+)" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+)`)
	regexCapturedGroupNotContainsPattern = regexp.MustCompile("[!-~]+")
	regexNonNamedCapturedGroupContainsPattern = regexp.MustCompile("(?P<field1>[!-~]+) ([!-~]+) (?P<field3>[!-~]+)")
	regexPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[ -~]+ [0-9+]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) "(?P<request_uri>[ -~]+)" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+) (?P<acl_required>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[ -~]+ [0-9+]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) "(?P<request_uri>[ -~]+)" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[ -~]+ [0-9+]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) "(?P<request_uri>[ -~]+)" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[ -~]+ [0-9+]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) "(?P<request_uri>[ -~]+)" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[ -~]+ [0-9+]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) "(?P<request_uri>[ -~]+)" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+)`),
	}
	regexCapturedGroupNotContainsPatterns = append(append(regexCapturedGroupNotContainsPatterns, regexPatterns...), regexCapturedGroupNotContainsPattern)
	regexNonNamedCapturedGroupContainsPatterns = append(append(regexNonNamedCapturedGroupContainsPatterns, regexPatterns...), regexNonNamedCapturedGroupContainsPattern)

	regexAllMatchInput = `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" - s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket43.s3.us-west-1.amazonaws.com TLSV1.1 -
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-" "S3Console/0.4" - 9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com TLSV1.1
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 - "-" "S3Console/0.4" - BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4" - Ke1bUcazaN1jWuUlPJaxF64cQVpUEhoZKEG/hmy/gijN/I1DeWqDfFvnpybfEseEME/u7ME1234= SigV2 ECDHE-RSA-AES128-SHA AuthHeader
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" -`
	regexAllMatchData = []string{
		`{"index":1,"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket43?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`{"index":2,"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","request_uri":"GET /awsrandombucket59?logging HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`{"index":3,"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","request_uri":"GET /awsrandombucket12?policy HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"index":4,"bucket_owner":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","bucket":"awsrandombucket89","time":"[03/Feb/2019:03:54:33 +0000]","remote_ip":"192.0.2.76","requester":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","request_id":"7B4A0FABBEXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket89?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"33","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
		`{"index":5,"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket77?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexAllMatchMetadata = &Metadata{
		Total:     5,
		Matched:   5,
		Unmatched: 0,
		Skipped:   0,
		Source:    "",
		Errors:    nil,
	}
	regexAllMatchMetadataSerialized = `{"total":5,"matched":5,"unmatched":0,"skipped":0,"source":"%s","errors":null}`

	regexContainsUnmatchInput = `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" - s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket43.s3.us-west-1.amazonaws.com TLSV1.1 -
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-" "S3Console/0.4" - 9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com TLSV1.1
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 - "-" "S3Console/0.4" - BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4"
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" -`
	regexContainsUnmatchData = []string{
		`{"index":1,"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket43?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`{"index":2,"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","request_uri":"GET /awsrandombucket59?logging HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`{"index":3,"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","request_uri":"GET /awsrandombucket12?policy HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"index":5,"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket77?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexContainsUnmatchMetadata = &Metadata{
		Total:     5,
		Matched:   4,
		Unmatched: 1,
		Skipped:   0,
		Source:    "",
		Errors: []ErrorRecord{
			{
				Index:  4,
				Record: "d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\"",
			},
		},
	}
	regexContainsUnmatchMetadataSerialized = `{"total":5,"matched":4,"unmatched":1,"skipped":0,"source":"%s","errors":[{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\""}]}`

	regexContainsSkipLines = []int{2, 4}
	regexContainsSkipData = []string{
		`{"index":1,"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket43?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`{"index":3,"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","request_uri":"GET /awsrandombucket12?policy HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"index":5,"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket77?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexContainsSkipMetadata = &Metadata{
		Total:     5,
		Matched:   3,
		Unmatched: 0,
		Skipped:   2,
		Source:    "",
		Errors:    nil,
	}
	regexContainsSkipMetadataSerialized = `{"total":5,"matched":3,"unmatched":0,"skipped":2,"source":"%s","errors":null}`

	regexAllUnmatchInput = `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4"
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-"
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 -
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 -`
	regexAllUnmatchMetadata = &Metadata{
		Total:     5,
		Matched:   0,
		Unmatched: 5,
		Skipped:   0,
		Source:    "",
		Errors: []ErrorRecord{
			{
				Index:  1,
				Record: "a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket43?versioning HTTP/1.1\" 200 - 113 - 7 - \"-\" \"S3Console/0.4\"",
			},
			{
				Index:  2,
				Record: "3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - \"GET /awsrandombucket59?logging HTTP/1.1\" 200 - 242 - 11 - \"-\"",
			},
			{
				Index:  3,
				Record: "8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - \"GET /awsrandombucket12?policy HTTP/1.1\" 404 NoSuchBucketPolicy 297 - 38 -",
			},
			{
				Index:  4,
				Record: "d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33",
			},
			{
				Index:  5,
				Record: "01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket77?versioning HTTP/1.1\" 200 - 113 -",
			},
		},
	}
	regexAllUnmatchMetadataSerialized = `{"total":5,"matched":0,"unmatched":5,"skipped":0,"source":"%s","errors":[{"index":1,"record":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket43?versioning HTTP/1.1\" 200 - 113 - 7 - \"-\" \"S3Console/0.4\""},{"index":2,"record":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - \"GET /awsrandombucket59?logging HTTP/1.1\" 200 - 242 - 11 - \"-\""},{"index":3,"record":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - \"GET /awsrandombucket12?policy HTTP/1.1\" 404 NoSuchBucketPolicy 297 - 38 -"},{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33"},{"index":5,"record":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket77?versioning HTTP/1.1\" 200 - 113 -"}]}`

	regexAllSkipLines = []int{1, 2, 3, 4, 5}
	regexAllSkipMetadata = &Metadata{
		Total:     5,
		Matched:   0,
		Unmatched: 0,
		Skipped:   5,
		Source:    "",
		Errors:    nil,
	}
	regexAllSkipMetadataSerialized = `{"total":5,"matched":0,"unmatched":0,"skipped":5,"source":"%s","errors":null}`

	regexEmptyMetadata = &Metadata{
		Total:     0,
		Matched:   0,
		Unmatched: 0,
		Skipped:   0,
		Source:    "",
		Errors:    nil,
	}
	regexEmptyMetadataSerialized = `{"total":0,"matched":0,"unmatched":0,"skipped":0,"source":"%s","errors":null}`

	regexMixedSkipLines = []int{1}
	regexMixedData = []string{
		`{"index":2,"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","request_uri":"GET /awsrandombucket59?logging HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`{"index":3,"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","request_uri":"GET /awsrandombucket12?policy HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"index":5,"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket77?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexMixedMetadata = &Metadata{
		Total:     5,
		Matched:   3,
		Unmatched: 1,
		Skipped:   1,
		Source:    "",
		Errors: []ErrorRecord{
			{
				Index:  4,
				Record: "d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\"",
			},
		},
	}
	regexMixedMetadataSerialized = `{"total":5,"matched":3,"unmatched":1,"skipped":1,"source":"%s","errors":[{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\""}]}`

	ltsvAllMatchInput = `remote_host:192.168.1.1	remote_logname:-	remote_user:john	datetime:[12/Mar/2023:10:55:36 +0000]	request:GET /index.html HTTP/1.1	status:200	size:1024	referer:http://www.example.com/	user_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)
remote_host:172.16.0.2	remote_logname:-	remote_user:jane	datetime:[12/Mar/2023:10:56:10 +0000]	request:POST /login HTTP/1.1	status:303	size:532	referer:http://www.example.com/login	user_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)
remote_host:10.0.0.3	remote_logname:-	remote_user:mike	datetime:[12/Mar/2023:10:57:15 +0000]	request:GET /about.html HTTP/1.1	status:200	size:749	referer:http://www.example.com/	user_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)
remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	request:GET /products HTTP/1.1	status:404	size:0
remote_host:192.168.1.10	remote_logname:-	remote_user:chris	datetime:[12/Mar/2023:11:04:16 +0000]	request:DELETE /account HTTP/1.1	status:200	size:204`
	ltsvAllMatchData = []string{
		`{"index":1,"remote_host":"192.168.1.1","remote_logname":"-","remote_user":"john","datetime":"[12/Mar/2023:10:55:36 +0000]","request":"GET /index.html HTTP/1.1","status":"200","size":"1024","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}`,
		`{"index":2,"remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		`{"index":3,"remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		`{"index":4,"remote_host":"192.168.1.4","remote_logname":"-","remote_user":"anna","datetime":"[12/Mar/2023:10:58:24 +0000]","request":"GET /products HTTP/1.1","status":"404","size":"0"}`,
		`{"index":5,"remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvAllMatchMetadata = &Metadata{
		Total:     5,
		Matched:   5,
		Unmatched: 0,
		Skipped:   0,
		Source:    "",
		Errors:    nil,
	}
	ltsvAllMatchMetadataSerialized = `{"total":5,"matched":5,"unmatched":0,"skipped":0,"source":"%s","errors":null}`

	ltsvContainsUnmatchInput = `remote_host:192.168.1.1	remote_logname:-	remote_user:john	datetime:[12/Mar/2023:10:55:36 +0000]	request:GET /index.html HTTP/1.1	status:200	size:1024	referer:http://www.example.com/	user_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)
remote_host:172.16.0.2	remote_logname:-	remote_user:jane	datetime:[12/Mar/2023:10:56:10 +0000]	request:POST /login HTTP/1.1	status:303	size:532	referer:http://www.example.com/login	user_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)
remote_host:10.0.0.3	remote_logname:-	remote_user:mike	datetime:[12/Mar/2023:10:57:15 +0000]	request:GET /about.html HTTP/1.1	status:200	size:749	referer:http://www.example.com/	user_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)
remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	request:GET /products HTTP/1.1	404	size:0
remote_host:192.168.1.10	remote_logname:-	remote_user:chris	datetime:[12/Mar/2023:11:04:16 +0000]	request:DELETE /account HTTP/1.1	status:200	size:204`
	ltsvContainsUnmatchData = []string{
		`{"index":1,"remote_host":"192.168.1.1","remote_logname":"-","remote_user":"john","datetime":"[12/Mar/2023:10:55:36 +0000]","request":"GET /index.html HTTP/1.1","status":"200","size":"1024","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}`,
		`{"index":2,"remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		`{"index":3,"remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		`{"index":5,"remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvContainsUnmatchMetadata = &Metadata{
		Total:     5,
		Matched:   4,
		Unmatched: 1,
		Skipped:   0,
		Source:    "",
		Errors: []ErrorRecord{
			{
				Index:  4,
				Record: "remote_host:192.168.1.4\tremote_logname:-\tremote_user:anna\tdatetime:[12/Mar/2023:10:58:24 +0000]\trequest:GET /products HTTP/1.1\t404\tsize:0",
			},
		},
	}
	ltsvContainsUnmatchMetadataSerialized = `{"total":5,"matched":4,"unmatched":1,"skipped":0,"source":"%s","errors":[{"index":4,"record":"remote_host:192.168.1.4\tremote_logname:-\tremote_user:anna\tdatetime:[12/Mar/2023:10:58:24 +0000]\trequest:GET /products HTTP/1.1\t404\tsize:0"}]}`

	ltsvContainsSkipLines = []int{2, 4}
	ltsvContainsSkipData = []string{
		`{"index":1,"remote_host":"192.168.1.1","remote_logname":"-","remote_user":"john","datetime":"[12/Mar/2023:10:55:36 +0000]","request":"GET /index.html HTTP/1.1","status":"200","size":"1024","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}`,
		`{"index":3,"remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		`{"index":5,"remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvContainsSkipMetadata = &Metadata{
		Total:     5,
		Matched:   3,
		Unmatched: 0,
		Skipped:   2,
		Source:    "",
		Errors:    nil,
	}
	ltsvContainsSkipMetadataSerialized = `{"total":5,"matched":3,"unmatched":0,"skipped":2,"source":"%s","errors":null}`

	ltsvAllUnmatchInput = `192.168.1.1	remote_logname:-	remote_user:john	datetime:[12/Mar/2023:10:55:36 +0000]	request:GET /index.html HTTP/1.1	status:200	size:1024	referer:http://www.example.com/	user_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)
remote_host:172.16.0.2	-	remote_user:jane	datetime:[12/Mar/2023:10:56:10 +0000]	request:POST /login HTTP/1.1	status:303	size:532	referer:http://www.example.com/login	user_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)
remote_host:10.0.0.3	remote_logname:-	mike	datetime:[12/Mar/2023:10:57:15 +0000]	request:GET /about.html HTTP/1.1	status:200	size:749	referer:http://www.example.com/	user_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)
remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	GET /products HTTP/1.1	status:404	size:0
remote_host:192.168.1.10	remote_logname:-	remote_user:chris	datetime:[12/Mar/2023:11:04:16 +0000]	request:DELETE /account HTTP/1.1	200	size:204`
	ltsvAllUnmatchMetadata = &Metadata{
		Total:     5,
		Matched:   0,
		Unmatched: 5,
		Skipped:   0,
		Source:    "",
		Errors: []ErrorRecord{
			{
				Index:  1,
				Record: "192.168.1.1\tremote_logname:-\tremote_user:john\tdatetime:[12/Mar/2023:10:55:36 +0000]\trequest:GET /index.html HTTP/1.1\tstatus:200\tsize:1024\treferer:http://www.example.com/\tuser_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			},
			{
				Index:  2,
				Record: "remote_host:172.16.0.2\t-\tremote_user:jane\tdatetime:[12/Mar/2023:10:56:10 +0000]\trequest:POST /login HTTP/1.1\tstatus:303\tsize:532\treferer:http://www.example.com/login\tuser_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
			},
			{
				Index:  3,
				Record: "remote_host:10.0.0.3\tremote_logname:-\tmike\tdatetime:[12/Mar/2023:10:57:15 +0000]\trequest:GET /about.html HTTP/1.1\tstatus:200\tsize:749\treferer:http://www.example.com/\tuser_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
			},
			{
				Index:  4,
				Record: "remote_host:192.168.1.4\tremote_logname:-\tremote_user:anna\tdatetime:[12/Mar/2023:10:58:24 +0000]\tGET /products HTTP/1.1\tstatus:404\tsize:0",
			},
			{
				Index:  5,
				Record: "remote_host:192.168.1.10\tremote_logname:-\tremote_user:chris\tdatetime:[12/Mar/2023:11:04:16 +0000]\trequest:DELETE /account HTTP/1.1\t200\tsize:204",
			},
		},
	}
	ltsvAllUnmatchMetadataSerialized = `{"total":5,"matched":0,"unmatched":5,"skipped":0,"source":"%s","errors":[{"index":1,"record":"192.168.1.1\tremote_logname:-\tremote_user:john\tdatetime:[12/Mar/2023:10:55:36 +0000]\trequest:GET /index.html HTTP/1.1\tstatus:200\tsize:1024\treferer:http://www.example.com/\tuser_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)"},{"index":2,"record":"remote_host:172.16.0.2\t-\tremote_user:jane\tdatetime:[12/Mar/2023:10:56:10 +0000]\trequest:POST /login HTTP/1.1\tstatus:303\tsize:532\treferer:http://www.example.com/login\tuser_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"},{"index":3,"record":"remote_host:10.0.0.3\tremote_logname:-\tmike\tdatetime:[12/Mar/2023:10:57:15 +0000]\trequest:GET /about.html HTTP/1.1\tstatus:200\tsize:749\treferer:http://www.example.com/\tuser_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"},{"index":4,"record":"remote_host:192.168.1.4\tremote_logname:-\tremote_user:anna\tdatetime:[12/Mar/2023:10:58:24 +0000]\tGET /products HTTP/1.1\tstatus:404\tsize:0"},{"index":5,"record":"remote_host:192.168.1.10\tremote_logname:-\tremote_user:chris\tdatetime:[12/Mar/2023:11:04:16 +0000]\trequest:DELETE /account HTTP/1.1\t200\tsize:204"}]}`

	ltsvAllSkipLines = []int{1, 2, 3, 4, 5}
	ltsvAllSkipMetadata = &Metadata{
		Total:     5,
		Matched:   0,
		Unmatched: 0,
		Skipped:   5,
		Source:    "",
		Errors:    nil,
	}
	ltsvAllSkipMetadataSerialized = `{"total":5,"matched":0,"unmatched":0,"skipped":5,"source":"%s","errors":null}`

	ltsvEmptyMetadata = &Metadata{
		Total:     0,
		Matched:   0,
		Unmatched: 0,
		Skipped:   0,
		Source:    "",
		Errors:    nil,
	}
	ltsvEmptyMetadataSerialized = `{"total":0,"matched":0,"unmatched":0,"skipped":0,"source":"%s","errors":null}`

	ltsvMixedSkipLines = []int{1}
	ltsvMixedData = []string{
		`{"index":2,"remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		`{"index":3,"remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		`{"index":5,"remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvMixedMetadata = &Metadata{
		Total:     5,
		Matched:   3,
		Unmatched: 1,
		Skipped:   1,
		Source:    "",
		Errors: []ErrorRecord{
			{
				Index:  4,
				Record: "remote_host:192.168.1.4\tremote_logname:-\tremote_user:anna\tdatetime:[12/Mar/2023:10:58:24 +0000]\trequest:GET /products HTTP/1.1\t404\tsize:0",
			},
		},
	}
	ltsvMixedMetadataSerialized = `{"total":5,"matched":3,"unmatched":1,"skipped":1,"source":"%s","errors":[{"index":4,"record":"remote_host:192.168.1.4\tremote_logname:-\tremote_user:anna\tdatetime:[12/Mar/2023:10:58:24 +0000]\trequest:GET /products HTTP/1.1\t404\tsize:0"}]}`

	fileNotFoundMessage = "no such file or directory"
	fileNotFoundMessageWindows = "The system cannot find the file specified."
	isDirMessage = "is a directory"
	isDirMessageWindows = "Incorrect function."
}

func Test_parse(t *testing.T) {
	type args struct {
		input           io.Reader
		skipLines       []int
		parser          parser
		patterns        []*regexp.Regexp
		lineHandler     LineHandler
		metadataHandler MetadataHandler
	}
	tests := []struct {
		name    string
		args    args
		want    *Result
		wantErr bool
	}{
		{
			name: "regexParser: all match",
			args: args{
				input:           strings.NewReader(regexAllMatchInput),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexAllMatchData,
				Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: all match",
			args: args{
				input:           strings.NewReader(ltsvAllMatchInput),
				skipLines:       nil,
				parser:          ltsvParser,
				patterns:        nil,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvAllMatchData,
				Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parse(tt.args.input, tt.args.skipLines, tt.args.parser, tt.args.patterns, tt.args.lineHandler, tt.args.metadataHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseString(t *testing.T) {
	type args struct {
		input           string
		skipLines       []int
		parser          parser
		patterns        []*regexp.Regexp
		lineHandler     LineHandler
		metadataHandler MetadataHandler
	}
	tests := []struct {
		name    string
		args    args
		want    *Result
		wantErr bool
	}{
		{
			name: "regexParser: all match",
			args: args{
				input:           regexAllMatchInput,
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexAllMatchData,
				Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "regexParser: contains unmatch",
			args: args{
				input:           regexContainsUnmatchInput,
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexContainsUnmatchData,
				Metadata: fmt.Sprintf(regexContainsUnmatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "regexParser: contains skip flag",
			args: args{
				input:           regexAllMatchInput,
				skipLines:       regexContainsSkipLines,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexContainsSkipData,
				Metadata: fmt.Sprintf(regexContainsSkipMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "regexParser: all unmatch",
			args: args{
				input:           regexAllUnmatchInput,
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(regexAllUnmatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "regexParser: all skip",
			args: args{
				input:           regexAllMatchInput,
				skipLines:       regexAllSkipLines,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(regexAllSkipMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "regexParser: mixed",
			args: args{
				input:           regexContainsUnmatchInput,
				skipLines:       regexMixedSkipLines,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexMixedData,
				Metadata: fmt.Sprintf(regexMixedMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "regexParser: nil input",
			args: args{
				input:           "",
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(regexEmptyMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "regexParser: line handler returns error",
			args: args{
				input:     regexAllMatchInput,
				skipLines: nil,
				parser:    regexParser,
				patterns:  regexPatterns,
				lineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: metadata handler returns error",
			args: args{
				input:       regexAllMatchInput,
				skipLines:   nil,
				parser:      regexParser,
				patterns:    regexPatterns,
				lineHandler: JSONLineHandler,
				metadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: all match",
			args: args{
				input:           ltsvAllMatchInput,
				skipLines:       nil,
				parser:          ltsvParser,
				patterns:        nil,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvAllMatchData,
				Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: contains unmatch",
			args: args{
				input:           ltsvContainsUnmatchInput,
				skipLines:       nil,
				parser:          ltsvParser,
				patterns:        nil,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvContainsUnmatchData,
				Metadata: fmt.Sprintf(ltsvContainsUnmatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: contains skip flag",
			args: args{
				input:           ltsvAllMatchInput,
				skipLines:       ltsvContainsSkipLines,
				parser:          ltsvParser,
				patterns:        nil,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvContainsSkipData,
				Metadata: fmt.Sprintf(ltsvContainsSkipMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: all unmatch",
			args: args{
				input:           ltsvAllUnmatchInput,
				skipLines:       nil,
				parser:          ltsvParser,
				patterns:        nil,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(ltsvAllUnmatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: all skip",
			args: args{
				input:           ltsvAllMatchInput,
				skipLines:       ltsvAllSkipLines,
				parser:          ltsvParser,
				patterns:        nil,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(ltsvAllSkipMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: mixed",
			args: args{
				input:           ltsvContainsUnmatchInput,
				skipLines:       ltsvMixedSkipLines,
				parser:          ltsvParser,
				patterns:        nil,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvMixedData,
				Metadata: fmt.Sprintf(ltsvMixedMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: nil input",
			args: args{
				input:           "",
				skipLines:       nil,
				parser:          ltsvParser,
				patterns:        nil,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(ltsvEmptyMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: line handler returns error",
			args: args{
				input:     ltsvAllMatchInput,
				skipLines: nil,
				parser:    ltsvParser,
				patterns:  nil,
				lineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: metadata handler returns error",
			args: args{
				input:       ltsvAllMatchInput,
				skipLines:   nil,
				parser:      ltsvParser,
				patterns:    nil,
				lineHandler: JSONLineHandler,
				metadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseString(tt.args.input, tt.args.skipLines, tt.args.parser, tt.args.patterns, tt.args.lineHandler, tt.args.metadataHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseFile(t *testing.T) {
	type args struct {
		input           string
		skipLines       []int
		parser          parser
		patterns        []*regexp.Regexp
		lineHandler     LineHandler
		metadataHandler MetadataHandler
	}
	tests := []struct {
		name    string
		args    args
		want    *Result
		wantErr bool
	}{
		{
			name: "regexParser: all match",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexAllMatchData,
				Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, "sample_s3_all_match.log"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: contains unmatch",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_contains_unmatch.log"),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexContainsUnmatchData,
				Metadata: fmt.Sprintf(regexContainsUnmatchMetadataSerialized, "sample_s3_contains_unmatch.log"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: contains skip flag",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines:       regexContainsSkipLines,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexContainsSkipData,
				Metadata: fmt.Sprintf(regexContainsSkipMetadataSerialized, "sample_s3_all_match.log"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: all unmatch",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_unmatch.log"),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(regexAllUnmatchMetadataSerialized, "sample_s3_all_unmatch.log"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: all skip",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines:       regexAllSkipLines,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(regexAllSkipMetadataSerialized, "sample_s3_all_match.log"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: mixed",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_contains_unmatch.log"),
				skipLines:       regexMixedSkipLines,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexMixedData,
				Metadata: fmt.Sprintf(regexMixedMetadataSerialized, "sample_s3_contains_unmatch.log"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: nil input",
			args: args{
				input:           "",
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: line handler returns error",
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines: nil,
				parser:    regexParser,
				patterns:  regexPatterns,
				lineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: metadata handler returns error",
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines:   nil,
				parser:      regexParser,
				patterns:    regexPatterns,
				lineHandler: JSONLineHandler,
				metadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: input file does not exists",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log.dummy"),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: input path is directory not file",
			args: args{
				input:           "testdata",
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: all match",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log"),
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvAllMatchData,
				Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, "sample_ltsv_all_match.log"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: contains unmatch",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_contains_unmatch.log"),
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvContainsUnmatchData,
				Metadata: fmt.Sprintf(ltsvContainsUnmatchMetadataSerialized, "sample_ltsv_contains_unmatch.log"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: contains skip flag",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log"),
				skipLines:       regexContainsSkipLines,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvContainsSkipData,
				Metadata: fmt.Sprintf(ltsvContainsSkipMetadataSerialized, "sample_ltsv_all_match.log"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: all unmatch",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_unmatch.log"),
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(ltsvAllUnmatchMetadataSerialized, "sample_ltsv_all_unmatch.log"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: all skip",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log"),
				skipLines:       regexAllSkipLines,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(ltsvAllSkipMetadataSerialized, "sample_ltsv_all_match.log"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: mixed",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_contains_unmatch.log"),
				skipLines:       ltsvMixedSkipLines,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvMixedData,
				Metadata: fmt.Sprintf(ltsvMixedMetadataSerialized, "sample_ltsv_contains_unmatch.log"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: nil input",
			args: args{
				input:           "",
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: line handler returns error",
			args: args{
				input:     filepath.Join("testdata", "sample_ltsv_all_match.log"),
				skipLines: nil,
				parser:    ltsvParser,
				lineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: metadata handler returns error",
			args: args{
				input:       filepath.Join("testdata", "sample_ltsv_all_match.log"),
				skipLines:   nil,
				parser:      ltsvParser,
				lineHandler: JSONLineHandler,
				metadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: input file does not exists",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log.dummy"),
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: input path is directory not file",
			args: args{
				input:           "testdata",
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFile(tt.args.input, tt.args.skipLines, tt.args.parser, tt.args.patterns, tt.args.lineHandler, tt.args.metadataHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseGzip(t *testing.T) {
	type args struct {
		input           string
		skipLines       []int
		parser          parser
		patterns        []*regexp.Regexp
		lineHandler     LineHandler
		metadataHandler MetadataHandler
	}
	tests := []struct {
		name    string
		args    args
		want    *Result
		wantErr bool
	}{
		{
			name: "regexParser: all match",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexAllMatchData,
				Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, "sample_s3_all_match.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: contains unmatch",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_contains_unmatch.log.gz"),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexContainsUnmatchData,
				Metadata: fmt.Sprintf(regexContainsUnmatchMetadataSerialized, "sample_s3_contains_unmatch.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: contains skip flag",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines:       regexContainsSkipLines,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexContainsSkipData,
				Metadata: fmt.Sprintf(regexContainsSkipMetadataSerialized, "sample_s3_all_match.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: all unmatch",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_unmatch.log.gz"),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(regexAllUnmatchMetadataSerialized, "sample_s3_all_unmatch.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: all skip",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines:       regexAllSkipLines,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(regexAllSkipMetadataSerialized, "sample_s3_all_match.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: mixed",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_contains_unmatch.log.gz"),
				skipLines:       regexMixedSkipLines,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexMixedData,
				Metadata: fmt.Sprintf(regexMixedMetadataSerialized, "sample_s3_contains_unmatch.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "regexParser: nil input",
			args: args{
				input:           "",
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: line handler returns error",
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines: nil,
				parser:    regexParser,
				patterns:  regexPatterns,
				lineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: metadata handler returns error",
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines:   nil,
				parser:      regexParser,
				patterns:    regexPatterns,
				lineHandler: JSONLineHandler,
				metadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: input file does not exists",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log.gz.dummy"),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: input path is directory not file",
			args: args{
				input:           "testdata",
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: input file is not gzip",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: all match",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvAllMatchData,
				Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, "sample_ltsv_all_match.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: contains unmatch",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_contains_unmatch.log.gz"),
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvContainsUnmatchData,
				Metadata: fmt.Sprintf(ltsvContainsUnmatchMetadataSerialized, "sample_ltsv_contains_unmatch.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: contains skip flag",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
				skipLines:       ltsvContainsSkipLines,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvContainsSkipData,
				Metadata: fmt.Sprintf(ltsvContainsSkipMetadataSerialized, "sample_ltsv_all_match.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: all unmatch",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_unmatch.log.gz"),
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(ltsvAllUnmatchMetadataSerialized, "sample_ltsv_all_unmatch.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: all skip",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
				skipLines:       ltsvAllSkipLines,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     nil,
				Metadata: fmt.Sprintf(ltsvAllSkipMetadataSerialized, "sample_ltsv_all_match.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: mixed",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_contains_unmatch.log.gz"),
				skipLines:       ltsvMixedSkipLines,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: &Result{
				Data:     ltsvMixedData,
				Metadata: fmt.Sprintf(ltsvMixedMetadataSerialized, "sample_ltsv_contains_unmatch.log.gz"),
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: nil input",
			args: args{
				input:           "",
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: line handler returns error",
			args: args{
				input:     filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
				skipLines: nil,
				parser:    ltsvParser,
				lineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: metadata handler returns error",
			args: args{
				input:       filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
				skipLines:   nil,
				parser:      ltsvParser,
				lineHandler: JSONLineHandler,
				metadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: input file does not exists",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log.gz.dummy"),
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: input path is directory not file",
			args: args{
				input:           "testdata",
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: input file is not gzip",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log"),
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGzip(tt.args.input, tt.args.skipLines, tt.args.parser, tt.args.patterns, tt.args.lineHandler, tt.args.metadataHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseGzip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseGzip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseZipEntries(t *testing.T) {
	type args struct {
		input           string
		skipLines       []int
		globPattern     string
		parser          parser
		patterns        []*regexp.Regexp
		lineHandler     LineHandler
		metadataHandler MetadataHandler
	}
	tests := []struct {
		name    string
		args    args
		want    []*Result
		wantErr bool
	}{
		{
			name: "regexParser: all match/one entry",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:       nil,
				globPattern:     "*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     regexAllMatchData,
					Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, "sample_s3_all_match.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "regexParser: contains unmatch/one entry",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_contains_unmatch.log.zip"),
				skipLines:       nil,
				globPattern:     "*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     regexContainsUnmatchData,
					Metadata: fmt.Sprintf(regexContainsUnmatchMetadataSerialized, "sample_s3_contains_unmatch.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "regexParser: contains skip flag/one entry",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:       regexAllSkipLines,
				globPattern:     "*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     nil,
					Metadata: fmt.Sprintf(regexAllSkipMetadataSerialized, "sample_s3_all_match.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "regexParser: all unmatch/one entry",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_unmatch.log.zip"),
				skipLines:       nil,
				globPattern:     "*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     nil,
					Metadata: fmt.Sprintf(regexAllUnmatchMetadataSerialized, "sample_s3_all_unmatch.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "regexParser: all skip/one entry",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:       regexAllSkipLines,
				globPattern:     "*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     nil,
					Metadata: fmt.Sprintf(regexAllSkipMetadataSerialized, "sample_s3_all_match.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "regexParser: mixed",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_contains_unmatch.log.zip"),
				skipLines:       regexMixedSkipLines,
				globPattern:     "*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     regexMixedData,
					Metadata: fmt.Sprintf(regexMixedMetadataSerialized, "sample_s3_contains_unmatch.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "regexParser: nil input",
			args: args{
				input:           "",
				skipLines:       nil,
				globPattern:     "*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: line handler returns error",
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:   nil,
				globPattern: "*",
				parser:      regexParser,
				patterns:    regexPatterns,
				lineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: metadata handler returns error",
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:   nil,
				globPattern: "*",
				parser:      regexParser,
				patterns:    regexPatterns,
				lineHandler: JSONLineHandler,
				metadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: input file does not exists",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log.zip.dummy"),
				skipLines:       nil,
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: input path is directory not file",
			args: args{
				input:           "testdata",
				skipLines:       nil,
				globPattern:     "*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: input file is not zip",
			args: args{
				input:           filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines:       nil,
				globPattern:     "*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "regexParser: multi entries",
			args: args{
				input:           filepath.Join("testdata", "sample_s3.zip"),
				skipLines:       nil,
				globPattern:     "*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     regexAllMatchData,
					Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, "sample_s3_all_match.log"),
				},
				{
					Data:     regexContainsUnmatchData,
					Metadata: fmt.Sprintf(regexContainsUnmatchMetadataSerialized, "sample_s3_contains_unmatch.log"),
				},
				{
					Data:     nil,
					Metadata: fmt.Sprintf(regexAllUnmatchMetadataSerialized, "sample_s3_all_unmatch.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "regexParser: multi entries and glob pattern filtering",
			args: args{
				input:           filepath.Join("testdata", "sample_s3.zip"),
				skipLines:       nil,
				globPattern:     "*all_match*",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     regexAllMatchData,
					Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, "sample_s3_all_match.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "regexParser: multi entries and glob pattern returns error",
			args: args{
				input:           filepath.Join("testdata", "log.zip"),
				skipLines:       nil,
				globPattern:     "[",
				parser:          regexParser,
				patterns:        regexPatterns,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: all match/one entry",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log.zip"),
				skipLines:       nil,
				globPattern:     "*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     ltsvAllMatchData,
					Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, "sample_ltsv_all_match.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: contains unmatch/one entry",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_contains_unmatch.log.zip"),
				skipLines:       nil,
				globPattern:     "*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     ltsvContainsUnmatchData,
					Metadata: fmt.Sprintf(ltsvContainsUnmatchMetadataSerialized, "sample_ltsv_contains_unmatch.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: contains skip flag/one entry",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log.zip"),
				skipLines:       ltsvAllSkipLines,
				globPattern:     "*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     nil,
					Metadata: fmt.Sprintf(ltsvAllSkipMetadataSerialized, "sample_ltsv_all_match.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: all unmatch/one entry",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_unmatch.log.zip"),
				skipLines:       nil,
				globPattern:     "*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     nil,
					Metadata: fmt.Sprintf(ltsvAllUnmatchMetadataSerialized, "sample_ltsv_all_unmatch.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: all skip/one entry",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log.zip"),
				skipLines:       ltsvAllSkipLines,
				globPattern:     "*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     nil,
					Metadata: fmt.Sprintf(ltsvAllSkipMetadataSerialized, "sample_ltsv_all_match.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: mixed",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_contains_unmatch.log.zip"),
				skipLines:       ltsvMixedSkipLines,
				globPattern:     "*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     ltsvMixedData,
					Metadata: fmt.Sprintf(ltsvMixedMetadataSerialized, "sample_ltsv_contains_unmatch.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: nil input",
			args: args{
				input:           "",
				skipLines:       nil,
				globPattern:     "*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: line handler returns error",
			args: args{
				input:       filepath.Join("testdata", "sample_ltsv_all_match.log.zip"),
				skipLines:   nil,
				globPattern: "*",
				parser:      ltsvParser,
				lineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: metadata handler returns error",
			args: args{
				input:       filepath.Join("testdata", "sample_ltsv_all_match.log.zip"),
				skipLines:   nil,
				globPattern: "*",
				parser:      ltsvParser,
				lineHandler: JSONLineHandler,
				metadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: input file does not exists",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log.zip.dummy"),
				skipLines:       nil,
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: input path is directory not file",
			args: args{
				input:           "testdata",
				skipLines:       nil,
				globPattern:     "*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: input file is not zip",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
				skipLines:       nil,
				globPattern:     "*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "ltsvParser: multi entries",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv.zip"),
				skipLines:       nil,
				globPattern:     "*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     ltsvAllMatchData,
					Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, "sample_ltsv_all_match.log"),
				},
				{
					Data:     ltsvContainsUnmatchData,
					Metadata: fmt.Sprintf(ltsvContainsUnmatchMetadataSerialized, "sample_ltsv_contains_unmatch.log"),
				},
				{
					Data:     nil,
					Metadata: fmt.Sprintf(ltsvAllUnmatchMetadataSerialized, "sample_ltsv_all_unmatch.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: multi entries and glob pattern filtering",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv.zip"),
				skipLines:       nil,
				globPattern:     "*all_match*",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want: []*Result{
				{
					Data:     ltsvAllMatchData,
					Metadata: fmt.Sprintf(ltsvAllMatchMetadataSerialized, "sample_ltsv_all_match.log"),
				},
			},
			wantErr: false,
		},
		{
			name: "ltsvParser: multi entries and glob pattern returns error",
			args: args{
				input:           filepath.Join("testdata", "sample_ltsv.zip"),
				skipLines:       nil,
				globPattern:     "[",
				parser:          ltsvParser,
				lineHandler:     JSONLineHandler,
				metadataHandler: JSONMetadataHandler,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseZipEntries(tt.args.input, tt.args.skipLines, tt.args.globPattern, tt.args.parser, tt.args.patterns, tt.args.lineHandler, tt.args.metadataHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseZipEntries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseZipEntries() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_regexParser(t *testing.T) {
	type args struct {
		input     io.Reader
		skipLines []int
		patterns  []*regexp.Regexp
		handler   LineHandler
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		want1   *Metadata
		wantErr bool
	}{
		{
			name: "all match",
			args: args{
				input:     strings.NewReader(regexAllMatchInput),
				skipLines: nil,
				patterns:  regexPatterns,
				handler:   JSONLineHandler,
			},
			want:    regexAllMatchData,
			want1:   regexAllMatchMetadata,
			wantErr: false,
		},
		{
			name: "contains unmatch",
			args: args{
				input:     strings.NewReader(regexContainsUnmatchInput),
				skipLines: nil,
				patterns:  regexPatterns,
				handler:   JSONLineHandler,
			},
			want:    regexContainsUnmatchData,
			want1:   regexContainsUnmatchMetadata,
			wantErr: false,
		},
		{
			name: "contains skip flag",
			args: args{
				input:     strings.NewReader(regexAllMatchInput),
				skipLines: regexContainsSkipLines,
				patterns:  regexPatterns,
				handler:   JSONLineHandler,
			},
			want:    regexContainsSkipData,
			want1:   regexContainsSkipMetadata,
			wantErr: false,
		},
		{
			name: "all unmatch",
			args: args{
				input:     strings.NewReader(regexAllUnmatchInput),
				skipLines: nil,
				patterns:  regexPatterns,
				handler:   JSONLineHandler,
			},
			want:    nil,
			want1:   regexAllUnmatchMetadata,
			wantErr: false,
		},
		{
			name: "all skip",
			args: args{
				input:     strings.NewReader(regexAllMatchInput),
				skipLines: regexAllSkipLines,
				patterns:  regexPatterns,
				handler:   JSONLineHandler,
			},
			want:    nil,
			want1:   regexAllSkipMetadata,
			wantErr: false,
		},
		{
			name: "mixed",
			args: args{
				input:     strings.NewReader(regexContainsUnmatchInput),
				skipLines: regexMixedSkipLines,
				patterns:  regexPatterns,
				handler:   JSONLineHandler,
			},
			want:    regexMixedData,
			want1:   regexMixedMetadata,
			wantErr: false,
		},
		{
			name: "nil input",
			args: args{
				input:     strings.NewReader(""),
				skipLines: nil,
				patterns:  regexPatterns,
				handler:   JSONLineHandler,
			},
			want:    nil,
			want1:   regexEmptyMetadata,
			wantErr: false,
		},
		{
			name: "line handler returns error",
			args: args{
				input:     strings.NewReader(regexAllMatchInput),
				skipLines: nil,
				patterns:  regexPatterns,
				handler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
		{
			name: "nil pattern",
			args: args{
				input:     strings.NewReader(regexAllMatchInput),
				skipLines: nil,
				patterns:  nil,
				handler:   JSONLineHandler,
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := regexParser(tt.args.input, tt.args.skipLines, tt.args.patterns, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("regexParser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("regexParser() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("regexParser() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_ltsvParser(t *testing.T) {
	type args struct {
		input     io.Reader
		skipLines []int
		handler   LineHandler
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		want1   *Metadata
		wantErr bool
	}{
		{
			name: "all match",
			args: args{
				input:     strings.NewReader(ltsvAllMatchInput),
				skipLines: nil,
				handler:   JSONLineHandler,
			},
			want:    ltsvAllMatchData,
			want1:   ltsvAllMatchMetadata,
			wantErr: false,
		},
		{
			name: "contains unmatch",
			args: args{
				input:     strings.NewReader(ltsvContainsUnmatchInput),
				skipLines: nil,
				handler:   JSONLineHandler,
			},
			want:    ltsvContainsUnmatchData,
			want1:   ltsvContainsUnmatchMetadata,
			wantErr: false,
		},
		{
			name: "contains skip flag",
			args: args{
				input:     strings.NewReader(ltsvAllMatchInput),
				skipLines: ltsvContainsSkipLines,
				handler:   JSONLineHandler,
			},
			want:    ltsvContainsSkipData,
			want1:   ltsvContainsSkipMetadata,
			wantErr: false,
		},
		{
			name: "all unmatch",
			args: args{
				input:     strings.NewReader(ltsvAllUnmatchInput),
				skipLines: nil,
				handler:   JSONLineHandler,
			},
			want:    nil,
			want1:   ltsvAllUnmatchMetadata,
			wantErr: false,
		},
		{
			name: "all skip",
			args: args{
				input:     strings.NewReader(ltsvAllMatchInput),
				skipLines: ltsvAllSkipLines,
				handler:   JSONLineHandler,
			},
			want:    nil,
			want1:   ltsvAllSkipMetadata,
			wantErr: false,
		},
		{
			name: "mixed",
			args: args{
				input:     strings.NewReader(ltsvContainsUnmatchInput),
				skipLines: ltsvMixedSkipLines,
				handler:   JSONLineHandler,
			},
			want:    ltsvMixedData,
			want1:   ltsvMixedMetadata,
			wantErr: false,
		},
		{
			name: "nil input",
			args: args{
				input:     strings.NewReader(""),
				skipLines: nil,
				handler:   JSONLineHandler,
			},
			want:    nil,
			want1:   ltsvEmptyMetadata,
			wantErr: false,
		},
		{
			name: "line handler returns error",
			args: args{
				input:     strings.NewReader(ltsvAllMatchInput),
				skipLines: nil,
				handler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			want1:   nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ltsvParser(tt.args.input, tt.args.skipLines, nil, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("ltsvParser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ltsvParser() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ltsvParser() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_createResult(t *testing.T) {
	type args struct {
		data     []string
		metadata *Metadata
		handler  MetadataHandler
	}
	tests := []struct {
		name    string
		args    args
		want    *Result
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				data:     regexAllMatchData,
				metadata: regexAllMatchMetadata,
				handler:  JSONMetadataHandler,
			},
			want: &Result{
				Data:     regexAllMatchData,
				Metadata: fmt.Sprintf(regexAllMatchMetadataSerialized, ""),
			},
			wantErr: false,
		},
		{
			name: "metadata handler returns error",
			args: args{
				data:     regexAllMatchData,
				metadata: regexAllMatchMetadata,
				handler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createResult(tt.args.data, tt.args.metadata, tt.args.handler)
			if (err != nil) != tt.wantErr {
				t.Errorf("createResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
