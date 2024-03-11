// Package parser provides utilities for parsing various types of logs (plain text, gzip, zip)
// and converting them into structured formats such as JSON or LTSV. It supports pattern matching,
// result extraction, and error handling.
package parser

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var (
	stringPattern                               string
	stringInvalidCapturePattern                 string
	stringCapturedGroupNotContainsPattern       string
	stringNonNamedCapturedGroupContainsPattern  string
	stringPatterns                              []string
	stringCapturedGroupNotContainsPatterns      []string
	stringNonNamedCapturedGroupContainsPatterns []string

	regexPattern                               *regexp.Regexp
	regexCapturedGroupNotContainsPattern       *regexp.Regexp
	regexNonNamedCapturedGroupContainsPattern  *regexp.Regexp
	regexPatterns                              []*regexp.Regexp
	regexCapturedGroupNotContainsPatterns      []*regexp.Regexp
	regexNonNamedCapturedGroupContainsPatterns []*regexp.Regexp

	regexAllMatchInput         string
	regexAllMatchData          []string
	regexAllMatchResult        *Result
	regexContainsUnmatchInput  string
	regexContainsUnmatchData   []string
	regexContainsUnmatchResult *Result
	regexContainsFilter        []string
	regexContainsFilterData    []string
	regexContainsFilterResult  *Result
	regexContainsSkipLines     []int
	regexContainsSkipData      []string
	regexContainsSkipResult    *Result
	regexAllUnmatchInput       string
	regexAllUnmatchResult      *Result
	regexAllSkipLines          []int
	regexAllSkipResult         *Result
	regexEmptyResult           *Result
	regexMixedSkipLines        []int
	regexMixedFilter           []string
	regexMixedData             []string
	regexMixedResult           *Result

	ltsvAllMatchInput         string
	ltsvAllMatchData          []string
	ltsvAllMatchResult        *Result
	ltsvContainsUnmatchInput  string
	ltsvContainsUnmatchData   []string
	ltsvContainsUnmatchResult *Result
	ltsvContainsFilter        []string
	ltsvContainsFilterData    []string
	ltsvContainsFilterResult  *Result
	ltsvContainsSkipLines     []int
	ltsvContainsSkipData      []string
	ltsvContainsSkipResult    *Result
	ltsvAllUnmatchInput       string
	ltsvAllUnmatchResult      *Result
	ltsvAllSkipLines          []int
	ltsvAllSkipResult         *Result
	ltsvEmptyResult           *Result
	ltsvMixedSkipLines        []int
	ltsvMixedFilter           []string
	ltsvMixedData             []string
	ltsvMixedResult           *Result

	regexAllMatchDataWithPrefix        []string
	regexContainsUnmatchDataWithPrefix []string
	regexAllUnmatchDataWithPrefix      []string
	ltsvAllMatchDataWithPrefix         []string
	ltsvContainsUnmatchDataWithPrefix  []string
	ltsvAllUnmatchDataWithPrefix       []string
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	stringPattern = `^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+)`
	stringInvalidCapturePattern = "(?P<field.1>[!-~]+)"
	stringCapturedGroupNotContainsPattern = "[!-~]+"
	stringNonNamedCapturedGroupContainsPattern = "(?P<field1>[!-~]+) ([!-~]+) (?P<field3>[!-~]+)"
	stringPatterns = []string{
		`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+) (?P<acl_required>[!-~]+)`,
		`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+)`,
		`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+)`,
		`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+)`,
		`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+)`,
	}
	stringCapturedGroupNotContainsPatterns = append(append(stringCapturedGroupNotContainsPatterns, stringPatterns...), stringCapturedGroupNotContainsPattern)
	stringNonNamedCapturedGroupContainsPatterns = append(append(stringNonNamedCapturedGroupContainsPatterns, stringPatterns...), stringNonNamedCapturedGroupContainsPattern)

	regexPattern = regexp.MustCompile(stringPattern)
	regexCapturedGroupNotContainsPattern = regexp.MustCompile(stringCapturedGroupNotContainsPattern)
	regexNonNamedCapturedGroupContainsPattern = regexp.MustCompile(stringNonNamedCapturedGroupContainsPattern)
	regexPatternsFunc := func() []*regexp.Regexp {
		var patterns []*regexp.Regexp
		for _, stringPattern := range stringPatterns {
			patterns = append(patterns, regexp.MustCompile(stringPattern))
		}
		return patterns
	}
	regexPatterns = regexPatternsFunc()
	regexCapturedGroupNotContainsPatterns = append(append(regexCapturedGroupNotContainsPatterns, regexPatterns...), regexCapturedGroupNotContainsPattern)
	regexNonNamedCapturedGroupContainsPatterns = append(append(regexNonNamedCapturedGroupContainsPatterns, regexPatterns...), regexNonNamedCapturedGroupContainsPattern)

	regexAllMatchInput = `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" - s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket43.s3.us-west-1.amazonaws.com TLSV1.1 -
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-" "S3Console/0.4" - 9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com TLSV1.1
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 - "-" "S3Console/0.4" - BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4" - Ke1bUcazaN1jWuUlPJaxF64cQVpUEhoZKEG/hmy/gijN/I1DeWqDfFvnpybfEseEME/u7ME1234= SigV2 ECDHE-RSA-AES128-SHA AuthHeader
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" -`
	regexAllMatchData = []string{
		`{"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket43?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`{"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`{"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"bucket_owner":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","bucket":"awsrandombucket89","time":"[03/Feb/2019:03:54:33 +0000]","remote_ip":"192.0.2.76","requester":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","request_id":"7B4A0FABBEXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket89?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"33","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
		`{"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexAllMatchResult = &Result{
		Total:     5,
		Matched:   5,
		Unmatched: 0,
		Excluded:  0,
		Skipped:   0,
		Source:    "",
		Errors:    []Errors{},
	}

	regexContainsUnmatchInput = `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" - s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket43.s3.us-west-1.amazonaws.com TLSV1.1 -
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-" "S3Console/0.4" - 9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com TLSV1.1
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 - "-" "S3Console/0.4" - BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4"
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" -`
	regexContainsUnmatchData = []string{
		`{"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket43?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`{"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`{"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexContainsUnmatchResult = &Result{
		Total:     5,
		Matched:   4,
		Unmatched: 1,
		Excluded:  0,
		Skipped:   0,
		Source:    "",
		Errors: []Errors{
			{
				LineNumber: 4,
				Line:       "d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\"",
			},
		},
	}

	regexContainsFilter = []string{"error_code == NoSuchBucketPolicy"}
	regexContainsFilterData = []string{`{"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`}
	regexContainsFilterResult = &Result{
		Total:     5,
		Matched:   1,
		Unmatched: 0,
		Excluded:  4,
		Skipped:   0,
		Source:    "",
		Errors:    []Errors{},
	}

	regexContainsSkipLines = []int{2, 4}
	regexContainsSkipData = []string{
		`{"no":"1","bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket43?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`{"no":"3","bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"no":"5","bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexContainsSkipResult = &Result{
		Total:     5,
		Matched:   3,
		Unmatched: 0,
		Excluded:  0,
		Skipped:   2,
		Source:    "",
		Errors:    []Errors{},
	}

	regexAllUnmatchInput = `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4"
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-"
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 -
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 -`
	regexAllUnmatchResult = &Result{
		Total:     5,
		Matched:   0,
		Unmatched: 5,
		Excluded:  0,
		Skipped:   0,
		Source:    "",
		Errors: []Errors{
			{
				LineNumber: 1,
				Line:       "a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket43?versioning HTTP/1.1\" 200 - 113 - 7 - \"-\" \"S3Console/0.4\"",
			},
			{
				LineNumber: 2,
				Line:       "3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - \"GET /awsrandombucket59?logging HTTP/1.1\" 200 - 242 - 11 - \"-\"",
			},
			{
				LineNumber: 3,
				Line:       "8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - \"GET /awsrandombucket12?policy HTTP/1.1\" 404 NoSuchBucketPolicy 297 - 38 -",
			},
			{
				LineNumber: 4,
				Line:       "d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33",
			},
			{
				LineNumber: 5,
				Line:       "01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket77?versioning HTTP/1.1\" 200 - 113 -",
			},
		},
	}

	regexAllSkipLines = []int{1, 2, 3, 4, 5}
	regexAllSkipResult = &Result{
		Total:     5,
		Matched:   0,
		Unmatched: 0,
		Excluded:  0,
		Skipped:   5,
		Source:    "",
		Errors:    []Errors{},
	}

	regexEmptyResult = &Result{
		Total:     0,
		Matched:   0,
		Unmatched: 0,
		Excluded:  0,
		Skipped:   0,
		Source:    "",
		Errors:    []Errors{},
	}

	regexMixedSkipLines = []int{1}
	regexMixedFilter = []string{"error_code != NoSuchBucketPolicy"}
	regexMixedData = []string{
		`{"no":"2","bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`{"no":"5","bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexMixedResult = &Result{
		Total:     5,
		Matched:   2,
		Unmatched: 1,
		Excluded:  1,
		Skipped:   1,
		Source:    "",
		Errors: []Errors{
			{
				LineNumber: 4,
				Line:       "d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\"",
			},
		},
	}

	ltsvAllMatchInput = `remote_host:192.168.1.1	remote_logname:-	remote_user:john	datetime:[12/Mar/2023:10:55:36 +0000]	request:GET /index.html HTTP/1.1	status:200	size:1024	referer:http://www.example.com/	user_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)
remote_host:172.16.0.2	remote_logname:-	remote_user:jane	datetime:[12/Mar/2023:10:56:10 +0000]	request:POST /login HTTP/1.1	status:303	size:532	referer:http://www.example.com/login	user_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)
remote_host:10.0.0.3	remote_logname:-	remote_user:mike	datetime:[12/Mar/2023:10:57:15 +0000]	request:GET /about.html HTTP/1.1	status:200	size:749	referer:http://www.example.com/	user_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)
remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	request:GET /products HTTP/1.1	status:404	size:0
remote_host:192.168.1.10	remote_logname:-	remote_user:chris	datetime:[12/Mar/2023:11:04:16 +0000]	request:DELETE /account HTTP/1.1	status:200	size:204`
	ltsvAllMatchData = []string{
		`{"remote_host":"192.168.1.1","remote_logname":"-","remote_user":"john","datetime":"[12/Mar/2023:10:55:36 +0000]","request":"GET /index.html HTTP/1.1","status":"200","size":"1024","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}`,
		`{"remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		`{"remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		`{"remote_host":"192.168.1.4","remote_logname":"-","remote_user":"anna","datetime":"[12/Mar/2023:10:58:24 +0000]","request":"GET /products HTTP/1.1","status":"404","size":"0"}`,
		`{"remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvAllMatchResult = &Result{
		Total:      5,
		Matched:    5,
		Unmatched:  0,
		Excluded:   0,
		Skipped:    0,
		Source:     "",
		ZipEntries: nil,
		Errors:     []Errors{},
	}

	ltsvContainsUnmatchInput = `remote_host:192.168.1.1	remote_logname:-	remote_user:john	datetime:[12/Mar/2023:10:55:36 +0000]	request:GET /index.html HTTP/1.1	status:200	size:1024	referer:http://www.example.com/	user_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)
remote_host:172.16.0.2	remote_logname:-	remote_user:jane	datetime:[12/Mar/2023:10:56:10 +0000]	request:POST /login HTTP/1.1	status:303	size:532	referer:http://www.example.com/login	user_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)
remote_host:10.0.0.3	remote_logname:-	remote_user:mike	datetime:[12/Mar/2023:10:57:15 +0000]	request:GET /about.html HTTP/1.1	status:200	size:749	referer:http://www.example.com/	user_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)
remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	request:GET /products HTTP/1.1	404	size:0
remote_host:192.168.1.10	remote_logname:-	remote_user:chris	datetime:[12/Mar/2023:11:04:16 +0000]	request:DELETE /account HTTP/1.1	status:200	size:204`
	ltsvContainsUnmatchData = []string{
		`{"no":"1","remote_host":"192.168.1.1","remote_logname":"-","remote_user":"john","datetime":"[12/Mar/2023:10:55:36 +0000]","request":"GET /index.html HTTP/1.1","status":"200","size":"1024","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}`,
		`{"no":"2","remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		`{"no":"3","remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		`{"no":"5","remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvContainsUnmatchResult = &Result{
		Total:     5,
		Matched:   4,
		Unmatched: 1,
		Excluded:  0,
		Skipped:   0,
		Source:    "",
		Errors: []Errors{
			{
				LineNumber: 4,
				Line:       "remote_host:192.168.1.4\tremote_logname:-\tremote_user:anna\tdatetime:[12/Mar/2023:10:58:24 +0000]\trequest:GET /products HTTP/1.1\t404\tsize:0",
			},
		},
	}

	ltsvContainsFilter = []string{"remote_user == mike"}
	ltsvContainsFilterData = []string{`{"remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`}
	ltsvContainsFilterResult = &Result{
		Total:     5,
		Matched:   1,
		Unmatched: 0,
		Excluded:  4,
		Skipped:   0,
		Source:    "",
		Errors:    []Errors{},
	}

	ltsvContainsSkipLines = []int{2, 4}
	ltsvContainsSkipData = []string{
		`{"no":"1","remote_host":"192.168.1.1","remote_logname":"-","remote_user":"john","datetime":"[12/Mar/2023:10:55:36 +0000]","request":"GET /index.html HTTP/1.1","status":"200","size":"1024","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}`,
		`{"no":"3","remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		`{"no":"5","remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvContainsSkipResult = &Result{
		Total:     5,
		Matched:   3,
		Unmatched: 0,
		Excluded:  0,
		Skipped:   2,
		Source:    "",
		Errors:    []Errors{},
	}

	ltsvAllUnmatchInput = `192.168.1.1	remote_logname:-	remote_user:john	datetime:[12/Mar/2023:10:55:36 +0000]	request:GET /index.html HTTP/1.1	status:200	size:1024	referer:http://www.example.com/	user_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)
remote_host:172.16.0.2	-	remote_user:jane	datetime:[12/Mar/2023:10:56:10 +0000]	request:POST /login HTTP/1.1	status:303	size:532	referer:http://www.example.com/login	user_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)
remote_host:10.0.0.3	remote_logname:-	mike	datetime:[12/Mar/2023:10:57:15 +0000]	request:GET /about.html HTTP/1.1	status:200	size:749	referer:http://www.example.com/	user_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)
remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	GET /products HTTP/1.1	status:404	size:0
remote_host:192.168.1.10	remote_logname:-	remote_user:chris	datetime:[12/Mar/2023:11:04:16 +0000]	request:DELETE /account HTTP/1.1	200	size:204`
	ltsvAllUnmatchResult = &Result{
		Total:     5,
		Matched:   0,
		Unmatched: 5,
		Excluded:  0,
		Skipped:   0,
		Source:    "",
		Errors: []Errors{
			{
				LineNumber: 1,
				Line:       "192.168.1.1\tremote_logname:-\tremote_user:john\tdatetime:[12/Mar/2023:10:55:36 +0000]\trequest:GET /index.html HTTP/1.1\tstatus:200\tsize:1024\treferer:http://www.example.com/\tuser_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			},
			{
				LineNumber: 2,
				Line:       "remote_host:172.16.0.2\t-\tremote_user:jane\tdatetime:[12/Mar/2023:10:56:10 +0000]\trequest:POST /login HTTP/1.1\tstatus:303\tsize:532\treferer:http://www.example.com/login\tuser_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
			},
			{
				LineNumber: 3,
				Line:       "remote_host:10.0.0.3\tremote_logname:-\tmike\tdatetime:[12/Mar/2023:10:57:15 +0000]\trequest:GET /about.html HTTP/1.1\tstatus:200\tsize:749\treferer:http://www.example.com/\tuser_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
			},
			{
				LineNumber: 4,
				Line:       "remote_host:192.168.1.4\tremote_logname:-\tremote_user:anna\tdatetime:[12/Mar/2023:10:58:24 +0000]\tGET /products HTTP/1.1\tstatus:404\tsize:0",
			},
			{
				LineNumber: 5,
				Line:       "remote_host:192.168.1.10\tremote_logname:-\tremote_user:chris\tdatetime:[12/Mar/2023:11:04:16 +0000]\trequest:DELETE /account HTTP/1.1\t200\tsize:204",
			},
		},
	}

	ltsvAllSkipLines = []int{1, 2, 3, 4, 5}
	ltsvAllSkipResult = &Result{
		Total:     5,
		Matched:   0,
		Unmatched: 0,
		Excluded:  0,
		Skipped:   5,
		Source:    "",
		Errors:    []Errors{},
	}

	ltsvEmptyResult = &Result{
		Total:     0,
		Matched:   0,
		Unmatched: 0,
		Excluded:  0,
		Skipped:   0,
		Source:    "",
		Errors:    []Errors{},
	}

	ltsvMixedSkipLines = []int{1}
	ltsvMixedFilter = []string{"remote_user != mike"}
	ltsvMixedData = []string{
		`{"no":"2","remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		`{"no":"5","remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvMixedResult = &Result{
		Total:     5,
		Matched:   2,
		Unmatched: 1,
		Excluded:  1,
		Skipped:   1,
		Source:    "",
		Errors: []Errors{
			{
				LineNumber: 4,
				Line:       "remote_host:192.168.1.4\tremote_logname:-\tremote_user:anna\tdatetime:[12/Mar/2023:10:58:24 +0000]\trequest:GET /products HTTP/1.1\t404\tsize:0",
			},
		},
	}

	regexAllMatchDataWithPrefix = []string{
		`[ PROCESSED ] {"no":"1","bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket43?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`[ PROCESSED ] {"no":"2","bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`[ PROCESSED ] {"no":"3","bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`[ PROCESSED ] {"no":"4","bucket_owner":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","bucket":"awsrandombucket89","time":"[03/Feb/2019:03:54:33 +0000]","remote_ip":"192.0.2.76","requester":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","request_id":"7B4A0FABBEXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket89?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"33","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
		`[ PROCESSED ] {"no":"5","bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexContainsUnmatchDataWithPrefix = []string{
		`[ PROCESSED ] {"no":"1","bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket43?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`[ PROCESSED ] {"no":"2","bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`[ PROCESSED ] {"no":"3","bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`[ UNMATCHED ] d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4"`,
		`[ PROCESSED ] {"no":"5","bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexAllUnmatchDataWithPrefix = []string{
		`[ UNMATCHED ] a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4"`,
		`[ UNMATCHED ] 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-"`,
		`[ UNMATCHED ] 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 -`,
		`[ UNMATCHED ] d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33`,
		`[ UNMATCHED ] 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 -`,
	}
	ltsvAllMatchDataWithPrefix = []string{
		`[ PROCESSED ] {"no":"1","remote_host":"192.168.1.1","remote_logname":"-","remote_user":"john","datetime":"[12/Mar/2023:10:55:36 +0000]","request":"GET /index.html HTTP/1.1","status":"200","size":"1024","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}`,
		`[ PROCESSED ] {"no":"2","remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		`[ PROCESSED ] {"no":"3","remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		`[ PROCESSED ] {"no":"4","remote_host":"192.168.1.4","remote_logname":"-","remote_user":"anna","datetime":"[12/Mar/2023:10:58:24 +0000]","request":"GET /products HTTP/1.1","status":"404","size":"0"}`,
		`[ PROCESSED ] {"no":"5","remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvContainsUnmatchDataWithPrefix = []string{
		`[ PROCESSED ] {"no":"1","remote_host":"192.168.1.1","remote_logname":"-","remote_user":"john","datetime":"[12/Mar/2023:10:55:36 +0000]","request":"GET /index.html HTTP/1.1","status":"200","size":"1024","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}`,
		`[ PROCESSED ] {"no":"2","remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		`[ PROCESSED ] {"no":"3","remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		`[ UNMATCHED ] remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	request:GET /products HTTP/1.1	404	size:0`,
		`[ PROCESSED ] {"no":"5","remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvAllUnmatchDataWithPrefix = []string{
		`[ UNMATCHED ] 192.168.1.1	remote_logname:-	remote_user:john	datetime:[12/Mar/2023:10:55:36 +0000]	request:GET /index.html HTTP/1.1	status:200	size:1024	referer:http://www.example.com/	user_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)`,
		`[ UNMATCHED ] remote_host:172.16.0.2	-	remote_user:jane	datetime:[12/Mar/2023:10:56:10 +0000]	request:POST /login HTTP/1.1	status:303	size:532	referer:http://www.example.com/login	user_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)`,
		`[ UNMATCHED ] remote_host:10.0.0.3	remote_logname:-	mike	datetime:[12/Mar/2023:10:57:15 +0000]	request:GET /about.html HTTP/1.1	status:200	size:749	referer:http://www.example.com/	user_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)`,
		`[ UNMATCHED ] remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	GET /products HTTP/1.1	status:404	size:0`,
		`[ UNMATCHED ] remote_host:192.168.1.10	remote_logname:-	remote_user:chris	datetime:[12/Mar/2023:11:04:16 +0000]	request:DELETE /account HTTP/1.1	200	size:204`,
	}
}

type wantResult struct {
	result    *Result
	source    string
	inputType inputType
}

func assertResult(t *testing.T, want wantResult, got *Result) {
	t.Helper()
	if want.result == nil {
		return
	}
	if !reflect.DeepEqual(got.Total, want.result.Total) {
		t.Errorf("total:\n\ngot:\n%v\nwant:\n%v\n", got.Total, want.result.Total)
	}
	if !reflect.DeepEqual(got.Matched, want.result.Matched) {
		t.Errorf("matched:\n\ngot:\n%v\nwant:\n%v\n", got.Matched, want.result.Matched)
	}
	if !reflect.DeepEqual(got.Unmatched, want.result.Unmatched) {
		t.Errorf("unmatched:\n\ngot:\n%v\nwant:\n%v\n", got.Unmatched, want.result.Unmatched)
	}
	if !reflect.DeepEqual(got.Excluded, want.result.Excluded) {
		t.Errorf("excluded:\n\ngot:\n%v\nwant:\n%v\n", got.Excluded, want.result.Excluded)
	}
	if !reflect.DeepEqual(got.Skipped, want.result.Skipped) {
		t.Errorf("skipped:\n\ngot:\n%v\nwant:\n%v\n", got.Skipped, want.result.Skipped)
	}
	if !reflect.DeepEqual(got.Errors, want.result.Errors) {
		t.Errorf("errors:\n\ngot:\n%v\nwant:\n%v\n", got.Errors, want.result.Errors)
	}
	if !reflect.DeepEqual(got.Source, want.source) {
		t.Errorf("source:\n\ngot:\n%v\nwant:\n%v\n", got.Source, want.source)
	}
	if !reflect.DeepEqual(got.inputType, want.inputType) {
		t.Errorf("inputType:\n\ngot:\n%v\nwant:\n%v\n", got.inputType, want.inputType)
	}
}

func Test_parse(t *testing.T) {
	type args struct {
		ctx      context.Context
		input    io.Reader
		decoder  lineDecoder
		opt      Option
		patterns []*regexp.Regexp
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "regex: all match",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(regexAllMatchInput),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(regexAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    regexAllMatchResult,
				source:    "",
				inputType: inputTypeStream,
			},
			wantErr: false,
		},
		{
			name: "ltsv: all match",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvAllMatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    ltsvAllMatchResult,
				source:    "",
				inputType: inputTypeStream,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			got, err := parse(tt.args.ctx, tt.args.input, output, tt.args.patterns, tt.args.decoder, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if out := output.String(); !reflect.DeepEqual(out, tt.wantOutput) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", out, tt.wantOutput)
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}

func Test_parseString(t *testing.T) {
	type args struct {
		ctx      context.Context
		s        string
		decoder  lineDecoder
		opt      Option
		patterns []*regexp.Regexp
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "regex: all match",
			args: args{
				ctx:     context.Background(),
				s:       regexAllMatchInput,
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(regexAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    regexAllMatchResult,
				source:    "",
				inputType: inputTypeString,
			},
			wantErr: false,
		},
		{
			name: "ltsv: all match",
			args: args{
				ctx:     context.Background(),
				s:       ltsvAllMatchInput,
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    ltsvAllMatchResult,
				source:    "",
				inputType: inputTypeString,
			},
			wantErr: false,
		},
		{
			name: "regex: line handler returns error",
			args: args{
				ctx:     context.Background(),
				s:       regexAllMatchInput,
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler: func(_, _ []string, _ int, _, _ bool) (string, error) {
						return "", fmt.Errorf("error")
					},
				},
				patterns: regexPatterns,
			},
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeString,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			got, err := parseString(tt.args.ctx, tt.args.s, output, tt.args.patterns, tt.args.decoder, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if out := output.String(); !reflect.DeepEqual(out, tt.wantOutput) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", out, tt.wantOutput)
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}

func Test_parseFile(t *testing.T) {
	type args struct {
		ctx      context.Context
		filePath string
		decoder  lineDecoder
		opt      Option
		patterns []*regexp.Regexp
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "regex: all match",
			args: args{
				ctx:      context.Background(),
				filePath: filepath.Join("testdata", "sample_s3_all_match.log"),
				decoder:  regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(regexAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    regexAllMatchResult,
				source:    "sample_s3_all_match.log",
				inputType: inputTypeFile,
			},
			wantErr: false,
		},
		{
			name: "ltsv: all match",
			args: args{
				ctx:      context.Background(),
				filePath: filepath.Join("testdata", "sample_s3_all_match.log"),
				decoder:  ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    ltsvAllMatchResult,
				source:    "sample_s3_all_match.log",
				inputType: inputTypeFile,
			},
			wantErr: false,
		},
		{
			name: "regex: line handler returns error",
			args: args{
				ctx:      context.Background(),
				filePath: filepath.Join("testdata", "sample_s3_all_match.log"),
				decoder:  regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler: func(_, _ []string, _ int, _, _ bool) (string, error) {
						return "", fmt.Errorf("error")
					},
				},
				patterns: regexPatterns,
			},
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeFile,
			},
			wantErr: true,
		},
		{
			name: "regex: input file does not exists",
			args: args{
				ctx:      context.Background(),
				filePath: filepath.Join("testdata", "sample_ltsv_all_match.log.dummy"),
				decoder:  regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeFile,
			},
			wantErr: true,
		},
		{
			name: "regex: input path is directory not file",
			args: args{
				ctx:      context.Background(),
				filePath: "testdata",
				decoder:  regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeFile,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			got, err := parseFile(tt.args.ctx, tt.args.filePath, output, tt.args.patterns, tt.args.decoder, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}

func Test_parseGzip(t *testing.T) {
	type args struct {
		ctx      context.Context
		gzipPath string
		decoder  lineDecoder
		opt      Option
		patterns []*regexp.Regexp
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "regex: all match",
			args: args{
				ctx:      context.Background(),
				gzipPath: filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				decoder:  regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(regexAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    regexAllMatchResult,
				source:    "sample_s3_all_match.log.gz",
				inputType: inputTypeGzip,
			},
			wantErr: false,
		},
		{
			name: "ltsv: all match",
			args: args{
				ctx:      context.Background(),
				gzipPath: filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
				decoder:  ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    ltsvAllMatchResult,
				source:    "sample_ltsv_all_match.log.gz",
				inputType: inputTypeGzip,
			},
			wantErr: false,
		},
		{
			name: "regex: line handler returns error",
			args: args{
				ctx:      context.Background(),
				gzipPath: filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				decoder:  regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler: func(_, _ []string, _ int, _, _ bool) (string, error) {
						return "", fmt.Errorf("error")
					},
				},
				patterns: regexPatterns,
			},
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeGzip,
			},
			wantErr: true,
		},
		{
			name: "regex: input file does not exists",
			args: args{
				ctx:      context.Background(),
				gzipPath: filepath.Join("testdata", "sample_ltsv_all_match.log.gz.dummy"),
				decoder:  regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeGzip,
			},
			wantErr: true,
		},
		{
			name: "regex: input path is directory not file",
			args: args{
				ctx:      context.Background(),
				gzipPath: "testdata",
				decoder:  regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeGzip,
			},
			wantErr: true,
		},
		{
			name: "regex: input file is not gzip",
			args: args{
				ctx:      context.Background(),
				gzipPath: filepath.Join("testdata", "sample_s3_all_match.log"),
				decoder:  regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeGzip,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			got, err := parseGzip(tt.args.ctx, tt.args.gzipPath, output, tt.args.patterns, tt.args.decoder, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}

func Test_parseZipEntries(t *testing.T) {
	type args struct {
		ctx         context.Context
		zipPath     string
		globPattern string
		decoder     lineDecoder
		opt         Option
		patterns    []*regexp.Regexp
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "regex: all match",
			args: args{
				ctx:         context.Background(),
				zipPath:     filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				globPattern: "*",
				decoder:     regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(regexAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    regexAllMatchResult,
				source:    "sample_s3_all_match.log.zip",
				inputType: inputTypeZip,
			},
			wantErr: false,
		},
		{
			name: "ltsv: all match",
			args: args{
				ctx:         context.Background(),
				zipPath:     filepath.Join("testdata", "sample_ltsv_all_match.log.zip"),
				globPattern: "*",
				decoder:     ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    ltsvAllMatchResult,
				source:    "sample_ltsv_all_match.log.zip",
				inputType: inputTypeZip,
			},
			wantErr: false,
		},
		{
			name: "regex: line handler returns error",
			args: args{
				ctx:         context.Background(),
				zipPath:     filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				globPattern: "*",
				decoder:     regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler: func(_, _ []string, _ int, _, _ bool) (string, error) {
						return "", fmt.Errorf("error")
					},
				},
				patterns: regexPatterns,
			},
			wantOutput: "",
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeZip,
			},
			wantErr: true,
		},
		{
			name: "regex: input path is directory not file",
			args: args{
				ctx:         context.Background(),
				zipPath:     "testdata",
				globPattern: "*",
				decoder:     regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: "",
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeZip,
			},
			wantErr: true,
		},
		{
			name: "regex: input file is not zip",
			args: args{
				ctx:         context.Background(),
				zipPath:     filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				globPattern: "*",
				decoder:     regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: "",
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeZip,
			},
			wantErr: true,
		},
		{
			name: "regex: input file does not exists",
			args: args{
				ctx:         context.Background(),
				zipPath:     filepath.Join("testdata", "sample_s3_all_match.log.zip.dummy"),
				globPattern: "*",
				decoder:     regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: "",
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeZip,
			},
			wantErr: true,
		},
		{
			name: "regex: multi entries",
			args: args{
				ctx:         context.Background(),
				zipPath:     filepath.Join("testdata", "sample_s3.zip"),
				globPattern: "*",
				decoder:     regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: `{"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket43?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}
{"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}
{"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}
{"bucket_owner":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","bucket":"awsrandombucket89","time":"[03/Feb/2019:03:54:33 +0000]","remote_ip":"192.0.2.76","requester":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","request_id":"7B4A0FABBEXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket89?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"33","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}
{"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}
{"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket43?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}
{"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}
{"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}
{"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}
`,
			wantResult: wantResult{
				result: &Result{
					Total:     15,
					Matched:   9,
					Unmatched: 6,
					Excluded:  0,
					Skipped:   0,
					Source:    "sample_s3.zip",
					ZipEntries: []string{
						"sample_s3_all_match.log",
						"sample_s3_contains_unmatch.log",
						"sample_s3_all_unmatch.log",
					},
					Errors: []Errors{
						{
							Entry:      "sample_s3_contains_unmatch.log",
							LineNumber: 4,
							Line:       `d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4"`,
						},
						{
							Entry:      "sample_s3_all_unmatch.log",
							LineNumber: 1,
							Line:       "a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket43?versioning HTTP/1.1\" 200 - 113 - 7 - \"-\" \"S3Console/0.4\"",
						},
						{
							Entry:      "sample_s3_all_unmatch.log",
							LineNumber: 2,
							Line:       "3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - \"GET /awsrandombucket59?logging HTTP/1.1\" 200 - 242 - 11 - \"-\"",
						},
						{
							Entry:      "sample_s3_all_unmatch.log",
							LineNumber: 3,
							Line:       "8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - \"GET /awsrandombucket12?policy HTTP/1.1\" 404 NoSuchBucketPolicy 297 - 38 -",
						},
						{
							Entry:      "sample_s3_all_unmatch.log",
							LineNumber: 4,
							Line:       "d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33",
						},
						{
							Entry:      "sample_s3_all_unmatch.log",
							LineNumber: 5,
							Line:       "01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket77?versioning HTTP/1.1\" 200 - 113 -",
						},
					},
					inputType: inputTypeZip,
				},
				source:    "sample_s3.zip",
				inputType: inputTypeZip,
			},
			wantErr: false,
		},
		{
			name: "regex: multi entries and glob pattern filtering",
			args: args{
				ctx:         context.Background(),
				zipPath:     filepath.Join("testdata", "sample_s3.zip"),
				globPattern: "*all_match*",
				decoder:     regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(regexAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result:    nil,
				source:    "sample_s3_all_match.log",
				inputType: inputTypeZip,
			},
			wantErr: false,
		},
		{
			name: "regex: multi entries and glob pattern returns error",
			args: args{
				ctx:         context.Background(),
				zipPath:     filepath.Join("testdata", "log.zip"),
				globPattern: "[",
				decoder:     regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: "",
			wantResult: wantResult{
				result:    nil,
				source:    "",
				inputType: inputTypeZip,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			got, err := parseZipEntries(tt.args.ctx, tt.args.zipPath, tt.args.globPattern, output, tt.args.patterns, tt.args.decoder, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if out := output.String(); !reflect.DeepEqual(out, tt.wantOutput) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", out, tt.wantOutput)
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}

func Test_parser(t *testing.T) {
	type args struct {
		ctx      context.Context
		input    io.Reader
		decoder  lineDecoder
		opt      Option
		patterns []*regexp.Regexp
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantResult wantResult
		wantErr    bool
	}{
		{
			name: "regex: all match",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(regexAllMatchInput),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(regexAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result: regexAllMatchResult,
			},
			wantErr: false,
		},
		{
			name: "regex: contains unmatch",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(regexContainsUnmatchInput),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(regexContainsUnmatchData, "\n") + "\n",
			wantResult: wantResult{
				result: regexContainsUnmatchResult,
			},
			wantErr: false,
		},
		{
			name: "regex: contains skip flag",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(regexAllMatchInput),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    regexContainsSkipLines,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(regexContainsSkipData, "\n") + "\n",
			wantResult: wantResult{
				result: regexContainsSkipResult,
			},
			wantErr: false,
		},
		{
			name: "regex: all unmatch",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(regexAllUnmatchInput),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: "",
			wantResult: wantResult{
				result: regexAllUnmatchResult,
			},
			wantErr: false,
		},
		{
			name: "regex: all skip",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(regexAllMatchInput),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    regexAllSkipLines,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: "",
			wantResult: wantResult{
				result: regexAllSkipResult,
			},
			wantErr: false,
		},
		{
			name: "regex: mixed",
			args: args{
				ctx:      context.Background(),
				input:    strings.NewReader(regexContainsUnmatchInput),
				patterns: regexPatterns,
				decoder:  regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      regexMixedFilter,
					SkipLines:    regexMixedSkipLines,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
			},
			wantOutput: strings.Join(regexMixedData, "\n") + "\n",
			wantResult: wantResult{
				result: regexMixedResult,
			},
			wantErr: false,
		},
		{
			name: "regex: with filter",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(regexAllMatchInput),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      regexContainsFilter,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: strings.Join(regexContainsFilterData, "\n") + "\n",
			wantResult: wantResult{
				result: regexContainsFilterResult,
			},
			wantErr: false,
		},
		{
			name: "regex: nil input",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(""),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: "",
			wantResult: wantResult{
				result: regexEmptyResult,
			},
			wantErr: false,
		},
		{
			name: "regex: line handler returns error",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(regexAllMatchInput),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler: func(_, _ []string, _ int, _, _ bool) (string, error) {
						return "", fmt.Errorf("error")
					},
				},
				patterns: regexPatterns,
			},
			wantOutput: "",
			wantResult: wantResult{
				result: nil,
			},
			wantErr: true,
		},
		{
			name: "regex: nil pattern",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(regexAllMatchInput),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: "",
			wantResult: wantResult{
				result: nil,
			},
			wantErr: true,
		},
		{
			name: "ltsv: all match",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvAllMatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvAllMatchData, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvAllMatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: contains unmatch",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvContainsUnmatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvContainsUnmatchData, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvContainsUnmatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: contains skip flag",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvAllMatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    ltsvContainsSkipLines,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvContainsSkipData, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvContainsSkipResult,
			},
			wantErr: false,
		},
		{
			name: "regex: select by columns",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(regexAllMatchInput),
				decoder: regexLineDecoder,
				opt: Option{
					Labels:       []string{"no", "bucket_owner"},
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: regexPatterns,
			},
			wantOutput: `{"no":"1","bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a"}
{"no":"2","bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23"}
{"no":"3","bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2"}
{"no":"4","bucket_owner":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01"}
{"no":"5","bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f"}
`,
			wantResult: wantResult{
				result: regexAllMatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: all unmatch",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvAllUnmatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: "",
			wantResult: wantResult{
				result: ltsvAllUnmatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: all skip",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvAllUnmatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    ltsvAllSkipLines,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: "",
			wantResult: wantResult{
				result: ltsvAllSkipResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: mixed",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvContainsUnmatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      ltsvMixedFilter,
					SkipLines:    ltsvMixedSkipLines,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvMixedData, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvMixedResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: with filter",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvAllMatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      ltsvContainsFilter,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvContainsFilterData, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvContainsFilterResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: nil input",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(""),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: "",
			wantResult: wantResult{
				result: ltsvEmptyResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: line handler returns error",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvAllMatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler: func(_, _ []string, _ int, _, _ bool) (string, error) {
						return "", fmt.Errorf("error")
					},
				},
				patterns: regexPatterns,
			},
			wantOutput: "",
			wantResult: wantResult{
				result: nil,
			},
			wantErr: true,
		},
		{
			name: "ltsv: select by columns",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvAllMatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       []string{"no", "remote_host"},
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: `{"no":"1","remote_host":"192.168.1.1"}
{"no":"2","remote_host":"172.16.0.2"}
{"no":"3","remote_host":"10.0.0.3"}
{"no":"4","remote_host":"192.168.1.4"}
{"no":"5","remote_host":"192.168.1.10"}
`,
			wantResult: wantResult{
				result: regexAllMatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: with prefix",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvAllMatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       true,
					UnmatchLines: false,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvAllMatchDataWithPrefix, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvAllMatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: with unmatch lines",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvContainsUnmatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       true,
					UnmatchLines: true,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: strings.Join(ltsvContainsUnmatchDataWithPrefix, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvContainsUnmatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: with invalid filter",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvContainsUnmatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      []string{"aaa := bbb"},
					SkipLines:    nil,
					Prefix:       true,
					UnmatchLines: true,
					LineNumber:   true,
					LineHandler:  JSONLineHandler,
				},
				patterns: nil,
			},
			wantOutput: "",
			wantResult: wantResult{},
			wantErr:    true,
		},
		{
			name: "ltsv: with tsv handler",
			args: args{
				ctx:     context.Background(),
				input:   strings.NewReader(ltsvAllMatchInput),
				decoder: ltsvLineDecoder,
				opt: Option{
					Labels:       nil,
					Filters:      nil,
					SkipLines:    nil,
					Prefix:       false,
					UnmatchLines: false,
					LineNumber:   false,
					LineHandler:  TSVLineHandler,
				},
				patterns: nil,
			},
			wantOutput: `remote_host	remote_logname	remote_user	datetime	request	status	size	referer	user_agent
192.168.1.1	-	john	[12/Mar/2023:10:55:36 +0000]	GET /index.html HTTP/1.1	200	1024	http://www.example.com/	Mozilla/5.0 (Windows NT 10.0; Win64; x64)
172.16.0.2	-	jane	[12/Mar/2023:10:56:10 +0000]	POST /login HTTP/1.1	303	532	http://www.example.com/login	Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)
10.0.0.3	-	mike	[12/Mar/2023:10:57:15 +0000]	GET /about.html HTTP/1.1	200	749	http://www.example.com/	Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)
192.168.1.4	-	anna	[12/Mar/2023:10:58:24 +0000]	GET /products HTTP/1.1	404	0
192.168.1.10	-	chris	[12/Mar/2023:11:04:16 +0000]	DELETE /account HTTP/1.1	200	204
`,
			wantResult: wantResult{
				result: ltsvAllMatchResult,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			got, err := parser(tt.args.ctx, tt.args.input, output, tt.args.patterns, tt.args.decoder, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if out := output.String(); !reflect.DeepEqual(out, tt.wantOutput) {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", out, tt.wantOutput)
			}
			assertResult(t, tt.wantResult, got)
		})
	}
}

func Test_handleFile(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty path",
			args: args{
				filePath: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := handleFile(tt.args.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_handleGzip(t *testing.T) {
	type args struct {
		gzipPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty path",
			args: args{
				gzipPath: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := handleGzip(tt.args.gzipPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_handleZipEntries(t *testing.T) {
	type args struct {
		zipPath     string
		globPattern string
		fn          func(f *zip.File) error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty path",
			args: args{
				zipPath: "",
			},
			wantErr: true,
		},
		{
			name: "invalid glob pattern",
			args: args{
				zipPath:     filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				globPattern: "[",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleZipEntries(tt.args.zipPath, tt.args.globPattern, tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
			}
		})
	}
}

func Test_applyPrefix(t *testing.T) {
	type args struct {
		line   string
		prefix string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "single line",
			args: args{
				line:   "This is a test line.",
				prefix: "> ",
			},
			want: "> This is a test line.",
		},
		{
			name: "multiple lines",
			args: args{
				line:   "First line.\nSecond line.",
				prefix: "> ",
			},
			want: "> First line.\n> Second line.",
		},
		{
			name: "empty line",
			args: args{
				line:   "",
				prefix: "> ",
			},
			want: "> ",
		},
		{
			name: "prefix with special characters",
			args: args{
				line:   "Special characters in prefix.",
				prefix: "**>>",
			},
			want: "**>>Special characters in prefix.",
		},
		{
			name: "multiple lines with empty line",
			args: args{
				line:   "First line.\n\nThird line.",
				prefix: "- ",
			},
			want: "- First line.\n- \n- Third line.",
		},
		{
			name: "only newline character",
			args: args{
				line:   "\n",
				prefix: "* ",
			},
			want: "* \n* ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := applyPrefix(tt.args.line, tt.args.prefix); got != tt.want {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", got, tt.want)
			}
		})
	}
}

func Test_applyFilter(t *testing.T) {
	type args struct {
		labels  []string
		values  []string
		filters []string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid filters with all true",
			args: args{
				labels:  []string{"status", "score", "age"},
				values:  []string{"active", "150", "25"},
				filters: []string{"status == active", "score > 100", "age < 30"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "valid filters with one false",
			args: args{
				labels:  []string{"status", "score", "age"},
				values:  []string{"inactive", "150", "25"},
				filters: []string{"status == active", "score > 100", "age < 30"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "invalid filter expression syntax",
			args: args{
				labels:  []string{"status"},
				values:  []string{"active"},
				filters: []string{"status = active"}, // invalid syntax
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "filter with label not in labels",
			args: args{
				labels:  []string{"status", "age"},
				values:  []string{"active", "25"},
				filters: []string{"status == active", "nonexistentlabel > 10"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid numeric comparison",
			args: args{
				labels:  []string{"age"},
				values:  []string{"dummy"},
				filters: []string{"age < 25"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "unknown operator in filter",
			args: args{
				labels:  []string{"status"},
				values:  []string{"active"},
				filters: []string{"status ?? active"},
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applyFilter(tt.args.labels, tt.args.values, tt.args.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("applyFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("applyFilter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFilter(t *testing.T) {
	tests := []struct {
		name           string
		labels         []string
		filters        []string
		want           map[string]func(string) (bool, error)
		wantErr        bool
		wantClosureErr map[string]bool
	}{
		{
			name:   "valid filters with correct labels",
			labels: []string{"status", "score", "name", "age"},
			filters: []string{
				"status == active",
				"score > 100",
				"name =~ ^John",
				"age != 30",
			},
			want: map[string]func(string) (bool, error){
				"status": func(v string) (bool, error) { return v == "active", nil },
				"score": func(v string) (bool, error) {
					val, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return false, err
					}
					return val > 100, nil
				},
				"name": func(v string) (bool, error) { return regexp.MustCompile("^John").MatchString(v), nil },
				"age":  func(v string) (bool, error) { return v != "30", nil },
			},
			wantErr: false,
			wantClosureErr: map[string]bool{
				"score": true,
			},
		},
		{
			name:   "label not in filters should not cause error",
			labels: []string{"status", "age"},
			filters: []string{
				"status == active",
			},
			want: map[string]func(string) (bool, error){
				"status": func(v string) (bool, error) { return v == "active", nil },
			},
			wantErr: false,
			wantClosureErr: map[string]bool{
				"status": false,
			},
		},
		{
			name:   "invalid operator in string filter",
			labels: []string{"status"},
			filters: []string{
				"status ~~ active",
			},
			wantErr: true,
		},
		{
			name:   "invalid regex pattern in filter",
			labels: []string{"name"},
			filters: []string{
				"name =~ [",
			},
			wantErr: true,
		},
		{
			name:   "invalid numeric value in filter",
			labels: []string{"score"},
			filters: []string{
				"score > not_a_number",
			},
			wantErr: true,
		},
		{
			name:   "filter with label not in labels",
			labels: []string{"status", "age"},
			filters: []string{
				"status == active",
				"nonexistentlabel > 10",
			},
			wantErr: true,
		},
		{
			name:   "invalid filter expression",
			labels: []string{"invalidfilter"},
			filters: []string{
				"invalidfilter",
			},
			wantErr: true,
		},
		{
			name:   "unknown operator",
			labels: []string{"status"},
			filters: []string{
				"status ?? active",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFilter(tt.labels, tt.filters)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && err == nil {
				for key, wantFunc := range tt.want {
					gotFunc, ok := got[key]
					if !ok {
						t.Errorf("getFilter() missing key %v", key)
						continue
					}
					wantResult, wantErr := wantFunc("test")
					gotResult, gotErr := gotFunc("test")
					if (gotErr != nil) != tt.wantClosureErr[key] {
						t.Errorf("getFilter() closure error for key %v = %v, wantClosureErr %v", key, gotErr, tt.wantClosureErr[key])
					} else if gotErr == nil && wantErr == nil && wantResult != gotResult {
						t.Errorf("getFilter() result for key %v = %v, want %v", key, gotResult, wantResult)
					}
				}
			}
		})
	}
}

func Test_getStringFilter(t *testing.T) {
	type args struct {
		operator string
		value    string
	}
	tests := []struct {
		name       string
		args       args
		wantInvoke string
		wantResult bool
		wantErr    bool
	}{
		{
			name: "equals true",
			args: args{
				operator: "==",
				value:    "test",
			},
			wantInvoke: "test",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "equals false",
			args: args{
				operator: "==",
				value:    "test",
			},
			wantInvoke: "not_test",
			wantResult: false,
			wantErr:    false,
		},
		{
			name: "not equals true",
			args: args{
				operator: "!=",
				value:    "test",
			},
			wantInvoke: "not_test",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "case insensitive equals true",
			args: args{
				operator: "==*",
				value:    "Test",
			},
			wantInvoke: "test",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "case insensitive not equals true",
			args: args{
				operator: "!=*",
				value:    "Test",
			},
			wantInvoke: "Other",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "unknown operator error",
			args: args{
				operator: "??",
				value:    "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getStringFilter(tt.args.operator, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotResult, gotErr := got(tt.wantInvoke)
				if gotErr != nil != tt.wantErr {
					t.Errorf("\ngot:\n%v\nwant:\n%v\n", gotErr, tt.wantErr)
					return
				}
				if gotResult != tt.wantResult {
					t.Errorf("\ngot:\n%v\nwant:\n%v\n", gotResult, tt.wantResult)
				}
			}
		})
	}
}

func Test_getNumericFilter(t *testing.T) {
	type args struct {
		operator string
		value    string
	}
	tests := []struct {
		name           string
		args           args
		wantInvoke     string
		wantResult     bool
		wantErr        bool
		wantClosureErr bool
	}{
		{
			name: "greater than true",
			args: args{
				operator: ">",
				value:    "5",
			},
			wantInvoke: "6",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "greater than false",
			args: args{
				operator: ">",
				value:    "5",
			},
			wantInvoke: "4",
			wantResult: false,
			wantErr:    false,
		},
		{
			name: "less than true",
			args: args{
				operator: "<",
				value:    "5",
			},
			wantInvoke: "4",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "less than equal true",
			args: args{
				operator: "<=",
				value:    "5",
			},
			wantInvoke: "5",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "less than equal false",
			args: args{
				operator: "<=",
				value:    "5",
			},
			wantInvoke: "6",
			wantResult: false,
			wantErr:    false,
		},
		{
			name: "greater or equal true",
			args: args{
				operator: ">=",
				value:    "5",
			},
			wantInvoke: "5",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "invalid value error",
			args: args{
				operator: ">",
				value:    "not_a_number",
			},
			wantErr: true,
		},
		{
			name: "invalid comparison value error 1",
			args: args{
				operator: ">",
				value:    "5",
			},
			wantInvoke:     "not_a_number",
			wantResult:     false,
			wantClosureErr: true,
		},
		{
			name: "invalid comparison value error 2",
			args: args{
				operator: ">=",
				value:    "5",
			},
			wantInvoke:     "not_a_number",
			wantResult:     false,
			wantClosureErr: true,
		},
		{
			name: "invalid comparison value error 3",
			args: args{
				operator: "<",
				value:    "5",
			},
			wantInvoke:     "not_a_number",
			wantResult:     false,
			wantClosureErr: true,
		},
		{
			name: "invalid comparison value error 4",
			args: args{
				operator: "<=",
				value:    "5",
			},
			wantInvoke:     "not_a_number",
			wantResult:     false,
			wantClosureErr: true,
		},
		{
			name: "unknown operator error",
			args: args{
				operator: "??",
				value:    "5",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilter, err := getNumericFilter(tt.args.operator, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotResult, gotErr := gotFilter(tt.wantInvoke)
				if (gotErr != nil) != tt.wantClosureErr {
					t.Errorf("\ngot:\n%v\nwant:\n%v\n", gotErr, tt.wantClosureErr)
					return
				}
				if !tt.wantClosureErr && gotResult != tt.wantResult {
					t.Errorf("\ngot:\n%v\nwant:\n%v\n", gotResult, tt.wantResult)
				}
			}
		})
	}
}

func Test_getRegexFilter(t *testing.T) {
	type args struct {
		operator string
		value    string
	}
	tests := []struct {
		name       string
		args       args
		wantInvoke string
		wantResult bool
		wantErr    bool
	}{
		{
			name: "match with =~ operator",
			args: args{
				operator: "=~",
				value:    "^test",
			},
			wantInvoke: "test123",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "no match with =~ operator",
			args: args{
				operator: "=~",
				value:    "^test",
			},
			wantInvoke: "nope",
			wantResult: false,
			wantErr:    false,
		},
		{
			name: "match with !~ operator",
			args: args{
				operator: "!~",
				value:    "^test",
			},
			wantInvoke: "nope",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "no match with !~ operator",
			args: args{
				operator: "!~",
				value:    "^test",
			},
			wantInvoke: "test123",
			wantResult: false,
			wantErr:    false,
		},
		{
			name: "case insensitive match with =~* operator",
			args: args{
				operator: "=~*",
				value:    "^TEST",
			},
			wantInvoke: "test123",
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "case insensitive no match with !~* operator",
			args: args{
				operator: "!~*",
				value:    "^TEST",
			},
			wantInvoke: "TEST123",
			wantResult: false,
			wantErr:    false,
		},
		{
			name: "invalid regex pattern",
			args: args{
				operator: "=~",
				value:    "[z-a]",
			},
			wantErr: true,
		},
		{
			name: "unknown operator error",
			args: args{
				operator: "??",
				value:    "5",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRegexFilter(tt.args.operator, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("\ngot:\n%v\nwant:\n%v\n", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				gotResult, _ := got(tt.wantInvoke)
				if gotResult != tt.wantResult {
					t.Errorf("\ngot:\n%v\nwant:\n%v\n", gotResult, tt.wantResult)
				}
			}
		})
	}
}
