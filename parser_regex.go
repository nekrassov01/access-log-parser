package parser

import (
	"context"
	"fmt"
	"io"
	"regexp"
)

var _ Parser = (*RegexParser)(nil)

// RegexParser implements the Parser interface using regular expressions to parse log data.
// It allows customization of line handling as well as pattern matching.
type RegexParser struct {
	ctx      context.Context
	writer   io.Writer
	decoder  lineDecoder
	patterns []*regexp.Regexp
	opt      Option
}

// NewRegexParser initializes a new RegexParser with default handlers for line decoding, line handling.
// It's ready to use with additional pattern setup.
func NewRegexParser(ctx context.Context, writer io.Writer, opt Option) *RegexParser {
	p := &RegexParser{
		ctx:     ctx,
		writer:  writer,
		decoder: regexLineDecoder,
		opt:     opt,
	}
	if opt.LineHandler == nil {
		p.opt.LineHandler = JSONLineHandler
	}
	return p
}

// Parse processes log data from an io.Reader, applying configured patterns and handlers.
// It supports context cancellation, prefixing, and exclusion of lines.
func (p *RegexParser) Parse(reader io.Reader) (*Result, error) {
	return parse(p.ctx, reader, p.writer, p.patterns, p.decoder, p.opt)
}

// ParseString processes a single log string, applying skip lines and line number handling.
// It's a convenience method for quick string parsing with the configured parser instance.
func (p *RegexParser) ParseString(s string) (*Result, error) {
	return parseString(p.ctx, s, p.writer, p.patterns, p.decoder, p.opt)
}

// ParseFile processes log data from a file, applying skip lines and line number handling.
// It leverages the parser's configured patterns and handlers for file-based log parsing.
func (p *RegexParser) ParseFile(filePath string) (*Result, error) {
	return parseFile(p.ctx, filePath, p.writer, p.patterns, p.decoder, p.opt)
}

// ParseGzip processes gzip-compressed log data, applying skip lines and line number handling.
// It utilizes the parser's configurations for compressed log parsing.
func (p *RegexParser) ParseGzip(gzipPath string) (*Result, error) {
	return parseGzip(p.ctx, gzipPath, p.writer, p.patterns, p.decoder, p.opt)
}

// ParseZipEntries processes log data within zip archive entries, applying skip lines, line number handling,
// and glob pattern matching. It extends the parser's capabilities to zip-compressed logs.
func (p *RegexParser) ParseZipEntries(zipPath, globPattern string) (*Result, error) {
	return parseZipEntries(p.ctx, zipPath, globPattern, p.writer, p.patterns, p.decoder, p.opt)
}

// Patterns returns the list of regular expression patterns currently configured in the parser.
func (p *RegexParser) Patterns() []*regexp.Regexp {
	return p.patterns
}

// AddPattern adds a new regular expression pattern to the parser's pattern list.
// It validates the pattern to ensure it has named capture groups for structured parsing.
func (p *RegexParser) AddPattern(pattern string) error {
	ptn, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("%s: %w", regexPatternError, err)
	}
	if len(ptn.SubexpNames()) <= 1 {
		return fmt.Errorf("%s: capture group not found", regexPatternError)
	}
	for j, name := range ptn.SubexpNames() {
		if j != 0 && name == "" {
			return fmt.Errorf("%s: non-named capture group detected", regexPatternError)
		}
	}
	p.patterns = append(p.patterns, ptn)
	return nil
}

// AddPatterns adds multiple regular expression patterns to the parser's list.
// It leverages AddPattern for individual pattern validation and addition.
func (p *RegexParser) AddPatterns(patterns []string) error {
	for _, pattern := range patterns {
		if err := p.AddPattern(pattern); err != nil {
			return err
		}
	}
	return nil
}

// NewApacheCLFRegexParser initializes a new RegexParser specifically for parsing Apache Common Log Format (CLF) logs.
// It preconfigures the parser with regular expression patterns that match the Apache CLF log format.
func NewApacheCLFRegexParser(ctx context.Context, writer io.Writer, opt Option) *RegexParser {
	p := &RegexParser{
		ctx:     ctx,
		writer:  writer,
		decoder: regexLineDecoder,
		opt:     opt,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)"`),
			regexp.MustCompile(`^(?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-)`),
			regexp.MustCompile(`^(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)\t"(?P<referer>[^\"]*)"\t"(?P<user_agent>[^\"]*)"`),
			regexp.MustCompile(`^(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)`),
		},
	}
	if opt.LineHandler == nil {
		p.opt.LineHandler = JSONLineHandler
	}
	return p
}

// NewApacheCLFWithVHostRegexParser initializes a new RegexParser for parsing Apache logs with Virtual Host information.
// It extends the Apache CLF parser to include patterns that also capture the virtual host of each log entry.
func NewApacheCLFWithVHostRegexParser(ctx context.Context, writer io.Writer, opt Option) *RegexParser {
	p := &RegexParser{
		ctx:     ctx,
		writer:  writer,
		decoder: regexLineDecoder,
		opt:     opt,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<virtual_host>\S+) (?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)"`),
			regexp.MustCompile(`^(?P<virtual_host>\S+) (?P<remote_host>\S+) (?P<remote_logname>\S+) (?P<remote_user>[\S ]+) (?P<datetime>\[[^\]]+\]) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\" (?P<status>[0-9]{3}) (?P<size>[0-9]+|-)`),
			regexp.MustCompile(`^(?P<virtual_host>\S+)\t(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)\t"(?P<referer>[^\"]*)"\t"(?P<user_agent>[^\"]*)"`),
			regexp.MustCompile(`^(?P<virtual_host>\S+)\t(?P<remote_host>\S+)\t(?P<remote_logname>\S+)\t(?P<remote_user>[\S ]+)\t(?P<datetime>\[[^\]]+\])\t\"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\"\t(?P<status>[0-9]{3})\t(?P<size>[0-9]+|-)`),
		},
	}
	if opt.LineHandler == nil {
		p.opt.LineHandler = JSONLineHandler
	}
	return p
}

// NewS3RegexParser initializes a new RegexParser for parsing Amazon S3 access logs.
// It is preconfigured with patterns that match the S3 access log format, facilitating easy parsing of S3 logs.
func NewS3RegexParser(ctx context.Context, writer io.Writer, opt Option) *RegexParser {
	p := &RegexParser{
		ctx:     ctx,
		writer:  writer,
		decoder: regexLineDecoder,
		opt:     opt,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+) (?P<acl_required>[!-~]+)`),
			regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+) (?P<access_point_arn>[!-~]+)`),
			regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+) (?P<tls_version>[!-~]+)`),
			regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+) (?P<host_id>[!-~]+) (?P<signature_version>[!-~]+) (?P<cipher_suite>[!-~]+) (?P<authentication_type>[!-~]+) (?P<host_header>[!-~]+)`),
			regexp.MustCompile(`^(?P<bucket_owner>[!-~]+) (?P<bucket>[!-~]+) (?P<time>\[[^\]]+\]) (?P<remote_ip>[!-~]+) (?P<requester>[!-~]+) (?P<request_id>[!-~]+) (?P<operation>[!-~]+) (?P<key>[!-~]+) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\" (?P<http_status>\d{1,3}) (?P<error_code>[!-~]+) (?P<bytes_sent>[\d\-.]+) (?P<object_size>[\d\-.]+) (?P<total_time>[\d\-.]+) (?P<turn_around_time>[\d\-.]+) "(?P<referer>[^\"]*)" "(?P<user_agent>[^\"]*)" (?P<version_id>[!-~]+)`),
		},
	}
	if opt.LineHandler == nil {
		p.opt.LineHandler = JSONLineHandler
	}
	return p
}

// NewCFRegexParser initializes a new RegexParser for parsing Amazon CloudFront logs.
// It keywords patterns tailored to the CloudFront log format, simplifying the parsing of CloudFront access logs.
func NewCFRegexParser(ctx context.Context, writer io.Writer, opt Option) *RegexParser {
	p := &RegexParser{
		ctx:     ctx,
		writer:  writer,
		decoder: regexLineDecoder,
		opt:     opt,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<date>[\d\-.:]+)\t(?P<time>[\d\-.:]+)\t(?P<x_edge_location>[ -~]+)\t(?P<sc_bytes>[\d\-.]+)\t(?P<c_ip>[ -~]+)\t(?P<cs_method>[ -~]+)\t(?P<cs_host>[ -~]+)\t(?P<cs_uri_stem>[ -~]+)\t(?P<sc_status>\d{1,3}|-)\t(?P<cs_referer>[^\"]*)\t(?P<cs_user_agent>[^\"]*)\t(?P<cs_uri_query>[ -~]+)\t(?P<cs_cookie>\S+)\t(?P<x_edge_result_type>[ -~]+)\t(?P<x_edge_request_id>[ -~]+)\t(?P<x_host_header>[ -~]+)\t(?P<cs_protocol>[ -~]+)\t(?P<cs_bytes>[\d\-.]+)\t(?P<time_taken>[\d\-.]+)\t(?P<x_forwarded_for>[ -~]+)\t(?P<ssl_protocol>[ -~]+)\t(?P<ssl_cipher>[ -~]+)\t(?P<x_edge_response_result_type>[ -~]+)\t(?P<cs_protocol_version>[ -~]+)\t(?P<fle_status>[ -~]+)\t(?P<fle_encrypted_fields>\S+)\t(?P<c_port>[\d\-.]+)\t(?P<time_to_first_byte>[\d\-.]+)\t(?P<x_edge_detailed_result_type>[ -~]+)\t(?P<sc_content_type>[ -~]+)\t(?P<sc_content_len>[\d\-.]+)\t(?P<sc_range_start>[\d\-.]+)\t(?P<sc_range_end>[\d\-.]+)`),
		},
	}
	if opt.LineHandler == nil {
		p.opt.LineHandler = JSONLineHandler
	}
	return p
}

// NewALBRegexParser initializes a new RegexParser for parsing AWS Application Load Balancer (ALB) access logs.
// It comes preconfigured with patterns designed to parse ALB logs, making it easier to extract useful data from ALB logs.
func NewALBRegexParser(ctx context.Context, writer io.Writer, opt Option) *RegexParser {
	p := &RegexParser{
		ctx:     ctx,
		writer:  writer,
		decoder: regexLineDecoder,
		opt:     opt,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<type>[!-~]+) (?P<time>[!-~]+) (?P<elb>[!-~]+) (?P<client_port>[!-~]+) (?P<target_port>[!-~]+) (?P<request_processing_time>[\d\-.]+) (?P<target_processing_time>[\d\-.]+) (?P<response_processing_time>[\d\-.]+) (?P<elb_status_code>\d{1,3}|-) (?P<target_status_code>\d{1,3}|-) (?P<received_bytes>[\d\-.]+) (?P<sent_bytes>[\d\-.]+) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-|-)\" "(?P<user_agent>[^\"]*)" (?P<ssl_cipher>[!-~]+) (?P<ssl_protocol>[!-~]+) (?P<target_group_arn>[!-~]+) "(?P<trace_id>[ -~]+)" "(?P<domain_name>[ -~]+)" "(?P<chosen_cert_arn>[ -~]+)" (?P<matched_rule_priority>[!-~]+) (?P<request_creation_time>[!-~]+) "(?P<actions_executed>[ -~]+)" "(?P<redirect_url>[ -~]+)" "(?P<error_reason>[ -~]+)" "(?P<target_port_list>[ -~]+)" "(?P<target_status_code_list>[ -~]+)" "(?P<classification>[ -~]+)" "(?P<classification_reason>[ -~]+)"`),
		},
	}
	if opt.LineHandler == nil {
		p.opt.LineHandler = JSONLineHandler
	}
	return p
}

// NewNLBRegexParser initializes a new RegexParser for parsing AWS Network Load Balancer (NLB) access logs.
// This parser is equipped with patterns that are specifically designed for the NLB log format.
func NewNLBRegexParser(ctx context.Context, writer io.Writer, opt Option) *RegexParser {
	p := &RegexParser{
		ctx:     ctx,
		writer:  writer,
		decoder: regexLineDecoder,
		opt:     opt,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<type>[!-~]+) (?P<version>[!-~]+) (?P<time>[!-~]+) (?P<elb>[!-~]+) (?P<listener>[!-~]+) (?P<client_port>[!-~]+) (?P<destination_port>[!-~]+) (?P<connection_time>[\d\-.]+) (?P<tls_handshake_time>[\d\-.]+) (?P<received_bytes>[!-~]+) (?P<sent_bytes>[!-~]+) (?P<incoming_tls_alert>[!-~]+) (?P<chosen_cert_arn>[!-~]+) (?P<chosen_cert_serial>[ -~]+) (?P<tls_cipher>\S+) (?P<tls_protocol_version>[!-~]+) (?P<tls_named_group>[!-~]+) (?P<domain_name>[!-~]+) (?P<alpn_fe_protocol>[!-~]+) (?P<alpn_be_protocol>[!-~]+) (?P<alpn_client_preference_list>[ -~]+) (?P<tls_connection_creation_time>[!-~]+)`),
		},
	}
	if opt.LineHandler == nil {
		p.opt.LineHandler = JSONLineHandler
	}
	return p
}

// NewCLBRegexParser initializes a new RegexParser for parsing AWS Classic Load Balancer (CLB) access logs.
// It provides patterns that are tailored to the CLB log format, enabling efficient parsing of CLB logs.
func NewCLBRegexParser(ctx context.Context, writer io.Writer, opt Option) *RegexParser {
	p := &RegexParser{
		ctx:     ctx,
		writer:  writer,
		decoder: regexLineDecoder,
		opt:     opt,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`^(?P<time>[!-~]+) (?P<elb>[!-~]+) (?P<client_port>[!-~]+) (?P<backend_port>[!-~]+) (?P<request_processing_time>[\d\-.]+) (?P<backend_processing_time>[\d\-.]+) (?P<response_processing_time>[\d\-.]+) (?P<elb_status_code>\d{1,3}|-) (?P<backend_status_code>\d{1,3}|-) (?P<received_bytes>[\d\-.]+) (?P<sent_bytes>[\d\-.]+) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\" "(?P<user_agent>[^\"]*)" (?P<ssl_cipher>[!-~]+) (?P<ssl_protocol>[!-~]+)`),
			regexp.MustCompile(`^(?P<time>[!-~]+) (?P<elb>[!-~]+) (?P<client_port>[!-~]+) (?P<backend_port>[!-~]+) (?P<request_processing_time>[\d\-.]+) (?P<backend_processing_time>[\d\-.]+) (?P<response_processing_time>[\d\-.]+) (?P<elb_status_code>\d{1,3}|-) (?P<backend_status_code>\d{1,3}|-) (?P<received_bytes>[\d\-.]+) (?P<sent_bytes>[\d\-.]+) \"(?P<method>[A-Z\-]+) (?P<request_uri>[^ \"]+) (?P<protocol>HTTP/[0-9.]+|-)\"`),
		},
	}
	if opt.LineHandler == nil {
		p.opt.LineHandler = JSONLineHandler
	}
	return p
}
