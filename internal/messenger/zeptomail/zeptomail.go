package zeptomail

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/knadh/listmonk/models"
)

const (
	MessengerName  = "zeptomail"
	apiBaseURL     = "https://api.zeptomail.com/v1.1/email"
	authHeaderKey  = "Authorization"
	authScheme     = "Zoho-enczapikey"
	hdrContentType = "Content-Type"
)

// Config holds ZeptoMail API configuration.
type Config struct {
	Enabled       bool          `json:"enabled"`
	Name          string        `json:"name"`
	APIKey        string        `json:"api_key"`
	FromEmail     string        `json:"from_email"`
	FromName      string        `json:"from_name"`
	TrackOpens    bool          `json:"track_opens"`
	TrackClicks   bool          `json:"track_clicks"`
	Timeout       string `json:"timeout"`
	MaxMsgRetries int           `json:"max_msg_retries"`
}

// apiPayload is the JSON structure sent to ZeptoMail API.
type apiPayload struct {
	From        *address           `json:"from"`
	To          []*emailRecipient  `json:"to"`
	Cc          []*emailRecipient  `json:"cc,omitempty"`
	Bcc         []*emailRecipient  `json:"bcc,omitempty"`
	Subject     string             `json:"subject"`
	HTMLBody    string             `json:"htmlbody,omitempty"`
	TextBody    string             `json:"textbody,omitempty"`
	TrackOpens  bool               `json:"track_opens"`
	TrackClicks bool               `json:"track_clicks"`
	Attachments []*apiAttachment   `json:"attachments,omitempty"`
	ReplyTo     []*address         `json:"reply_to,omitempty"`
	Headers     []*apiHeader       `json:"headers,omitempty"`
}

type address struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}

type emailAddress struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}

type emailRecipient struct {
	EmailAddress emailAddress `json:"email_address"`
}

type apiAttachment struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type apiHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// apiError represents an error response from ZeptoMail.
type apiError struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// ZeptoMail is the ZeptoMail HTTP API messenger.
type ZeptoMail struct {
	conf Config
	c    *http.Client
}

// New returns a new ZeptoMail messenger.
func New(conf Config) (*ZeptoMail, error) {
	timeout := 30 * time.Second
	if conf.Timeout != "" {
		d, err := time.ParseDuration(conf.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout: %v", err)
		}
		timeout = d
	}
	if conf.Name == "" {
		conf.Name = MessengerName
	}

	return &ZeptoMail{
		conf: conf,
		c: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   10,
				MaxConnsPerHost:       10,
				ResponseHeaderTimeout: timeout,
				IdleConnTimeout:       90 * time.Second,
			},
		},
	}, nil
}

// Name returns the messenger name.
func (z *ZeptoMail) Name() string {
	return z.conf.Name
}

// Push sends an email through the ZeptoMail API.
func (z *ZeptoMail) Push(m models.Message) error {
	p := z.buildPayload(m)

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("zeptomail: error marshalling payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiBaseURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("zeptomail: error creating request: %w", err)
	}

	req.Header.Set(hdrContentType, "application/json")
	req.Header.Set(authHeaderKey, authScheme+" "+z.conf.APIKey)
	req.Header.Set("User-Agent", "listmonk")

	resp, err := z.c.Do(req)
	if err != nil {
		return fmt.Errorf("zeptomail: request failed: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		var ae apiError
		if err := json.Unmarshal(bodyBytes, &ae); err == nil && ae.Error.Message != "" {
			return fmt.Errorf("zeptomail: API error (HTTP %d): %s", resp.StatusCode, ae.Error.Message)
		}
		return fmt.Errorf("zeptomail: API error (HTTP %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Flush is a no-op for ZeptoMail.
func (z *ZeptoMail) Flush() error {
	return nil
}

// Close closes the ZeptoMail messenger's HTTP connections.
func (z *ZeptoMail) Close() error {
	z.c.CloseIdleConnections()
	return nil
}

// buildPayload converts a models.Message into a ZeptoMail API payload.
func (z *ZeptoMail) buildPayload(m models.Message) *apiPayload {
	p := &apiPayload{
		Subject:     m.Subject,
		TrackOpens:  z.conf.TrackOpens,
		TrackClicks: z.conf.TrackClicks,
	}
	p.From = z.parseAddress(m.From)

	fromAddr := p.From.Address

	to := make([]*emailRecipient, 0, len(m.To))
	for _, addr := range m.To {
		to = append(to, &emailRecipient{
			EmailAddress: emailAddress{Address: addr},
		})
	}
	p.To = to

	cc, bcc := z.extractEnvelopeHeaders(m.Headers)
	if len(cc) > 0 {
		p.Cc = cc
	}
	if len(bcc) > 0 {
		p.Bcc = bcc
	}

	if fromAddr != "" {
		p.ReplyTo = []*address{{Address: fromAddr}}
	}

	switch m.ContentType {
	case "plain":
		p.TextBody = string(m.Body)
	default:
		p.HTMLBody = string(m.Body)
		if len(m.AltBody) > 0 {
			p.TextBody = string(m.AltBody)
		}
	}

	if len(m.Attachments) > 0 {
		atts := make([]*apiAttachment, 0, len(m.Attachments))
		for _, a := range m.Attachments {
			atts = append(atts, &apiAttachment{
				Name:    a.Name,
				Content: base64.StdEncoding.EncodeToString(a.Content),
			})
		}
		p.Attachments = atts
	}

	seenHdrs := map[string]bool{}

	if len(m.Headers) > 0 {
		for k, vals := range m.Headers {
			lower := strings.ToLower(k)
			if lower == "bcc" || lower == "cc" || lower == "return-path" {
				continue
			}
			for _, v := range vals {
				if seenHdrs[lower] {
					continue
				}
				seenHdrs[lower] = true
				p.Headers = append(p.Headers, &apiHeader{
					Name:  k,
					Value: v,
				})
			}
		}
	}

	return p
}

// parseAddress parses an email address string (possibly with name) into an address struct.
func (z *ZeptoMail) parseAddress(addr string) *address {
	if addr == "" {
		a := &address{}
		if z.conf.FromEmail != "" {
			a.Address = z.conf.FromEmail
			a.Name = z.conf.FromName
		}
		return a
	}

	if a, err := mail.ParseAddress(addr); err == nil {
		return &address{
			Address: a.Address,
			Name:    a.Name,
		}
	}

	return &address{Address: addr}
}

// extractEnvelopeHeaders extracts Cc and Bcc recipients from MIME headers.
func (z *ZeptoMail) extractEnvelopeHeaders(hdr textproto.MIMEHeader) (cc, bcc []*emailRecipient) {
	if hdr == nil {
		return nil, nil
	}

	parse := func(val string) []*emailRecipient {
		if val == "" {
			return nil
		}
		var out []*emailRecipient
		for _, part := range strings.Split(val, ",") {
			addr := strings.TrimSpace(part)
			if addr != "" {
				out = append(out, &emailRecipient{
					EmailAddress: emailAddress{Address: addr},
				})
			}
		}
		return out
	}

	cc = parse(hdr.Get("Cc"))
	bcc = parse(hdr.Get("Bcc"))

	return cc, bcc
}
