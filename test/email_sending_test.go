package test

import (
	"context"
	"currency-rates-notifier/internal/api/mailpit"
	"fmt"
	smtpmock "github.com/mocktools/go-smtp-mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/wneessen/go-mail"
	"net/smtp"
	"net/url"
	"runtime"
	"testing"
	"time"
)

func TestEmailSending(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	if runtime.GOOS != "linux" {
		t.Skip("Works only on Linux (Testcontainers)")
	}

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: "axllent/mailpit:v1.24",
		ExposedPorts: []string{
			"1025/tcp",
			"8025/tcp",
		},
		Env: map[string]string{
			"MP_SMTP_AUTH_ACCEPT_ANY":     "1",
			"MP_SMTP_AUTH_ALLOW_INSECURE": "1",
		},
		WaitingFor: wait.ForHTTP("/readyz").
			WithPort("8025").
			WithStartupTimeout(10 * time.Second),
	}

	ctrReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	ctr, err := testcontainers.GenericContainer(ctx, ctrReq)
	require.NoError(t, err)
	defer func() { require.NoError(t, ctr.Terminate(ctx)) }()

	host, err := ctr.Host(ctx)
	require.NoError(t, err)

	smtpPort, err := ctr.MappedPort(ctx, "1025/tcp")
	require.NoError(t, err)

	webPort, err := ctr.MappedPort(ctx, "8025/tcp")
	require.NoError(t, err)

	t.Logf("Mailpit started on %s ports: %d (smtp) %d (web)", host, smtpPort.Int(), webPort.Int())

	sender := "sender@example.com"
	recipients := []string{"recipients@example.com"}
	subject := "Subject: Hello! This is a test email"
	content := "This will be the content of the mail"

	message := mail.NewMsg()
	if err := message.From(sender); err != nil {
		t.Fatalf("failed recipients set FROM address: %s", err)
	}
	if err := message.To(recipients...); err != nil {
		t.Fatalf("failed recipients set TO address: %s", err)
	}
	message.Subject(subject)
	message.SetBodyString(mail.TypeTextPlain, content)

	client, err := mail.NewClient(host,
		mail.WithPort(smtpPort.Int()),
		mail.WithTLSPolicy(mail.NoTLS),
	)

	if err != nil {
		t.Fatalf("failed recipients create new mail delivery client: %s", err)
	}
	if err := client.DialAndSend(message); err != nil {
		t.Fatalf("failed recipients deliver mail: %s", err)
	}

	u := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", host, webPort.Int()),
	}

	mailpitClient := mailpit.NewClient(u.String())

	messages, err := mailpitClient.GetMessages()
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	t.Logf("Total messages: %d", len(messages))
	require.Len(t, messages, 1)

	if len(messages) > 0 {
		detail, err := mailpitClient.GetMessageDetail(messages[0].ID)
		if err != nil {
			t.Fatalf("Failed to get message detail: %v", err)
		}

		require.Equal(t, sender, detail.From.Address)
		require.Equal(t, recipients[0], detail.To[0].Address)
		require.Equal(t, subject, detail.Subject)
		require.Contains(t, detail.Text, content)
	}

	err = mailpitClient.DeleteAllMessages()
	if err != nil {
		t.Fatalf("Failed to delete all messages: %v", err)
	}
}

func TestEmailSendingSimpleClient(t *testing.T) {
	t.Skip("This test is for example")
	server := smtpmock.New(smtpmock.ConfigurationAttr{})
	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Failed to start SMTP mock server: %v", err)
			return
		}
	}()
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("Failed to stop SMTP mock server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	portNumber := server.PortNumber

	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	body := []byte("Subject: Hello!\r\n\r\nThis is a test email.")

	err := smtp.SendMail(fmt.Sprintf("localhost:%d", portNumber), nil, from, to, body)
	if err != nil {
		t.Fatalf("Failed to send email: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if len(server.Messages()) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(server.Messages()))
	}

	msg := server.Messages()[0]

	require.Contains(t, msg.MailfromRequest(), "sender@example.com")
	require.Contains(t, msg.RcpttoRequest(), "recipient@example.com")
	require.Contains(t, msg.MsgRequest(), "Subject: Hello!")
	require.Contains(t, msg.MsgRequest(), "This is a test email.")

	require.True(t, msg.Helo())
	require.True(t, msg.Mailfrom())
	require.True(t, msg.Rcptto())
	require.True(t, msg.Data())
	require.True(t, msg.Msg())
	require.True(t, msg.QuitSent())

	require.False(t, msg.Rset())
	//todo: check response codes
}

func TestEmailSendingGoMailClient(t *testing.T) {
	t.Skip("This test is for example")
	server := smtpmock.New(smtpmock.ConfigurationAttr{
		LogServerActivity: true,
		LogToStdout:       true,
	})

	go func() {
		if err := server.Start(); err != nil {
			t.Errorf("Failed recipient start SMTP mock server: %v", err)
			return
		}
	}()
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatalf("Failed recipient stop SMTP mock server: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	portNumber := server.PortNumber

	sender := "sender@example.com"
	recipient := []string{"recipient@example.com"}
	subject := "Subject: Hello! This is a test email"
	content := "This will be the content of the mail"

	message := mail.NewMsg()
	if err := message.From(sender); err != nil {
		t.Fatalf("failed recipient set FROM address: %s", err)
	}
	if err := message.To(recipient...); err != nil {
		t.Fatalf("failed recipient set TO address: %s", err)
	}
	message.Subject(subject)
	message.SetBodyString(mail.TypeTextPlain, content)

	client, err := mail.NewClient("127.0.0.1",
		mail.WithPort(portNumber),
		mail.WithSMTPAuth(mail.SMTPAuthNoAuth),
		mail.WithHELO("test.example.com"),
		mail.WithTLSPolicy(mail.NoTLS),
		mail.WithoutNoop(),
		mail.WithTimeout(5*time.Minute),
	)

	if err != nil {
		t.Fatalf("failed recipient create new mail delivery client: %s", err)
	}
	if err := client.DialAndSend(message); err != nil {
		t.Fatalf("failed recipient deliver mail: %s", err)
	}

	time.Sleep(300 * time.Millisecond)

	if len(server.Messages()) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(server.Messages()))
	}

	//impossible to make assertions as go-mail client sends RSET (SMTP request) which clears everything
}
