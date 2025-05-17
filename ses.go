package xaws

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/mail"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/jhillyerd/enmime/v2"
)

// CoreEmailSpec describes the contents of an email.
type CoreEmailSpec struct {
	Subject     string
	From        mail.Address
	ReplyTo     []mail.Address
	To          []mail.Address
	Cc          []mail.Address
	HTML        string
	InReplyTo   string
	References  []string
	Attachments []Attachment
}

// Attachment holds metadata and content for an email attachment.
type Attachment struct {
	Filename  string
	Content   []byte
	MimeType  string
	ContentID string // optional, will be inline if present
}

// SendEmail attempts to send an email and retries without attachments if it exceeds the size limit.
func SendEmail(svc *sesv2.Client, es CoreEmailSpec) (string, error) {
	s, n, err := SendEmailWithLength(svc, es)
	if err != nil {
		max := MaxEmailBytesSESV2
		if n > max {
			log.Printf("going to try re-sending email without attachments; got error %v", err)
			w := new(bytes.Buffer)
			fmt.Fprintf(w, "we encountered an error when first trying to send this email, due to its size, so we have removed the attachments and are trying again.\n\n")
			fmt.Fprintf(w, "here were the original attachment(s):\n\n")
			for i, a := range es.Attachments {
				fmt.Fprintf(w, "attachment #%d: filename %q, with content type %q, contained %d bytes.\n\n", i+1, a.Filename, a.MimeType, len(a.Content))
			}
			es.Attachments = []Attachment{
				{
					Filename: "error.txt",
					Content:  w.Bytes(),
					MimeType: "text/plain",
				},
			}
			return SendEmail(svc, es)
		}
		return s, err
	}
	return s, err
}

// SendEmailWithLength sends an email and reports the raw message size.
func SendEmailWithLength(svc *sesv2.Client, es CoreEmailSpec) (string, int, error) {
	master := enmime.Builder().
		Subject(es.Subject).
		HTML([]byte(es.HTML)).
		From(es.From.Name, es.From.Address)
	for _, to := range es.To {
		master = master.To(to.Name, to.Address)
	}
	for _, cc := range es.Cc {
		master = master.CC(cc.Name, cc.Address)
	}
	if irt := es.InReplyTo; len(irt) > 0 {
		master = master.Header("In-Reply-To", irt)
	}
	if refs := es.References; len(refs) > 0 {
		master = master.Header("References", strings.Join(refs, " "))
	}
	{
		uniques := make(map[string]bool)
		for _, a := range es.Attachments {
			if id := a.ContentID; len(id) > 0 {
				if uniques[id] {
					return "", 0, fmt.Errorf("duplicate content ID %q", id)
				}
				uniques[id] = true
				master = master.AddInline(a.Content, a.MimeType, a.Filename, id)
			} else {
				master = master.AddAttachment(a.Content, a.MimeType, a.Filename)
			}
		}
	}
	p, err := master.Build()
	if err != nil {
		return "", 0, err
	}
	rawData := new(bytes.Buffer)
	if err := p.Encode(rawData); err != nil {
		return "", 0, err
	}
	resp, err := svc.SendEmail(context.Background(), &sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Raw: &types.RawMessage{
				Data: rawData.Bytes(),
			},
		},
		Destination: &types.Destination{
			CcAddresses: convertAddrs(es.Cc),
			ToAddresses: convertAddrs(es.To),
		},
		FromEmailAddress: ptr(es.From.String()),
		ReplyToAddresses: convertAddrs(es.ReplyTo),
	})
	if err != nil {
		return "", rawData.Len(), err
	}
	return fmt.Sprintf("<%s@email.amazonses.com>", *resp.MessageId), rawData.Len(), nil
}

// ptr returns a pointer to the provided value.
func ptr[T any](t T) *T {
	return &t
}

// convertAddrs extracts the address strings from a slice of mail.Address.
func convertAddrs(list []mail.Address) []string {
	var out []string
	for _, x := range list {
		out = append(out, x.Address)
	}
	return out
}
