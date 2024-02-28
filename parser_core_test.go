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

	regexAllMatchInput         string
	regexAllMatchData          []string
	regexAllMatchResult        *Result
	regexContainsUnmatchInput  string
	regexContainsUnmatchData   []string
	regexContainsUnmatchResult *Result
	regexContainsKeyword       []string
	regexContainsKeywordData   []string
	regexContainsKeywordResult *Result
	regexContainsSkipLines     []int
	regexContainsSkipData      []string
	regexContainsSkipResult    *Result
	regexAllUnmatchInput       string
	regexAllUnmatchResult      *Result
	regexAllSkipLines          []int
	regexAllSkipResult         *Result
	regexEmptyResult           *Result
	regexMixedSkipLines        []int
	regexMixedData             []string
	regexMixedResult           *Result

	ltsvAllMatchInput         string
	ltsvAllMatchData          []string
	ltsvAllMatchResult        *Result
	ltsvContainsUnmatchInput  string
	ltsvContainsUnmatchData   []string
	ltsvContainsUnmatchResult *Result
	ltsvContainsKeyword       []string
	ltsvContainsKeywordData   []string
	ltsvContainsKeywordResult *Result
	ltsvContainsSkipLines     []int
	ltsvContainsSkipData      []string
	ltsvContainsSkipResult    *Result
	ltsvAllUnmatchInput       string
	ltsvAllUnmatchResult      *Result
	ltsvAllSkipLines          []int
	ltsvAllSkipResult         *Result
	ltsvEmptyResult           *Result
	ltsvMixedSkipLines        []int
	ltsvMixedData             []string
	ltsvMixedResult           *Result

	regexAllMatchDataForStream        []string
	regexContainsUnmatchDataForStream []string
	regexAllUnmatchDataForStream      []string
	ltsvAllMatchDataForStream         []string
	ltsvContainsUnmatchDataForStream  []string
	ltsvAllUnmatchDataForStream       []string
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	regexPattern = regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+)`)
	regexCapturedGroupNotContainsPattern = regexp.MustCompile("[!-~]+")
	regexNonNamedCapturedGroupContainsPattern = regexp.MustCompile("(?P<field1>[!-~]+) ([!-~]+) (?P<field3>[!-~]+)")
	regexPatterns = []*regexp.Regexp{
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+) (?P<acl_required>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+)`),
	}
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

	regexContainsKeyword = []string{"NoSuchBucketPolicy"}
	regexContainsKeywordData = []string{`{"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`}
	regexContainsKeywordResult = &Result{
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
	regexMixedData = []string{
		`{"no":"2","bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`{"no":"3","bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"no":"5","bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexMixedResult = &Result{
		Total:     5,
		Matched:   3,
		Unmatched: 1,
		Excluded:  0,
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

	ltsvContainsKeyword = []string{"mike"}
	ltsvContainsKeywordData = []string{`{"remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`}
	ltsvContainsKeywordResult = &Result{
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
	ltsvMixedData = []string{
		`{"no":"2","remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		`{"no":"3","remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		`{"no":"5","remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvMixedResult = &Result{
		Total:     5,
		Matched:   3,
		Unmatched: 1,
		Excluded:  0,
		Skipped:   1,
		Source:    "",
		Errors: []Errors{
			{
				LineNumber: 4,
				Line:       "remote_host:192.168.1.4\tremote_logname:-\tremote_user:anna\tdatetime:[12/Mar/2023:10:58:24 +0000]\trequest:GET /products HTTP/1.1\t404\tsize:0",
			},
		},
	}

	regexAllMatchDataForStream = []string{
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket43?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"bucket_owner":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","bucket":"awsrandombucket89","time":"[03/Feb/2019:03:54:33 +0000]","remote_ip":"192.0.2.76","requester":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","request_id":"7B4A0FABBEXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket89?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"33","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexContainsUnmatchDataForStream = []string{
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket43?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","method":"GET","request_uri":"/awsrandombucket59?logging","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","method":"GET","request_uri":"/awsrandombucket12?policy","protocol":"HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4"`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","method":"GET","request_uri":"/awsrandombucket77?versioning","protocol":"HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	regexAllUnmatchDataForStream = []string{
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4"`,
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-"`,
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 -`,
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33`,
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 -`,
	}
	ltsvAllMatchDataForStream = []string{
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"remote_host":"192.168.1.1","remote_logname":"-","remote_user":"john","datetime":"[12/Mar/2023:10:55:36 +0000]","request":"GET /index.html HTTP/1.1","status":"200","size":"1024","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"remote_host":"192.168.1.4","remote_logname":"-","remote_user":"anna","datetime":"[12/Mar/2023:10:58:24 +0000]","request":"GET /products HTTP/1.1","status":"404","size":"0"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvContainsUnmatchDataForStream = []string{
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"remote_host":"192.168.1.1","remote_logname":"-","remote_user":"john","datetime":"[12/Mar/2023:10:55:36 +0000]","request":"GET /index.html HTTP/1.1","status":"200","size":"1024","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64)"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"remote_host":"172.16.0.2","remote_logname":"-","remote_user":"jane","datetime":"[12/Mar/2023:10:56:10 +0000]","request":"POST /login HTTP/1.1","status":"303","size":"532","referer":"http://www.example.com/login","user_agent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)"}`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"remote_host":"10.0.0.3","remote_logname":"-","remote_user":"mike","datetime":"[12/Mar/2023:10:57:15 +0000]","request":"GET /about.html HTTP/1.1","status":"200","size":"749","referer":"http://www.example.com/","user_agent":"Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"}`,
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	request:GET /products HTTP/1.1	404	size:0`,
		"\033[1;32m" + "[ PROCESSED ] " + "\033[0m" + `{"remote_host":"192.168.1.10","remote_logname":"-","remote_user":"chris","datetime":"[12/Mar/2023:11:04:16 +0000]","request":"DELETE /account HTTP/1.1","status":"200","size":"204"}`,
	}
	ltsvAllUnmatchDataForStream = []string{
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `192.168.1.1	remote_logname:-	remote_user:john	datetime:[12/Mar/2023:10:55:36 +0000]	request:GET /index.html HTTP/1.1	status:200	size:1024	referer:http://www.example.com/	user_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)`,
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `remote_host:172.16.0.2	-	remote_user:jane	datetime:[12/Mar/2023:10:56:10 +0000]	request:POST /login HTTP/1.1	status:303	size:532	referer:http://www.example.com/login	user_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)`,
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `remote_host:10.0.0.3	remote_logname:-	mike	datetime:[12/Mar/2023:10:57:15 +0000]	request:GET /about.html HTTP/1.1	status:200	size:749	referer:http://www.example.com/	user_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)`,
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	GET /products HTTP/1.1	status:404	size:0`,
		"\033[1;31m" + "[ UNMATCHED ] " + "\033[0m" + `remote_host:192.168.1.10	remote_logname:-	remote_user:chris	datetime:[12/Mar/2023:11:04:16 +0000]	request:DELETE /account HTTP/1.1	200	size:204`,
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
		t.Errorf("\ngot:\n%v\nwant:\n%v\n", got.Total, want.result.Total)
	}
	if !reflect.DeepEqual(got.Matched, want.result.Matched) {
		t.Errorf("\ngot:\n%v\nwant:\n%v\n", got.Matched, want.result.Matched)
	}
	if !reflect.DeepEqual(got.Unmatched, want.result.Unmatched) {
		t.Errorf("\ngot:\n%v\nwant:\n%v\n", got.Unmatched, want.result.Unmatched)
	}
	if !reflect.DeepEqual(got.Excluded, want.result.Excluded) {
		t.Errorf("\ngot:\n%v\nwant:\n%v\n", got.Excluded, want.result.Excluded)
	}
	if !reflect.DeepEqual(got.Skipped, want.result.Skipped) {
		t.Errorf("\ngot:\n%v\nwant:\n%v\n", got.Skipped, want.result.Skipped)
	}
	if !reflect.DeepEqual(got.Errors, want.result.Errors) {
		t.Errorf("\ngot:\n%v\nwant:\n%v\n", got.Errors, want.result.Errors)
	}
	if !reflect.DeepEqual(got.Source, want.source) {
		t.Errorf("\ngot:\n%v\nwant:\n%v\n", got.Source, want.source)
	}
	if !reflect.DeepEqual(got.inputType, want.inputType) {
		t.Errorf("\ngot:\n%v\nwant:\n%v\n", got.inputType, want.inputType)
	}
}

func Test_parse(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx            context.Context
		input          io.Reader
		patterns       []*regexp.Regexp
		keywords       []string
		labels         []string
		hasPrefix      bool
		disableUnmatch bool
		decoder        lineDecoder
		handler        LineHandler
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
				ctx:       ctx,
				input:     strings.NewReader(regexAllMatchInput),
				patterns:  regexPatterns,
				keywords:  nil,
				labels:    nil,
				hasPrefix: false,
				decoder:   regexLineDecoder,
				handler:   JSONLineHandler,
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
				ctx:       ctx,
				input:     strings.NewReader(ltsvAllMatchInput),
				patterns:  nil,
				keywords:  nil,
				labels:    nil,
				hasPrefix: false,
				decoder:   ltsvLineDecoder,
				handler:   JSONLineHandler,
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
			got, err := parse(tt.args.ctx, tt.args.input, output, tt.args.patterns, tt.args.keywords, tt.args.labels, tt.args.hasPrefix, tt.args.disableUnmatch, tt.args.decoder, tt.args.handler)
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
		s             string
		patterns      []*regexp.Regexp
		keywords      []string
		labels        []string
		skipLines     []int
		hasLineNumber bool
		decoder       lineDecoder
		handler       LineHandler
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
				s:             regexAllMatchInput,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				s:             ltsvAllMatchInput,
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
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
				s:             regexAllMatchInput,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler: func(labels, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error) {
					return "", fmt.Errorf("error")
				},
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
			got, err := parseString(tt.args.s, output, tt.args.patterns, tt.args.keywords, tt.args.labels, tt.args.skipLines, tt.args.hasLineNumber, tt.args.decoder, tt.args.handler)
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
		filePath      string
		patterns      []*regexp.Regexp
		keywords      []string
		labels        []string
		skipLines     []int
		hasLineNumber bool
		decoder       lineDecoder
		handler       LineHandler
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
				filePath:      filepath.Join("testdata", "sample_s3_all_match.log"),
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				filePath:      filepath.Join("testdata", "sample_s3_all_match.log"),
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
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
				filePath:      filepath.Join("testdata", "sample_s3_all_match.log"),
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler: func(labels, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error) {
					return "", fmt.Errorf("error")
				},
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
				filePath:      filepath.Join("testdata", "sample_ltsv_all_match.log.dummy"),
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				filePath:      "testdata",
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
			got, err := parseFile(tt.args.filePath, output, tt.args.patterns, tt.args.keywords, tt.args.labels, tt.args.skipLines, tt.args.hasLineNumber, tt.args.decoder, tt.args.handler)
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
		gzipPath      string
		patterns      []*regexp.Regexp
		keywords      []string
		labels        []string
		skipLines     []int
		hasLineNumber bool
		decoder       lineDecoder
		handler       LineHandler
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
				gzipPath:      filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				gzipPath:      filepath.Join("testdata", "sample_ltsv_all_match.log.gz"),
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
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
				gzipPath:      filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler: func(labels, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error) {
					return "", fmt.Errorf("error")
				},
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
				gzipPath:      filepath.Join("testdata", "sample_ltsv_all_match.log.gz.dummy"),
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				gzipPath:      "testdata",
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				gzipPath:      filepath.Join("testdata", "sample_s3_all_match.log"),
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
			got, err := parseGzip(tt.args.gzipPath, output, tt.args.patterns, tt.args.keywords, tt.args.labels, tt.args.skipLines, tt.args.hasLineNumber, tt.args.decoder, tt.args.handler)
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
		zipPath       string
		globPattern   string
		patterns      []*regexp.Regexp
		keywords      []string
		labels        []string
		skipLines     []int
		hasLineNumber bool
		decoder       lineDecoder
		handler       LineHandler
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
				zipPath:       filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				globPattern:   "*",
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				zipPath:       filepath.Join("testdata", "sample_ltsv_all_match.log.zip"),
				globPattern:   "*",
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
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
				zipPath:       filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				globPattern:   "*",
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler: func(labels, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error) {
					return "", fmt.Errorf("error")
				},
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
				zipPath:       "testdata",
				globPattern:   "*",
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				zipPath:       filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				globPattern:   "*",
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				zipPath:       filepath.Join("testdata", "sample_s3_all_match.log.zip.dummy"),
				globPattern:   "*",
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				zipPath:       filepath.Join("testdata", "sample_s3.zip"),
				globPattern:   "*",
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				zipPath:       filepath.Join("testdata", "sample_s3.zip"),
				globPattern:   "*all_match*",
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				zipPath:       filepath.Join("testdata", "log.zip"),
				globPattern:   "[",
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				skipLines:     nil,
				hasLineNumber: false,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
			got, err := parseZipEntries(tt.args.zipPath, tt.args.globPattern, output, tt.args.patterns, tt.args.keywords, tt.args.labels, tt.args.skipLines, tt.args.hasLineNumber, tt.args.decoder, tt.args.handler)
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
		input         io.Reader
		patterns      []*regexp.Regexp
		keywords      []string
		labels        []string
		skipLines     []int
		hasLineNumber bool
		decoder       lineDecoder
		handler       LineHandler
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
				input:         strings.NewReader(regexAllMatchInput),
				skipLines:     nil,
				hasLineNumber: false,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(regexContainsUnmatchInput),
				skipLines:     nil,
				hasLineNumber: false,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(regexAllMatchInput),
				skipLines:     regexContainsSkipLines,
				hasLineNumber: true,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(regexAllUnmatchInput),
				skipLines:     nil,
				hasLineNumber: true,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(regexAllMatchInput),
				skipLines:     regexAllSkipLines,
				hasLineNumber: true,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(regexContainsUnmatchInput),
				skipLines:     regexMixedSkipLines,
				hasLineNumber: true,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
			},
			wantOutput: strings.Join(regexMixedData, "\n") + "\n",
			wantResult: wantResult{
				result: regexMixedResult,
			},
			wantErr: false,
		},
		{
			name: "regex: nil input",
			args: args{
				input:         strings.NewReader(""),
				skipLines:     nil,
				hasLineNumber: true,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(regexAllMatchInput),
				skipLines:     nil,
				hasLineNumber: true,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				decoder:       regexLineDecoder,
				handler: func(labels, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error) {
					return "", fmt.Errorf("error")
				},
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
				input:         strings.NewReader(regexAllMatchInput),
				skipLines:     nil,
				hasLineNumber: true,
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(ltsvAllMatchInput),
				skipLines:     nil,
				hasLineNumber: false,
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(ltsvContainsUnmatchInput),
				skipLines:     nil,
				hasLineNumber: true,
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(ltsvAllMatchInput),
				skipLines:     ltsvContainsSkipLines,
				hasLineNumber: true,
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
			},
			wantOutput: strings.Join(ltsvContainsSkipData, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvContainsSkipResult,
			},
			wantErr: false,
		},
		{
			name: "regex: keyword filtering",
			args: args{
				input:         strings.NewReader(regexAllMatchInput),
				patterns:      regexPatterns,
				skipLines:     nil,
				hasLineNumber: false,
				keywords:      regexContainsKeyword,
				labels:        nil,
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
			},
			wantOutput: strings.Join(regexContainsKeywordData, "\n") + "\n",
			wantResult: wantResult{
				result: regexContainsKeywordResult,
			},
			wantErr: false,
		},
		{
			name: "regex: select by columns",
			args: args{
				input:         strings.NewReader(regexAllMatchInput),
				patterns:      regexPatterns,
				skipLines:     nil,
				hasLineNumber: true,
				keywords:      nil,
				labels:        []string{"no", "bucket_owner"},
				decoder:       regexLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(ltsvAllUnmatchInput),
				skipLines:     nil,
				hasLineNumber: true,
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(ltsvAllUnmatchInput),
				skipLines:     ltsvAllSkipLines,
				hasLineNumber: true,
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(ltsvContainsUnmatchInput),
				skipLines:     ltsvMixedSkipLines,
				hasLineNumber: true,
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
			},
			wantOutput: strings.Join(ltsvMixedData, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvMixedResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: nil input",
			args: args{
				input:         strings.NewReader(""),
				skipLines:     nil,
				hasLineNumber: true,
				patterns:      nil,
				keywords:      nil,
				labels:        nil,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
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
				input:         strings.NewReader(ltsvAllMatchInput),
				skipLines:     nil,
				hasLineNumber: true,
				patterns:      regexPatterns,
				keywords:      nil,
				labels:        nil,
				decoder:       ltsvLineDecoder,
				handler: func(labels, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			wantOutput: "",
			wantResult: wantResult{
				result: nil,
			},
			wantErr: true,
		},
		{
			name: "ltsv: keyword filtering",
			args: args{
				input:         strings.NewReader(ltsvAllMatchInput),
				patterns:      nil,
				skipLines:     nil,
				hasLineNumber: false,
				keywords:      ltsvContainsKeyword,
				labels:        nil,
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
			},
			wantOutput: strings.Join(ltsvContainsKeywordData, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvContainsKeywordResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: select by columns",
			args: args{
				input:         strings.NewReader(ltsvAllMatchInput),
				patterns:      nil,
				skipLines:     nil,
				hasLineNumber: true,
				keywords:      nil,
				labels:        []string{"no", "remote_host"},
				decoder:       ltsvLineDecoder,
				handler:       JSONLineHandler,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			got, err := parser(tt.args.input, output, tt.args.patterns, tt.args.keywords, tt.args.labels, tt.args.skipLines, tt.args.hasLineNumber, tt.args.decoder, tt.args.handler)
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

func Test_streamer(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx            context.Context
		input          io.Reader
		patterns       []*regexp.Regexp
		keywords       []string
		labels         []string
		hasPrefix      bool
		disableUnmatch bool
		decoder        lineDecoder
		handler        LineHandler
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
				ctx:            ctx,
				input:          strings.NewReader(regexAllMatchInput),
				patterns:       regexPatterns,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      true,
				disableUnmatch: false,
				decoder:        regexLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: strings.Join(regexAllMatchDataForStream, "\n") + "\n",
			wantResult: wantResult{
				result: regexAllMatchResult,
			},
			wantErr: false,
		},
		{
			name: "regex: contains unmatch",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(regexContainsUnmatchInput),
				patterns:       regexPatterns,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      true,
				disableUnmatch: false,
				decoder:        regexLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: strings.Join(regexContainsUnmatchDataForStream, "\n") + "\n",
			wantResult: wantResult{
				result: regexContainsUnmatchResult,
			},
			wantErr: false,
		},
		{
			name: "regex: all unmatch",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(regexAllUnmatchInput),
				patterns:       regexPatterns,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      true,
				disableUnmatch: false,
				decoder:        regexLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: strings.Join(regexAllUnmatchDataForStream, "\n") + "\n",
			wantResult: wantResult{
				result: regexAllUnmatchResult,
			},
			wantErr: false,
		},
		{
			name: "regex: nil input",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(""),
				patterns:       regexPatterns,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      false,
				disableUnmatch: false,
				decoder:        regexLineDecoder,
				handler:        JSONLineHandler,
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
				ctx:            ctx,
				input:          strings.NewReader(regexAllMatchInput),
				patterns:       regexPatterns,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      true,
				disableUnmatch: false,
				decoder:        regexLineDecoder,
				handler: func(labels, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error) {
					return "", fmt.Errorf("error")
				},
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
				ctx:            ctx,
				input:          strings.NewReader(regexAllMatchInput),
				patterns:       nil,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      true,
				disableUnmatch: false,
				decoder:        regexLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: "",
			wantResult: wantResult{
				result: nil,
			},
			wantErr: true,
		},
		{
			name: "regex: keyword filtering",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(regexAllMatchInput),
				patterns:       regexPatterns,
				keywords:       regexContainsKeyword,
				labels:         nil,
				hasPrefix:      false,
				disableUnmatch: false,
				decoder:        regexLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: strings.Join(regexContainsKeywordData, "\n") + "\n",
			wantResult: wantResult{
				result: regexContainsKeywordResult,
			},
			wantErr: false,
		},
		{
			name: "regex: select by columns",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(regexAllMatchInput),
				patterns:       regexPatterns,
				keywords:       nil,
				labels:         []string{"no", "bucket_owner"},
				hasPrefix:      false,
				disableUnmatch: false,
				decoder:        regexLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: `{"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a"}
{"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23"}
{"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2"}
{"bucket_owner":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01"}
{"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f"}
`,
			wantResult: wantResult{
				result: regexAllMatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: all match",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(ltsvAllMatchInput),
				patterns:       nil,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      true,
				disableUnmatch: false,
				decoder:        ltsvLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: strings.Join(ltsvAllMatchDataForStream, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvAllMatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: contains unmatch",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(ltsvContainsUnmatchInput),
				patterns:       nil,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      true,
				disableUnmatch: false,
				decoder:        ltsvLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: strings.Join(ltsvContainsUnmatchDataForStream, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvContainsUnmatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: all unmatch",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(ltsvAllUnmatchInput),
				patterns:       nil,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      true,
				disableUnmatch: false,
				decoder:        ltsvLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: strings.Join(ltsvAllUnmatchDataForStream, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvAllUnmatchResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: nil input",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(""),
				patterns:       nil,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      true,
				disableUnmatch: false,
				decoder:        ltsvLineDecoder,
				handler:        JSONLineHandler,
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
				ctx:            ctx,
				input:          strings.NewReader(ltsvAllMatchInput),
				patterns:       nil,
				keywords:       nil,
				labels:         nil,
				hasPrefix:      true,
				disableUnmatch: false,
				decoder:        ltsvLineDecoder,
				handler: func(labels, values []string, lineNumber int, hasLineNumber, isFirst bool) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			wantOutput: "",
			wantResult: wantResult{
				result: nil,
			},
			wantErr: true,
		},
		{
			name: "ltsv: keyword filtering",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(ltsvAllMatchInput),
				patterns:       nil,
				keywords:       ltsvContainsKeyword,
				labels:         nil,
				hasPrefix:      false,
				disableUnmatch: false,
				decoder:        ltsvLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: strings.Join(ltsvContainsKeywordData, "\n") + "\n",
			wantResult: wantResult{
				result: ltsvContainsKeywordResult,
			},
			wantErr: false,
		},
		{
			name: "ltsv: select by columns",
			args: args{
				ctx:            ctx,
				input:          strings.NewReader(ltsvAllMatchInput),
				patterns:       nil,
				keywords:       nil,
				labels:         []string{"no", "remote_host"},
				hasPrefix:      false,
				disableUnmatch: false,
				decoder:        ltsvLineDecoder,
				handler:        JSONLineHandler,
			},
			wantOutput: `{"remote_host":"192.168.1.1"}
{"remote_host":"172.16.0.2"}
{"remote_host":"10.0.0.3"}
{"remote_host":"192.168.1.4"}
{"remote_host":"192.168.1.10"}
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
			got, err := streamer(tt.args.ctx, tt.args.input, output, tt.args.patterns, tt.args.keywords, tt.args.labels, tt.args.hasPrefix, tt.args.disableUnmatch, tt.args.decoder, tt.args.handler)
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
