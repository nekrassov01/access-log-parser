package logparser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var (
	pattern                               *regexp.Regexp
	capturedGroupNotContainsPattern       *regexp.Regexp
	nonNamedCapturedGroupContainsPattern  *regexp.Regexp
	patterns                              []*regexp.Regexp
	capturedGroupNotContainsPatterns      []*regexp.Regexp
	nonNamedCapturedGroupContainsPatterns []*regexp.Regexp
	allMatchInput                         string
	allMatchData                          []string
	allMatchMetadata                      *Metadata
	allMatchMetadataSerialized            string
	containsUnmatchInput                  string
	containsUnmatchData                   []string
	containsUnmatchMetadata               *Metadata
	containsUnmatchMetadataSerialized     string
	containsSkipLines                     []int
	containsSkipData                      []string
	containsSkipMetadata                  *Metadata
	containsSkipMetadataSerialized        string
	allUnmatchInput                       string
	allUnmatchMetadata                    *Metadata
	allUnmatchMetadataSerialized          string
	allSkipLines                          []int
	allSkipMetadata                       *Metadata
	allSkipMetadataSerialized             string
	emptyMetadata                         *Metadata
	emptyMetadataSerialized               string
	mixedSkipLines                        []int
	mixedData                             []string
	mixedMetadata                         *Metadata
	mixedMetadataSerialized               string
	fileNotFoundMessage                   string
	fileNotFoundMessageWindows            string
	isDirMessage                          string
	isDirMessageWindows                   string
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	sep := " "
	basePattern := []string{
		`^(?P<bucket_owner>[!-~]+)`,
		`(?P<bucket>[!-~]+)`,
		`(?P<time>\[[ -~]+ [0-9+]+\])`,
		`(?P<remote_ip>[!-~]+)`,
		`(?P<requester>[!-~]+)`,
		`(?P<request_id>[!-~]+)`,
		`(?P<operation>[!-~]+)`,
		`(?P<key>[!-~]+)`,
		`"(?P<request_uri>[ -~]+)"`,
		`(?P<http_status>\d{1,3})`,
		`(?P<error_code>[!-~]+)`,
		`(?P<bytes_sent>[\d\-.]+)`,
		`(?P<object_size>[\d\-.]+)`,
		`(?P<total_time>[\d\-.]+)`,
		`(?P<turn_around_time>[\d\-.]+)`,
		`"(?P<referer>[ -~]+)"`,
		`"(?P<user_agent>[ -~]+)"`,
		`(?P<version_id>[!-~]+)`,
	}
	additions := [][]string{
		{
			`(?P<host_id>[!-~]+)`,
			`(?P<signature_version>[!-~]+)`,
			`(?P<cipher_suite>[!-~]+)`,
			`(?P<authentication_type>[!-~]+)`,
			`(?P<host_header>[!-~]+)`,
			`(?P<tls_version>[!-~]+)`,
			`(?P<access_point_arn>[!-~]+)`,
			`(?P<acl_required>[!-~]+)`,
		},
		{
			`(?P<host_id>[!-~]+)`,
			`(?P<signature_version>[!-~]+)`,
			`(?P<cipher_suite>[!-~]+)`,
			`(?P<authentication_type>[!-~]+)`,
			`(?P<host_header>[!-~]+)`,
			`(?P<tls_version>[!-~]+)`,
			`(?P<access_point_arn>[!-~]+)`,
		},
		{
			`(?P<host_id>[!-~]+)`,
			`(?P<signature_version>[!-~]+)`,
			`(?P<cipher_suite>[!-~]+)`,
			`(?P<authentication_type>[!-~]+)`,
			`(?P<host_header>[!-~]+)`,
			`(?P<tls_version>[!-~]+)`,
		},
		{
			`(?P<host_id>[!-~]+)`,
			`(?P<signature_version>[!-~]+)`,
			`(?P<cipher_suite>[!-~]+)`,
			`(?P<authentication_type>[!-~]+)`,
			`(?P<host_header>[!-~]+)`,
		},
		{},
	}

	pattern = regexp.MustCompile(strings.Join(basePattern, sep))
	capturedGroupNotContainsPattern = regexp.MustCompile("[!-~]+")
	nonNamedCapturedGroupContainsPattern = regexp.MustCompile("(?P<field1>[!-~]+) ([!-~]+) (?P<field3>[!-~]+)")
	patterns = make([]*regexp.Regexp, len(additions))
	for i, addition := range additions {
		patterns[i] = regexp.MustCompile(strings.Join(append(basePattern, addition...), sep))
	}
	capturedGroupNotContainsPatterns = append(append(capturedGroupNotContainsPatterns, patterns...), capturedGroupNotContainsPattern)
	nonNamedCapturedGroupContainsPatterns = append(append(nonNamedCapturedGroupContainsPatterns, patterns...), nonNamedCapturedGroupContainsPattern)

	allMatchInput = `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" - s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket43.s3.us-west-1.amazonaws.com TLSV1.1 -
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-" "S3Console/0.4" - 9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com TLSV1.1
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 - "-" "S3Console/0.4" - BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4" - Ke1bUcazaN1jWuUlPJaxF64cQVpUEhoZKEG/hmy/gijN/I1DeWqDfFvnpybfEseEME/u7ME1234= SigV2 ECDHE-RSA-AES128-SHA AuthHeader
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" -`
	allMatchData = []string{
		`{"index":1,"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket43?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`{"index":2,"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","request_uri":"GET /awsrandombucket59?logging HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`{"index":3,"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","request_uri":"GET /awsrandombucket12?policy HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"index":4,"bucket_owner":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","bucket":"awsrandombucket89","time":"[03/Feb/2019:03:54:33 +0000]","remote_ip":"192.0.2.76","requester":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01","request_id":"7B4A0FABBEXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket89?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"33","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
		`{"index":5,"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket77?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	allMatchMetadata = &Metadata{
		Total:     5,
		Matched:   5,
		Unmatched: 0,
		Skipped:   0,
		Source:    "",
		Errors:    nil,
	}
	allMatchMetadataSerialized = `{"total":5,"matched":5,"unmatched":0,"skipped":0,"source":"%s","errors":null}`

	containsUnmatchInput = `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" - s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket43.s3.us-west-1.amazonaws.com TLSV1.1 -
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-" "S3Console/0.4" - 9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com TLSV1.1
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 - "-" "S3Console/0.4" - BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4"
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" -`
	containsUnmatchData = []string{
		`{"index":1,"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket43?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`{"index":2,"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","request_uri":"GET /awsrandombucket59?logging HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`{"index":3,"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","request_uri":"GET /awsrandombucket12?policy HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"index":5,"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket77?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	containsUnmatchMetadata = &Metadata{
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
	containsUnmatchMetadataSerialized = `{"total":5,"matched":4,"unmatched":1,"skipped":0,"source":"%s","errors":[{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\""}]}`

	containsSkipLines = []int{2, 4}
	containsSkipData = []string{
		`{"index":1,"bucket_owner":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","bucket":"awsrandombucket43","time":"[16/Feb/2019:11:23:45 +0000]","remote_ip":"192.0.2.132","requester":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket43?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket43.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1","access_point_arn":"-"}`,
		`{"index":3,"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","request_uri":"GET /awsrandombucket12?policy HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"index":5,"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket77?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	containsSkipMetadata = &Metadata{
		Total:     5,
		Matched:   3,
		Unmatched: 0,
		Skipped:   2,
		Source:    "",
		Errors:    nil,
	}
	containsSkipMetadataSerialized = `{"total":5,"matched":3,"unmatched":0,"skipped":2,"source":"%s","errors":null}`

	allUnmatchInput = `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4"
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-"
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 -
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 -`
	allUnmatchMetadata = &Metadata{
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
	allUnmatchMetadataSerialized = `{"total":5,"matched":0,"unmatched":5,"skipped":0,"source":"%s","errors":[{"index":1,"record":"a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket43?versioning HTTP/1.1\" 200 - 113 - 7 - \"-\" \"S3Console/0.4\""},{"index":2,"record":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - \"GET /awsrandombucket59?logging HTTP/1.1\" 200 - 242 - 11 - \"-\""},{"index":3,"record":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - \"GET /awsrandombucket12?policy HTTP/1.1\" 404 NoSuchBucketPolicy 297 - 38 -"},{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33"},{"index":5,"record":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket77?versioning HTTP/1.1\" 200 - 113 -"}]}`

	allSkipLines = []int{1, 2, 3, 4, 5}
	allSkipMetadata = &Metadata{
		Total:     5,
		Matched:   0,
		Unmatched: 0,
		Skipped:   5,
		Source:    "",
		Errors:    nil,
	}
	allSkipMetadataSerialized = `{"total":5,"matched":0,"unmatched":0,"skipped":5,"source":"%s","errors":null}`

	emptyMetadata = &Metadata{
		Total:     0,
		Matched:   0,
		Unmatched: 0,
		Skipped:   0,
		Source:    "",
		Errors:    nil,
	}
	emptyMetadataSerialized = `{"total":0,"matched":0,"unmatched":0,"skipped":0,"source":"%s","errors":null}`

	mixedSkipLines = []int{1}
	mixedData = []string{
		`{"index":2,"bucket_owner":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","bucket":"awsrandombucket59","time":"[24/Feb/2019:07:45:11 +0000]","remote_ip":"192.0.2.45","requester":"3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23","request_id":"891CE47D2EXAMPLE","operation":"REST.GET.LOGGING_STATUS","key":"-","request_uri":"GET /awsrandombucket59?logging HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"242","object_size":"-","total_time":"11","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com","tls_version":"TLSV1.1"}`,
		`{"index":3,"bucket_owner":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","bucket":"awsrandombucket12","time":"[12/Feb/2019:18:32:21 +0000]","remote_ip":"192.0.2.189","requester":"8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2","request_id":"A1206F460EXAMPLE","operation":"REST.GET.BUCKETPOLICY","key":"-","request_uri":"GET /awsrandombucket12?policy HTTP/1.1","http_status":"404","error_code":"NoSuchBucketPolicy","bytes_sent":"297","object_size":"-","total_time":"38","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-","host_id":"BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234=","signature_version":"SigV2","cipher_suite":"ECDHE-RSA-AES128-GCM-SHA256","authentication_type":"AuthHeader","host_header":"awsrandombucket59.s3.us-west-1.amazonaws.com"}`,
		`{"index":5,"bucket_owner":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","bucket":"awsrandombucket77","time":"[28/Feb/2019:14:12:59 +0000]","remote_ip":"192.0.2.213","requester":"01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f","request_id":"3E57427F3EXAMPLE","operation":"REST.GET.VERSIONING","key":"-","request_uri":"GET /awsrandombucket77?versioning HTTP/1.1","http_status":"200","error_code":"-","bytes_sent":"113","object_size":"-","total_time":"7","turn_around_time":"-","referer":"-","user_agent":"S3Console/0.4","version_id":"-"}`,
	}
	mixedMetadata = &Metadata{
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
	mixedMetadataSerialized = `{"total":5,"matched":3,"unmatched":1,"skipped":1,"source":"%s","errors":[{"index":4,"record":"d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - \"GET /awsrandombucket89?versioning HTTP/1.1\" 200 - 113 - 33 - \"-\" \"S3Console/0.4\""}]}`

	fileNotFoundMessage = "no such file or directory"
	fileNotFoundMessageWindows = "The system cannot find the file specified."
	isDirMessage = "is a directory"
	isDirMessageWindows = "Incorrect function."
}

func TestNew(t *testing.T) {
	customLineHandler := func(matches []string, fields []string, index int) (string, error) {
		return "", nil
	}
	customMetadataHandler := func(metadata *Metadata) (string, error) {
		return "", nil
	}
	type args struct {
		lineHandler     LineHandler
		metadataHandler MetadataHandler
	}
	type want struct {
		parser *Parser
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "use default handlers",
			args: args{
				lineHandler:     nil,
				metadataHandler: nil,
			},
			want: want{
				parser: &Parser{
					Patterns:        nil,
					LineHandler:     DefaultLineHandler,
					MetadataHandler: DefaultMetadataHandler,
				},
			},
		},
		{
			name: "use custom handlers",
			args: args{
				lineHandler:     customLineHandler,
				metadataHandler: customMetadataHandler,
			},
			want: want{
				parser: &Parser{
					Patterns:        nil,
					LineHandler:     customLineHandler,
					MetadataHandler: customMetadataHandler,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []Option{}
			if tt.args.lineHandler != nil {
				opts = append(opts, WithLineHandler(tt.args.lineHandler))
			}
			if tt.args.metadataHandler != nil {
				opts = append(opts, WithMetadataHandler(tt.args.metadataHandler))
			}
			p := NewParser(opts...)
			if reflect.ValueOf(p.LineHandler).Pointer() != reflect.ValueOf(tt.want.parser.LineHandler).Pointer() {
				t.Errorf("got: %v, want: %v", p.LineHandler, tt.want.parser.LineHandler)
			}
			if reflect.ValueOf(p.MetadataHandler).Pointer() != reflect.ValueOf(tt.want.parser.MetadataHandler).Pointer() {
				t.Errorf("got: %v, want: %v", p.MetadataHandler, tt.want.parser.MetadataHandler)
			}
		})
	}
}

func TestAddPattern(t *testing.T) {
	type args struct {
		pattern *regexp.Regexp
	}
	type want struct {
		parser *Parser
		err    error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				pattern: pattern,
			},
			want: want{
				parser: &Parser{
					Patterns: []*regexp.Regexp{pattern},
				},
				err: nil,
			},
		},
		{
			name: "caputure group not found",
			args: args{
				pattern: capturedGroupNotContainsPattern,
			},
			want: want{
				parser: &Parser{
					Patterns: nil,
				},
				err: fmt.Errorf("invalid pattern detected: capture group not found"),
			},
		},
		{
			name: "non-named caputure group detected",
			args: args{
				pattern: nonNamedCapturedGroupContainsPattern,
			},
			want: want{
				parser: &Parser{
					Patterns: nil,
				},
				err: fmt.Errorf("invalid pattern detected: non-named capture group detected"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			err := p.AddPattern(tt.args.pattern)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Fatalf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(p.Patterns, tt.want.parser.Patterns) {
				t.Errorf("got: %v, want: %v", p.Patterns, tt.want.parser.Patterns)
			}
		})
	}
}

func TestAddPatterns(t *testing.T) {
	type args struct {
		patterns []*regexp.Regexp
	}
	type want struct {
		parser *Parser
		err    error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "success",
			args: args{
				patterns: patterns,
			},
			want: want{
				parser: &Parser{
					Patterns: patterns,
				},
				err: nil,
			},
		},
		{
			name: "caputure group not found",
			args: args{
				patterns: capturedGroupNotContainsPatterns,
			},
			want: want{
				parser: &Parser{
					Patterns: nil,
				},
				err: fmt.Errorf("invalid pattern detected: capture group not found"),
			},
		},
		{
			name: "non-named caputure group detected",
			args: args{
				patterns: nonNamedCapturedGroupContainsPatterns,
			},
			want: want{
				parser: &Parser{
					Patterns: nil,
				},
				err: fmt.Errorf("invalid pattern detected: non-named capture group detected"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			err := p.AddPatterns(tt.args.patterns)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Fatalf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(p.Patterns, tt.want.parser.Patterns) {
				t.Errorf("got: %v, want: %v", p.Patterns, tt.want.parser.Patterns)
			}
		})
	}
}

func TestParser_parse(t *testing.T) {
	type parser struct {
		Fields          []string
		Patterns        []*regexp.Regexp
		LineHandler     LineHandler
		MetadataHandler MetadataHandler
	}
	type args struct {
		input     io.Reader
		skipLines []int
	}
	type want struct {
		data     []string
		metadata *Metadata
		err      error
	}
	tests := []struct {
		name   string
		parser parser
		args   args
		want   want
	}{
		{
			name: "all match",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(allMatchInput),
				skipLines: nil,
			},
			want: want{
				data:     allMatchData,
				metadata: allMatchMetadata,
				err:      nil,
			},
		},
		{
			name: "contains unmatch",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(containsUnmatchInput),
				skipLines: nil,
			},
			want: want{
				data:     containsUnmatchData,
				metadata: containsUnmatchMetadata,
				err:      nil,
			},
		},
		{
			name: "contains skip flag",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(allMatchInput),
				skipLines: containsSkipLines,
			},
			want: want{
				data:     containsSkipData,
				metadata: containsSkipMetadata,
				err:      nil,
			},
		},
		{
			name: "all unmatch",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(allUnmatchInput),
				skipLines: nil,
			},
			want: want{
				data:     nil,
				metadata: allUnmatchMetadata,
				err:      nil,
			},
		},
		{
			name: "all skip",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(allMatchInput),
				skipLines: allSkipLines,
			},
			want: want{
				data:     nil,
				metadata: allSkipMetadata,
				err:      nil,
			},
		},
		{
			name: "mixed",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(containsUnmatchInput),
				skipLines: mixedSkipLines,
			},
			want: want{
				data:     mixedData,
				metadata: mixedMetadata,
				err:      nil,
			},
		},
		{
			name: "nil input",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(""),
				skipLines: nil,
			},
			want: want{
				data:     nil,
				metadata: emptyMetadata,
				err:      nil,
			},
		},
		{
			name: "line handler returns error",
			parser: parser{
				Patterns: patterns,
				LineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(allMatchInput),
				skipLines: nil,
			},
			want: want{
				data:     nil,
				metadata: nil,
				err:      fmt.Errorf("error"),
			},
		},
		{
			name: "metadata handler returns error",
			parser: parser{
				Patterns:    patterns,
				LineHandler: DefaultLineHandler,
				MetadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			args: args{
				input:     strings.NewReader(allMatchInput),
				skipLines: nil,
			},
			want: want{
				data:     allMatchData,
				metadata: allMatchMetadata,
				err:      nil, // This function does not invoke metadata handler directly
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				Patterns:        tt.parser.Patterns,
				LineHandler:     tt.parser.LineHandler,
				MetadataHandler: tt.parser.MetadataHandler,
			}
			data, metadata, err := p.parse(tt.args.input, tt.args.skipLines)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Errorf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(data, tt.want.data) {
				t.Errorf("got: %v, want: %v", data, tt.want.data)
			}
			if !reflect.DeepEqual(metadata, tt.want.metadata) {
				t.Errorf("got: %v, want: %v", metadata, tt.want.metadata)
			}
		})
	}
}

func TestParser_Parse(t *testing.T) {
	type parser struct {
		Fields          []string
		Patterns        []*regexp.Regexp
		LineHandler     LineHandler
		MetadataHandler MetadataHandler
	}
	type args struct {
		input     io.Reader
		skipLines []int
	}
	type want struct {
		got *Result
		err error
	}
	tests := []struct {
		name   string
		parser parser
		args   args
		want   want
	}{
		{
			name: "all match",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(allMatchInput),
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     allMatchData,
					Metadata: fmt.Sprintf(allMatchMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "contains unmatch",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(containsUnmatchInput),
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     containsUnmatchData,
					Metadata: fmt.Sprintf(containsUnmatchMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "contains skip flag",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(allMatchInput),
				skipLines: containsSkipLines,
			},
			want: want{
				got: &Result{
					Data:     containsSkipData,
					Metadata: fmt.Sprintf(containsSkipMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "all unmatch",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(allUnmatchInput),
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     nil,
					Metadata: fmt.Sprintf(allUnmatchMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "all skip",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(allMatchInput),
				skipLines: allSkipLines,
			},
			want: want{
				got: &Result{
					Data:     nil,
					Metadata: fmt.Sprintf(allSkipMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "mixed",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(containsUnmatchInput),
				skipLines: mixedSkipLines,
			},
			want: want{
				got: &Result{
					Data:     mixedData,
					Metadata: fmt.Sprintf(mixedMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "nil input",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(""),
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     nil,
					Metadata: fmt.Sprintf(emptyMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "line handler returns error",
			parser: parser{
				Patterns: patterns,
				LineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     strings.NewReader(allMatchInput),
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
		{
			name: "metadata handler returns error",
			parser: parser{
				Patterns:    patterns,
				LineHandler: DefaultLineHandler,
				MetadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			args: args{
				input:     strings.NewReader(allMatchInput),
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				Patterns:        tt.parser.Patterns,
				LineHandler:     tt.parser.LineHandler,
				MetadataHandler: tt.parser.MetadataHandler,
			}
			got, err := p.Parse(tt.args.input, tt.args.skipLines)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Errorf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(got, tt.want.got) {
				t.Errorf("got: %v, want: %v", got, tt.want.got)
			}
		})
	}
}

func TestParser_ParseString(t *testing.T) {
	type parser struct {
		Fields          []string
		Patterns        []*regexp.Regexp
		LineHandler     LineHandler
		MetadataHandler MetadataHandler
	}
	type args struct {
		input     string
		skipLines []int
	}
	type want struct {
		got *Result
		err error
	}
	tests := []struct {
		name   string
		parser parser
		args   args
		want   want
	}{
		{
			name: "all match",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     allMatchInput,
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     allMatchData,
					Metadata: fmt.Sprintf(allMatchMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "contains unmatch",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     containsUnmatchInput,
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     containsUnmatchData,
					Metadata: fmt.Sprintf(containsUnmatchMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "contains skip flag",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     allMatchInput,
				skipLines: containsSkipLines,
			},
			want: want{
				got: &Result{
					Data:     containsSkipData,
					Metadata: fmt.Sprintf(containsSkipMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "all unmatch",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     allUnmatchInput,
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     nil,
					Metadata: fmt.Sprintf(allUnmatchMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "all skip",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     allMatchInput,
				skipLines: allSkipLines,
			},
			want: want{
				got: &Result{
					Data:     nil,
					Metadata: fmt.Sprintf(allSkipMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "mixed",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     containsUnmatchInput,
				skipLines: mixedSkipLines,
			},
			want: want{
				got: &Result{
					Data:     mixedData,
					Metadata: fmt.Sprintf(mixedMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "nil input",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     "",
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     nil,
					Metadata: fmt.Sprintf(emptyMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "line handler returns error",
			parser: parser{
				Patterns: patterns,
				LineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     allMatchInput,
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
		{
			name: "metadata handler returns error",
			parser: parser{
				Patterns:    patterns,
				LineHandler: DefaultLineHandler,
				MetadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			args: args{
				input:     allMatchInput,
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				Patterns:        tt.parser.Patterns,
				LineHandler:     tt.parser.LineHandler,
				MetadataHandler: tt.parser.MetadataHandler,
			}
			got, err := p.ParseString(tt.args.input, tt.args.skipLines)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Errorf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(got, tt.want.got) {
				t.Errorf("got: %v, want: %v", got, tt.want.got)
			}
		})
	}
}

func TestParser_ParseFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		fileNotFoundMessage = fileNotFoundMessageWindows
		isDirMessage = isDirMessageWindows
	}
	type parser struct {
		Fields          []string
		Patterns        []*regexp.Regexp
		LineHandler     LineHandler
		MetadataHandler MetadataHandler
	}
	type args struct {
		input     string
		skipLines []int
	}
	type want struct {
		got *Result
		err error
	}
	tests := []struct {
		name   string
		parser parser
		args   args
		want   want
	}{
		{
			name: "all match",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     allMatchData,
					Metadata: fmt.Sprintf(allMatchMetadataSerialized, "sample_s3_all_match.log"),
				},
				err: nil,
			},
		},
		{
			name: "contains unmatch",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_contains_unmatch.log"),
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     containsUnmatchData,
					Metadata: fmt.Sprintf(containsUnmatchMetadataSerialized, "sample_s3_contains_unmatch.log"),
				},
				err: nil,
			},
		},
		{
			name: "contains skip flag",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines: containsSkipLines,
			},
			want: want{
				got: &Result{
					Data:     containsSkipData,
					Metadata: fmt.Sprintf(containsSkipMetadataSerialized, "sample_s3_all_match.log"),
				},
				err: nil,
			},
		},
		{
			name: "all unmatch",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_unmatch.log"),
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     nil,
					Metadata: fmt.Sprintf(allUnmatchMetadataSerialized, "sample_s3_all_unmatch.log"),
				},
				err: nil,
			},
		},
		{
			name: "all skip",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines: allSkipLines,
			},
			want: want{
				got: &Result{
					Data:     nil,
					Metadata: fmt.Sprintf(allSkipMetadataSerialized, "sample_s3_all_match.log"),
				},
				err: nil,
			},
		},
		{
			name: "mixed",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_contains_unmatch.log"),
				skipLines: mixedSkipLines,
			},
			want: want{
				got: &Result{
					Data:     mixedData,
					Metadata: fmt.Sprintf(mixedMetadataSerialized, "sample_s3_contains_unmatch.log"),
				},
				err: nil,
			},
		},
		{
			name: "nil input",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     "",
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("empty path detected"),
			},
		},
		{
			name: "line handler returns error",
			parser: parser{
				Patterns: patterns,
				LineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
		{
			name: "metadata handler returns error",
			parser: parser{
				Patterns:    patterns,
				LineHandler: DefaultLineHandler,
				MetadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
		{
			name: "input file does not exists",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log.dummy"),
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("cannot open file: open %s: %s", filepath.Join("testdata", "sample_s3_all_match.log.dummy"), fileNotFoundMessage),
			},
		},
		{
			name: "input path is directory not file",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     "testdata",
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("cannot read stream: read testdata: %s", isDirMessage),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				Patterns:        tt.parser.Patterns,
				LineHandler:     tt.parser.LineHandler,
				MetadataHandler: tt.parser.MetadataHandler,
			}
			got, err := p.ParseFile(tt.args.input, tt.args.skipLines)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Errorf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(got, tt.want.got) {
				t.Errorf("got: %v, want: %v", got, tt.want.got)
			}
		})
	}
}

func TestParser_ParseGzip(t *testing.T) {
	if runtime.GOOS == "windows" {
		fileNotFoundMessage = fileNotFoundMessageWindows
		isDirMessage = isDirMessageWindows
	}
	type parser struct {
		Fields          []string
		Patterns        []*regexp.Regexp
		LineHandler     LineHandler
		MetadataHandler MetadataHandler
	}
	type args struct {
		input     string
		skipLines []int
	}
	type want struct {
		got *Result
		err error
	}
	tests := []struct {
		name   string
		parser parser
		args   args
		want   want
	}{
		{
			name: "all match",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     allMatchData,
					Metadata: fmt.Sprintf(allMatchMetadataSerialized, "sample_s3_all_match.log.gz"),
				},
				err: nil,
			},
		},
		{
			name: "contains unmatch",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_contains_unmatch.log.gz"),
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     containsUnmatchData,
					Metadata: fmt.Sprintf(containsUnmatchMetadataSerialized, "sample_s3_contains_unmatch.log.gz"),
				},
				err: nil,
			},
		},
		{
			name: "contains skip flag",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines: containsSkipLines,
			},
			want: want{
				got: &Result{
					Data:     containsSkipData,
					Metadata: fmt.Sprintf(containsSkipMetadataSerialized, "sample_s3_all_match.log.gz"),
				},
				err: nil,
			},
		},
		{
			name: "all unmatch",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_unmatch.log.gz"),
				skipLines: nil,
			},
			want: want{
				got: &Result{
					Data:     nil,
					Metadata: fmt.Sprintf(allUnmatchMetadataSerialized, "sample_s3_all_unmatch.log.gz"),
				},
				err: nil,
			},
		},
		{
			name: "all skip",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines: allSkipLines,
			},
			want: want{
				got: &Result{
					Data:     nil,
					Metadata: fmt.Sprintf(allSkipMetadataSerialized, "sample_s3_all_match.log.gz"),
				},
				err: nil,
			},
		},
		{
			name: "mixed",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_contains_unmatch.log.gz"),
				skipLines: mixedSkipLines,
			},
			want: want{
				got: &Result{
					Data:     mixedData,
					Metadata: fmt.Sprintf(mixedMetadataSerialized, "sample_s3_contains_unmatch.log.gz"),
				},
				err: nil,
			},
		},
		{
			name: "nil input",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     "",
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("empty path detected"),
			},
		},
		{
			name: "line handler returns error",
			parser: parser{
				Patterns: patterns,
				LineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
		{
			name: "metadata handler returns error",
			parser: parser{
				Patterns:    patterns,
				LineHandler: DefaultLineHandler,
				MetadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
		{
			name: "input file does not exists",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log.gz.dummy"),
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("cannot open file: open %s: %s", filepath.Join("testdata", "sample_s3_all_match.log.gz.dummy"), fileNotFoundMessage),
			},
		},
		{
			name: "input path is directory not file",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     "testdata",
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("cannot create gzip reader for testdata: read testdata: %s", isDirMessage),
			},
		},
		{
			name: "input file is not gzip",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log"),
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("cannot create gzip reader for %s: gzip: invalid header", filepath.Join("testdata", "sample_s3_all_match.log")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				Patterns:        tt.parser.Patterns,
				LineHandler:     tt.parser.LineHandler,
				MetadataHandler: tt.parser.MetadataHandler,
			}
			got, err := p.ParseGzip(tt.args.input, tt.args.skipLines)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Errorf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(got, tt.want.got) {
				t.Errorf("got: %v, want: %v", got, tt.want.got)
			}
		})
	}
}

func TestParser_ParseZipEntries(t *testing.T) {
	if runtime.GOOS == "windows" {
		fileNotFoundMessage = fileNotFoundMessageWindows
		isDirMessage = isDirMessageWindows
	}
	type parser struct {
		Fields          []string
		Patterns        []*regexp.Regexp
		LineHandler     LineHandler
		MetadataHandler MetadataHandler
	}
	type args struct {
		input       string
		skipLines   []int
		globPattern string
	}
	type want struct {
		got []*Result
		err error
	}
	tests := []struct {
		name   string
		parser parser
		args   args
		want   want
	}{
		{
			name: "all match/one entry",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:   nil,
				globPattern: "*",
			},
			want: want{
				got: []*Result{
					{
						Data:     allMatchData,
						Metadata: fmt.Sprintf(allMatchMetadataSerialized, "sample_s3_all_match.log"),
					},
				},
				err: nil,
			},
		},
		{
			name: "contains unmatch/one entry",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "sample_s3_contains_unmatch.log.zip"),
				skipLines:   nil,
				globPattern: "*",
			},
			want: want{
				got: []*Result{
					{
						Data:     containsUnmatchData,
						Metadata: fmt.Sprintf(containsUnmatchMetadataSerialized, "sample_s3_contains_unmatch.log"),
					},
				},
				err: nil,
			},
		},
		{
			name: "contains skip flag/one entry",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:   allSkipLines,
				globPattern: "*",
			},
			want: want{
				got: []*Result{
					{
						Data:     nil,
						Metadata: fmt.Sprintf(allSkipMetadataSerialized, "sample_s3_all_match.log"),
					},
				},
				err: nil,
			},
		},
		{
			name: "all unmatch/one entry",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_unmatch.log.zip"),
				skipLines:   nil,
				globPattern: "*",
			},
			want: want{
				got: []*Result{
					{
						Data:     nil,
						Metadata: fmt.Sprintf(allUnmatchMetadataSerialized, "sample_s3_all_unmatch.log"),
					},
				},
				err: nil,
			},
		},
		{
			name: "all skip/one entry",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:   allSkipLines,
				globPattern: "*",
			},
			want: want{
				got: []*Result{
					{
						Data:     nil,
						Metadata: fmt.Sprintf(allSkipMetadataSerialized, "sample_s3_all_match.log"),
					},
				},
				err: nil,
			},
		},
		{
			name: "mixed",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "sample_s3_contains_unmatch.log.zip"),
				skipLines:   mixedSkipLines,
				globPattern: "*",
			},
			want: want{
				got: []*Result{
					{
						Data:     mixedData,
						Metadata: fmt.Sprintf(mixedMetadataSerialized, "sample_s3_contains_unmatch.log"),
					},
				},
				err: nil,
			},
		},
		{
			name: "nil input",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       "",
				skipLines:   nil,
				globPattern: "*",
			},
			want: want{
				got: nil,
				err: fmt.Errorf("empty path detected"),
			},
		},
		{
			name: "line handler returns error",
			parser: parser{
				Patterns: patterns,
				LineHandler: func(matches []string, fields []string, index int) (string, error) {
					return "", fmt.Errorf("error")
				},
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:   nil,
				globPattern: "*",
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
		{
			name: "metadata handler returns error",
			parser: parser{
				Patterns:    patterns,
				LineHandler: DefaultLineHandler,
				MetadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log.zip"),
				skipLines:   nil,
				globPattern: "*",
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
		{
			name: "input file does not exists",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:     filepath.Join("testdata", "sample_s3_all_match.log.zip.dummy"),
				skipLines: nil,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("cannot open zip file: open %s: %s", filepath.Join("testdata", "sample_s3_all_match.log.zip.dummy"), fileNotFoundMessage),
			},
		},
		{
			name: "input path is directory not file",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       "testdata",
				skipLines:   nil,
				globPattern: "*",
			},
			want: want{
				got: nil,
				err: fmt.Errorf("cannot open zip file: read testdata: %s", isDirMessage),
			},
		},
		{
			name: "input file is not zip",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "sample_s3_all_match.log.gz"),
				skipLines:   nil,
				globPattern: "*",
			},
			want: want{
				got: nil,
				err: fmt.Errorf("cannot open zip file: zip: not a valid zip file"),
			},
		},
		{
			name: "multi entries",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "log.zip"),
				skipLines:   nil,
				globPattern: "*",
			},
			want: want{
				got: []*Result{
					{
						Data:     allMatchData,
						Metadata: fmt.Sprintf(allMatchMetadataSerialized, "sample_s3_all_match.log"),
					},
					{
						Data:     containsUnmatchData,
						Metadata: fmt.Sprintf(containsUnmatchMetadataSerialized, "sample_s3_contains_unmatch.log"),
					},
					{
						Data:     nil,
						Metadata: fmt.Sprintf(allUnmatchMetadataSerialized, "sample_s3_all_unmatch.log"),
					},
				},
				err: nil,
			},
		},
		{
			name: "multi entries and glob pattern filtering",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "log.zip"),
				skipLines:   nil,
				globPattern: "*all_match*",
			},
			want: want{
				got: []*Result{
					{
						Data:     allMatchData,
						Metadata: fmt.Sprintf(allMatchMetadataSerialized, "sample_s3_all_match.log"),
					},
				},
				err: nil,
			},
		},
		{
			name: "multi entries and glob pattern returns error",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				input:       filepath.Join("testdata", "log.zip"),
				skipLines:   nil,
				globPattern: "[",
			},
			want: want{
				got: nil,
				err: fmt.Errorf("invalid glob pattern: syntax error in pattern"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				Patterns:        tt.parser.Patterns,
				LineHandler:     tt.parser.LineHandler,
				MetadataHandler: tt.parser.MetadataHandler,
			}
			got, err := p.ParseZipEntries(tt.args.input, tt.args.skipLines, tt.args.globPattern)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Errorf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(got, tt.want.got) {
				t.Errorf("got: %v, want: %v", got, tt.want.got)
			}
		})
	}
}

func TestParser_createResult(t *testing.T) {
	type parser struct {
		Fields          []string
		Patterns        []*regexp.Regexp
		LineHandler     LineHandler
		MetadataHandler MetadataHandler
	}
	type args struct {
		data     []string
		metadata *Metadata
	}
	type want struct {
		got *Result
		err error
	}
	tests := []struct {
		name   string
		parser parser
		args   args
		want   want
	}{
		{
			name: "basic",
			parser: parser{
				Patterns:        patterns,
				LineHandler:     DefaultLineHandler,
				MetadataHandler: DefaultMetadataHandler,
			},
			args: args{
				data:     allMatchData,
				metadata: allMatchMetadata,
			},
			want: want{
				got: &Result{
					Data:     allMatchData,
					Metadata: fmt.Sprintf(allMatchMetadataSerialized, ""),
				},
				err: nil,
			},
		},
		{
			name: "metadata handler returns error",
			parser: parser{
				Patterns:    patterns,
				LineHandler: DefaultLineHandler,
				MetadataHandler: func(metadata *Metadata) (string, error) {
					return "", fmt.Errorf("error")
				},
			},
			args: args{
				data:     allMatchData,
				metadata: allMatchMetadata,
			},
			want: want{
				got: nil,
				err: fmt.Errorf("error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Parser{
				Patterns:        tt.parser.Patterns,
				LineHandler:     tt.parser.LineHandler,
				MetadataHandler: tt.parser.MetadataHandler,
			}
			got, err := p.createResult(tt.args.data, tt.args.metadata)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Errorf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(got, tt.want.got) {
				t.Errorf("got: %v, want: %v", got, tt.want.got)
			}
		})
	}
}

func TestDefaultLineHandler(t *testing.T) {
	type args struct {
		matches []string
		fields  []string
		index   int
	}
	type want struct {
		got string
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "basic",
			args: args{
				matches: []string{"value1", "value2"},
				fields:  []string{"field1", "field2"},
				index:   1,
			},
			want: want{
				got: `{"index":1,"field1":"value1","field2":"value2"}`,
				err: nil,
			},
		},
		{
			name: "invalid json character",
			args: args{
				matches: []string{"value1", "val\"ue2"},
				fields:  []string{"field1", "field2"},
				index:   2,
			},
			want: want{
				got: `{"index":2,"field1":"value1","field2":"val\"ue2"}`,
				err: nil,
			},
		},
		{
			name: "more matches than fields",
			args: args{
				matches: []string{"value1", "value2", "value3"},
				fields:  []string{"field1", "field2"},
				index:   3,
			},
			want: want{
				got: `{"index":3,"field1":"value1","field2":"value2"}`,
				err: nil,
			},
		},
		{
			name: "more fields than matches",
			args: args{
				matches: []string{"value1"},
				fields:  []string{"field1", "field2"},
				index:   4,
			},
			want: want{
				got: `{"index":4,"field1":"value1"}`,
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultLineHandler(tt.args.matches, tt.args.fields, tt.args.index)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Errorf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(got, tt.want.got) {
				t.Errorf("got: %v, want: %v", got, tt.want.got)
			}
		})
	}
}

func TestDefaultMetadataHandler(t *testing.T) {
	type args struct {
		m *Metadata
	}
	type want struct {
		got string
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "all match",
			args: args{
				m: allMatchMetadata,
			},
			want: want{
				got: fmt.Sprintf(allMatchMetadataSerialized, ""),
				err: nil,
			},
		},
		{
			name: "contains unmatch",
			args: args{
				m: containsUnmatchMetadata,
			},
			want: want{
				got: fmt.Sprintf(containsUnmatchMetadataSerialized, ""),
				err: nil,
			},
		},
		{
			name: "contains skip",
			args: args{
				m: containsSkipMetadata,
			},
			want: want{
				got: fmt.Sprintf(containsSkipMetadataSerialized, ""),
				err: nil,
			},
		},
		{
			name: "all unmatch",
			args: args{
				m: allUnmatchMetadata,
			},
			want: want{
				got: fmt.Sprintf(allUnmatchMetadataSerialized, ""),
				err: nil,
			},
		},
		{
			name: "all skip",
			args: args{
				m: allSkipMetadata,
			},
			want: want{
				got: fmt.Sprintf(allSkipMetadataSerialized, ""),
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultMetadataHandler(tt.args.m)
			if err != nil && err.Error() != tt.want.err.Error() {
				t.Errorf("got: %v, want: %v", err.Error(), tt.want.err.Error())
			}
			if !reflect.DeepEqual(got, tt.want.got) {
				t.Errorf("got: %v, want: %v", got, tt.want.got)
			}
		})
	}
}
