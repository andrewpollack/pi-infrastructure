package meal_email

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"mime/quotedprintable"
	"net/textproto"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type EmailSender interface {
	SendEmail(subject, bodyHtml string, attachmentBytes []byte, attachmentFilename string) error
}

type SESEmailSender struct {
	From string
	To   []string
}

func (s SESEmailSender) SendEmail(subject, body string, attachmentBytes []byte, attachmentFilename string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}
	client := ses.NewFromConfig(cfg)

	// Process multiple recipients.
	toHeader := strings.Join(s.To, ", ")

	var emailRaw bytes.Buffer
	boundaryMixed := "NextPartMixedBoundary"
	boundaryAlternative := "NextPartAlternativeBoundary"

	// Headers
	emailRaw.WriteString(fmt.Sprintf("From: %s\r\n", s.From))
	emailRaw.WriteString(fmt.Sprintf("To: %s\r\n", toHeader))
	emailRaw.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	emailRaw.WriteString("MIME-Version: 1.0\r\n")
	emailRaw.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundaryMixed))
	emailRaw.WriteString("\r\n") // End headers

	// Start multipart/mixed section.
	emailRaw.WriteString(fmt.Sprintf("--%s\r\n", boundaryMixed))
	// Create multipart/alternative section for HTML content.
	emailRaw.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundaryAlternative))
	emailRaw.WriteString("\r\n")

	// Add the HTML part.
	emailRaw.WriteString(fmt.Sprintf("--%s\r\n", boundaryAlternative))
	emailRaw.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	emailRaw.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	emailRaw.WriteString("\r\n")
	qp := quotedprintable.NewWriter(&emailRaw)
	_, err = qp.Write([]byte(body))
	if err != nil {
		return fmt.Errorf("failed to write html body: %v", err)
	}
	defer func() {
		if err := qp.Close(); err != nil {
			fmt.Printf("error closing writer: %v\n", err)
		}
	}()
	emailRaw.WriteString("\r\n")
	// End alternative part.
	emailRaw.WriteString(fmt.Sprintf("--%s--\r\n", boundaryAlternative))

	// Add attachment part.
	emailRaw.WriteString(fmt.Sprintf("--%s\r\n", boundaryMixed))
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", "application/octet-stream")
	h.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachmentFilename))
	h.Set("Content-Transfer-Encoding", "base64")
	for key, vals := range h {
		for _, v := range vals {
			emailRaw.WriteString(fmt.Sprintf("%s: %s\r\n", key, v))
		}
	}
	emailRaw.WriteString("\r\n")
	encodedAttachment := base64.StdEncoding.EncodeToString(attachmentBytes)
	// RFC 2045 recommends splitting base64 into 76-character lines.
	for i := 0; i < len(encodedAttachment); i += 76 {
		end := i + 76
		if end > len(encodedAttachment) {
			end = len(encodedAttachment)
		}
		emailRaw.WriteString(encodedAttachment[i:end] + "\r\n")
	}
	emailRaw.WriteString(fmt.Sprintf("--%s--\r\n", boundaryMixed))

	// Prepare the raw email message.
	rawMessage := types.RawMessage{
		Data: emailRaw.Bytes(),
	}

	input := &ses.SendRawEmailInput{
		RawMessage: &rawMessage,
	}

	_, err = client.SendRawEmail(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to send raw email: %v", err)
	}

	log.Printf("📧 Email sent to %s.", toHeader)

	return nil
}
