package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/nekrassov01/access-log-parser"
)

func main() {
	regexPattern := []*regexp.Regexp{
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+) (?P<acl_required>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+)`),
		regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+)`),
	}

	regexSample := `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" - s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket43.s3.us-west-1.amazonaws.com TLSV1.1 -
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-" "S3Console/0.4" - 9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com TLSV1.1
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 - "-" "S3Console/0.4" - BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4" - Ke1bUcazaN1jWuUlPJaxF64cQVpUEhoZKEG/hmy/gijN/I1DeWqDfFvnpybfEseEME/u7ME1234= SigV2 ECDHE-RSA-AES128-SHA AuthHeader
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" -`

	ltsvSample := `remote_host:192.168.1.1	remote_logname:-	remote_user:john	datetime:[12/Mar/2023:10:55:36 +0000]	request:GET /index.html HTTP/1.1	status:200	size:1024	referer:http://www.example.com/	user_agent:Mozilla/5.0 (Windows NT 10.0; Win64; x64)
remote_host:172.16.0.2	remote_logname:-	remote_user:jane	datetime:[12/Mar/2023:10:56:10 +0000]	request:POST /login HTTP/1.1	status:303	size:532	referer:http://www.example.com/login	user_agent:Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)
remote_host:10.0.0.3	remote_logname:-	remote_user:mike	datetime:[12/Mar/2023:10:57:15 +0000]	request:GET /about.html HTTP/1.1	status:200	size:749	referer:http://www.example.com/	user_agent:Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)
remote_host:192.168.1.4	remote_logname:-	remote_user:anna	datetime:[12/Mar/2023:10:58:24 +0000]	request:GET /products HTTP/1.1	status:404	size:0
remote_host:192.168.1.10	remote_logname:-	remote_user:chris	datetime:[12/Mar/2023:11:04:16 +0000]	request:DELETE /account HTTP/1.1	status:200	size:204`

	/* RegexParser */

	// json format
	rp := parser.NewRegexParser()
	if err := rp.AddPatterns(regexPattern); err != nil {
		log.Fatal(err)
	}
	out, err := rp.ParseString(regexSample, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)
	fmt.Println("")

	// pretty-json format
	rp.SetLineHandler(parser.PrettyJSONLineHandler)
	rp.SetMetadataHandler(parser.PrettyJSONMetadataHandler)
	out, err = rp.ParseString(regexSample, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)
	fmt.Println("")

	// key=value pair format
	rp.SetLineHandler(parser.KeyValuePairLineHandler)
	rp.SetMetadataHandler(parser.KeyValuePairMetadataHandler)
	out, err = rp.ParseString(regexSample, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)
	fmt.Println("")

	// ltsv format
	rp.SetLineHandler(parser.LTSVLineHandler)
	rp.SetMetadataHandler(parser.LTSVMetadataHandler)
	out, err = rp.Parse(strings.NewReader(regexSample), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)
	fmt.Println("")

	/* LTSVParser */

	// json format
	lp := parser.NewLTSVParser()
	out, err = lp.ParseString(ltsvSample, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)
	fmt.Println("")

	// pretty-json format
	lp.SetLineHandler(parser.PrettyJSONLineHandler)
	lp.SetMetadataHandler(parser.PrettyJSONMetadataHandler)
	out, err = lp.ParseString(ltsvSample, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)
	fmt.Println("")

	// key-value pair format
	lp.SetLineHandler(parser.KeyValuePairLineHandler)
	lp.SetMetadataHandler(parser.KeyValuePairMetadataHandler)
	out, err = lp.ParseString(ltsvSample, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)
	fmt.Println("")

	// ltsv format
	lp.SetLineHandler(parser.LTSVLineHandler)
	lp.SetMetadataHandler(parser.LTSVMetadataHandler)
	out, err = lp.ParseString(ltsvSample, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println(out.Metadata)
	fmt.Println("")
}
