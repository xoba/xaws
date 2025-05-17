package xaws

import (
	"net/mail"
	"strings"
	"testing"
)

func TestConvertAddrs(t *testing.T) {
	list := []mail.Address{{Address: "a@example.com"}, {Address: "b@test.com"}}
	got := convertAddrs(list)
	want := []string{"a@example.com", "b@test.com"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestSendEmailWithLengthDuplicateCID(t *testing.T) {
	es := CoreEmailSpec{
		Subject: "",
		From:    mail.Address{Address: "from@example.com"},
		To:      []mail.Address{{Address: "to@example.com"}},
		HTML:    "<p>hi</p>",
		Attachments: []Attachment{
			{Filename: "a.txt", Content: []byte("hi"), MimeType: "text/plain", ContentID: "1"},
			{Filename: "b.txt", Content: []byte("bye"), MimeType: "text/plain", ContentID: "1"},
		},
	}
	_, _, err := SendEmailWithLength(nil, es)
	if err == nil || !strings.Contains(err.Error(), "duplicate content ID") {
		t.Fatalf("expected duplicate content ID error, got %v", err)
	}
}
