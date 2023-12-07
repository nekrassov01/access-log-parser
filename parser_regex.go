package parser

import (
	"fmt"
	"io"
	"regexp"
)

// RegexParser is a parser that uses regular expressions to parse log entries.
// It enables the definition of complex parsing rules through regular expressions, making it versatile for different log formats.
// The parser allows for customization of log line processing and metadata handling through user-defined functions.
type RegexParser struct {
	decoder         decoder          // Internal parser function that utilizes regular expressions for parsing.
	lineHandler     LineHandler      // Custom function to process and format individual log lines after they are matched by a regex pattern.
	metadataHandler MetadataHandler  // Custom function to process and format metadata about the parsing process, such as total logs processed and errors.
	patterns        []*regexp.Regexp // Collection of regular expression patterns. Each pattern is used to match and extract data from log lines.
}

// NewRegexParser initializes and returns a new instance of RegexParser with default JSON handlers.
// This setup is suitable for cases where log data is expected to be output in a structured JSON format.
// Users can customize the parsing behavior by adding different regex patterns and by setting custom line and metadata handlers.
func NewRegexParser() *RegexParser {
	return &RegexParser{
		decoder:         regexDecoder,        // Internal parsing function based on regular expressions.
		lineHandler:     JSONLineHandler,     // Default line handler to convert matched log lines to JSON format.
		metadataHandler: JSONMetadataHandler, // Default metadata handler to convert parsing metadata to JSON format.
	}
}

// SetLineHandler sets a custom line handler for the parser.
func (p *RegexParser) SetLineHandler(handler LineHandler) {
	if handler == nil {
		return
	}
	p.lineHandler = handler
}

// SetMetadataHandler sets a custom metadata handler for the parser.
func (p *RegexParser) SetMetadataHandler(handler MetadataHandler) {
	if handler == nil {
		return
	}
	p.metadataHandler = handler
}

// Parse processes the given io.Reader input using the configured patterns and handlers.
func (p *RegexParser) Parse(input io.Reader, skipLines []int, hasIndex bool) (*Result, error) {
	return parse(input, skipLines, hasIndex, p.decoder, p.patterns, p.lineHandler, p.metadataHandler)
}

// ParseString processes the given string input as a reader.
func (p *RegexParser) ParseString(input string, skipLines []int, hasIndex bool) (*Result, error) {
	return parseString(input, skipLines, hasIndex, p.decoder, p.patterns, p.lineHandler, p.metadataHandler)
}

// ParseFile processes the content of the specified file.
func (p *RegexParser) ParseFile(input string, skipLines []int, hasIndex bool) (*Result, error) {
	return parseFile(input, skipLines, hasIndex, p.decoder, p.patterns, p.lineHandler, p.metadataHandler)
}

// ParseGzip processes the content of a gzipped file.
func (p *RegexParser) ParseGzip(input string, skipLines []int, hasIndex bool) (*Result, error) {
	return parseGzip(input, skipLines, hasIndex, p.decoder, p.patterns, p.lineHandler, p.metadataHandler)
}

// ParseZipEntries processes the contents of zip file entries matching the specified glob pattern.
func (p *RegexParser) ParseZipEntries(input string, skipLines []int, hasIndex bool, globPattern string) ([]*Result, error) {
	return parseZipEntries(input, skipLines, hasIndex, globPattern, p.decoder, p.patterns, p.lineHandler, p.metadataHandler)
}

// AddPattern adds a new regular expression pattern to the parser for matching log lines.
// Each pattern must contain at least one capture group and all capture groups should be named.
func (p *RegexParser) AddPattern(pattern *regexp.Regexp) error {
	if len(pattern.SubexpNames()) <= 1 {
		return fmt.Errorf("invalid pattern detected: capture group not found")
	}
	for j, name := range pattern.SubexpNames() {
		if j != 0 && name == "" {
			return fmt.Errorf("invalid pattern detected: non-named capture group detected")
		}
	}
	p.patterns = append(p.patterns, pattern)
	return nil
}

// AddPatterns adds multiple regular expression patterns to the parser.
func (p *RegexParser) AddPatterns(patterns []*regexp.Regexp) error {
	for _, pattern := range patterns {
		if err := p.AddPattern(pattern); err != nil {
			return err
		}
	}
	return nil
}

// GetPatterns gets regular expression patterns of the parser.
func (p *RegexParser) GetPatterns() []*regexp.Regexp {
	return p.patterns
}

// NewApacheCLFRegexParser creates a new RegexParser pre-configured for parsing Apache Common Log Format (CLF).
// This parser is suitable for logs typically found in Apache HTTP server access logs.
func NewApacheCLFRegexParser() *RegexParser {
	return &RegexParser{
		decoder:         regexDecoder,
		lineHandler:     JSONLineHandler,
		metadataHandler: JSONMetadataHandler,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)"`),
			regexp.MustCompile(`^(?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-)`),
			regexp.MustCompile(`^(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)\t"(?P<referer>[^\"]*)"\t"(?P<user_agent>[^\"]*)"`),
			regexp.MustCompile(`^(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)`),
		},
	}
}

// NewApacheCLFWithVHostRegexParser creates a new RegexParser configured for Apache CLF with an additional
// virtual host field. This parser is particularly useful for Apache logs that include the virtual host as part
// of the log entry, which is common in setups hosting multiple domains.
func NewApacheCLFWithVHostRegexParser() *RegexParser {
	return &RegexParser{
		decoder:         regexDecoder,
		lineHandler:     JSONLineHandler,
		metadataHandler: JSONMetadataHandler,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<virtual_host>\S+) (?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)"`),
			regexp.MustCompile(`^(?P<virtual_host>\S+) (?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-)`),
			regexp.MustCompile(`^(?P<virtual_host>\S+)\t(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)\t"(?P<referer>[^\"]*)"\t"(?P<user_agent>[^\"]*)"`),
			regexp.MustCompile(`^(?P<virtual_host>\S+)\t(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)`),
		},
	}
}

// NewS3RegexParser creates a new RegexParser for parsing Amazon S3 access logs.
// It is configured with patterns tailored to the structure of S3 logs, making it suitable for extracting
// detailed information from S3 access logs like operation, requester, error code, etc.
func NewS3RegexParser() *RegexParser {
	return &RegexParser{
		decoder:         regexDecoder,
		lineHandler:     JSONLineHandler,
		metadataHandler: JSONMetadataHandler,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+) (?P<acl_required>[!-~]+)`),
			regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+)`),
			regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+)`),
			regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+)`),
			regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+)`),
		},
	}
}

// NewCFRegexParser creates a new RegexParser for parsing Amazon CloudFront logs.
// This parser is equipped with patterns to parse the CloudFront access logs, which are useful for
// analyzing viewer traffic and understanding content delivery metrics.
func NewCFRegexParser() *RegexParser {
	return &RegexParser{
		decoder:         regexDecoder,
		lineHandler:     JSONLineHandler,
		metadataHandler: JSONMetadataHandler,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<date>[\d\-.:]+)\t(?P<time>[\d\-.:]+)\t(?P<x_edge_location>[ -~]+)\t(?P<sc_bytes>[\d\-.]+)\t(?P<c_ip>[ -~]+)\t(?P<cs_method>[ -~]+)\t(?P<cs_host>[ -~]+)\t(?P<cs_uri_stem>[ -~]+)\t(?P<sc_status>\d{1,3}|-)\t(?P<cs_referer>[^\"]*)\t(?P<cs_user_agent>[^\"]*)\t(?P<cs_uri_query>[ -~]+)\t(?P<cs_cookie>\S+)\t(?P<x_edge_result_type>[ -~]+)\t(?P<x_edge_request_id>[ -~]+)\t(?P<x_host_header>[ -~]+)\t(?P<cs_protocol>[ -~]+)\t(?P<cs_bytes>[\d\-.]+)\t(?P<time_taken>[\d\-.]+)\t(?P<x_forwarded_for>[ -~]+)\t(?P<ssl_protocol>[ -~]+)\t(?P<ssl_cipher>[ -~]+)\t(?P<x_edge_response_result_type>[ -~]+)\t(?P<cs_protocol_version>[ -~]+)\t(?P<fle_status>[ -~]+)\t(?P<fle_encrypted_fields>\S+)\t(?P<c_port>[\d\-.]+)\t(?P<time_to_first_byte>[\d\-.]+)\t(?P<x_edge_detailed_result_type>[ -~]+)\t(?P<sc_content_type>[ -~]+)\t(?P<sc_content_len>[\d\-.]+)\t(?P<sc_range_start>[\d\-.]+)\t(?P<sc_range_end>[\d\-.]+)`),
		},
	}
}

// NewALBRegexParser creates a new RegexParser for parsing AWS Application Load Balancer (ALB) access logs.
// The parser's patterns are designed to handle the log format generated by ALB, allowing for detailed
// analysis of client requests and backend responses.
func NewALBRegexParser() *RegexParser {
	return &RegexParser{
		decoder:         regexDecoder,
		lineHandler:     JSONLineHandler,
		metadataHandler: JSONMetadataHandler,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<type>[!-~]+) (?P<time>[!-~]+) (?P<elb>[!-~]+) (?P<client_port>[!-~]+) (?P<target_port>[!-~]+) (?P<request_processing_time>[\d\-.]+) (?P<target_processing_time>[\d\-.]+) (?P<response_processing_time>[\d\-.]+) (?P<elb_status_code>\d{1,3}|-) (?P<target_status_code>\d{1,3}|-) (?P<received_bytes>[\d\-.]+) (?P<sent_bytes>[\d\-.]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" "(?P<user_agent>[^\"]*)" (?P<ssl_cipher>[!-~]+) (?P<ssl_protocol>[!-~]+) (?P<target_group_arn>[!-~]+) "(?P<trace_id>[ -~]+)" "(?P<domain_name>[ -~]+)" "(?P<chosen_cert_arn>[ -~]+)" (?P<matched_rule_priority>[!-~]+) (?P<request_creation_time>[!-~]+) "(?P<actions_executed>[ -~]+)" "(?P<redirect_url>[ -~]+)" "(?P<error_reason>[ -~]+)" "(?P<target_port_list>[ -~]+)" "(?P<target_status_code_list>[ -~]+)" "(?P<classification>[ -~]+)" "(?P<classification_reason>[ -~]+)"`),
		},
	}
}

// NewNLBRegexParser creates a new RegexParser for parsing AWS Network Load Balancer (NLB) access logs.
// It contains patterns that are specific to the NLB log format, aiding in the processing and analysis of
// network traffic flowing through NLB instances.
func NewNLBRegexParser() *RegexParser {
	return &RegexParser{
		decoder:         regexDecoder,
		lineHandler:     JSONLineHandler,
		metadataHandler: JSONMetadataHandler,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<type>[!-~]+) (?P<version>[!-~]+) (?P<time>[!-~]+) (?P<elb>[!-~]+) (?P<listener>[!-~]+) (?P<client_port>[!-~]+) (?P<destination_port>[!-~]+) (?P<connection_time>[\d\-.]+) (?P<tls_handshake_time>[\d\-.]+) (?P<received_bytes>[!-~]+) (?P<sent_bytes>[!-~]+) (?P<incoming_tls_alert>[!-~]+) (?P<chosen_cert_arn>[!-~]+) (?P<chosen_cert_serial>[ -~]+) (?P<tls_cipher>\S+) (?P<tls_protocol_version>[!-~]+) (?P<tls_named_group>[!-~]+) (?P<domain_name>[!-~]+) (?P<alpn_fe_protocol>[!-~]+) (?P<alpn_be_protocol>[!-~]+) (?P<alpn_client_preference_list>[ -~]+) (?P<tls_connection_creation_time>[!-~]+)`),
		},
	}
}

// NewCLBRegexParser creates a new RegexParser for parsing AWS Classic Load Balancer (CLB) access logs.
// This parser is tailored for the log format produced by CLB, focusing on key metrics such as request
// processing time, backend processing time, and response processing time.
func NewCLBRegexParser() *RegexParser {
	return &RegexParser{
		decoder:         regexDecoder,
		lineHandler:     JSONLineHandler,
		metadataHandler: JSONMetadataHandler,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<time>[!-~]+) (?P<elb>[!-~]+) (?P<client_port>[!-~]+) (?P<backend_port>[!-~]+) (?P<request_processing_time>[\d\-.]+) (?P<backend_processing_time>[\d\-.]+) (?P<response_processing_time>[\d\-.]+) (?P<elb_status_code>\d{1,3}|-) (?P<backend_status_code>\d{1,3}|-) (?P<received_bytes>[\d\-.]+) (?P<sent_bytes>[\d\-.]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\" "(?P<user_agent>[^\"]*)" (?P<ssl_cipher>[!-~]+) (?P<ssl_protocol>[!-~]+)`),
			regexp.MustCompile(`^(?P<time>[!-~]+) (?P<elb>[!-~]+) (?P<client_port>[!-~]+) (?P<backend_port>[!-~]+) (?P<request_processing_time>[\d\-.]+) (?P<backend_processing_time>[\d\-.]+) (?P<response_processing_time>[\d\-.]+) (?P<elb_status_code>\d{1,3}|-) (?P<backend_status_code>\d{1,3}|-) (?P<received_bytes>[\d\-.]+) (?P<sent_bytes>[\d\-.]+) \"(?P<method>[A-Z]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+)\"`),
		},
	}
}
