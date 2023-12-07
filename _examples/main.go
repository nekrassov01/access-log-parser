package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/nekrassov01/access-log-parser"
)

func main() {
	s3input := `a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a awsrandombucket43 [16/Feb/2019:11:23:45 +0000] 192.0.2.132 a19b12df90c456a18e96d34c56d23c56a78f0d89a45f6a78901b23c45d67ef8a 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket43?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" - s9lzHYrFp76ZVxRcpX9+5cjAnEH2ROuNkd2BHfIa6UkFVdtjf5mKR3/eTPFvsiP/XV/VLi31234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket43.s3.us-west-1.amazonaws.com TLSV1.1 -
3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 awsrandombucket59 [24/Feb/2019:07:45:11 +0000] 192.0.2.45 3b24c35d67a89f01b23c45d67890a12b345c67d89a0b12c3d45e67fa89b01c23 891CE47D2EXAMPLE REST.GET.LOGGING_STATUS - "GET /awsrandombucket59?logging HTTP/1.1" 200 - 242 - 11 - "-" "S3Console/0.4" - 9vKBE6vMhrNiWHZmb2L0mXOcqPGzQOI5XLnCtZNPxev+Hf+7tpT6sxDwDty4LHBUOZJG96N1234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com TLSV1.1
8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 awsrandombucket12 [12/Feb/2019:18:32:21 +0000] 192.0.2.189 8f90a1b23c45d67e89a01b23c45d6789f01a23b45c67890d12e34f56a78901b2 A1206F460EXAMPLE REST.GET.BUCKETPOLICY - "GET /awsrandombucket12?policy HTTP/1.1" 404 NoSuchBucketPolicy 297 - 38 - "-" "S3Console/0.4" - BNaBsXZQQDbssi6xMBdBU2sLt+Yf5kZDmeBUP35sFoKa3sLLeMC78iwEIWxs99CRUrbS4n11234= SigV2 ECDHE-RSA-AES128-GCM-SHA256 AuthHeader awsrandombucket59.s3.us-west-1.amazonaws.com
d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 awsrandombucket89 [03/Feb/2019:03:54:33 +0000] 192.0.2.76 d45e67fa89b012c3a45678901b234c56d78a90f12b3456789a012345c6789d01 7B4A0FABBEXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket89?versioning HTTP/1.1" 200 - 113 - 33 - "-" "S3Console/0.4" - Ke1bUcazaN1jWuUlPJaxF64cQVpUEhoZKEG/hmy/gijN/I1DeWqDfFvnpybfEseEME/u7ME1234= SigV2 ECDHE-RSA-AES128-SHA AuthHeader
01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f awsrandombucket77 [28/Feb/2019:14:12:59 +0000] 192.0.2.213 01b23c45d67890a12b345c6789d01a23b45c67d89012a34b5678c90d1234e56f 3E57427F3EXAMPLE REST.GET.VERSIONING - "GET /awsrandombucket77?versioning HTTP/1.1" 200 - 113 - 7 - "-" "S3Console/0.4" -`

	// instantiate parser for s3
	p := parser.NewS3RegexParser()

	// parse from string input
	out, err := p.ParseString(s3input, nil, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("-- Processed Data --")
	fmt.Println(strings.Join(out.Data, "\n"))
	fmt.Println("")

	fmt.Println("-- Processed Metadata --")
	fmt.Println(out.Metadata)
	fmt.Println("")

	fmt.Println("-- Extracted Labels --")
	for _, label := range out.Labels {
		fmt.Println(strings.Join(label, "\t"))
	}
	fmt.Println("")

	fmt.Println("-- Extracted Values --")
	for _, value := range out.Values {
		fmt.Println(strings.Join(value, "\t"))
	}
	fmt.Println("")
}
