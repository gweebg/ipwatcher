package watcher

import (
	"context"
	"fmt"
	"github.com/gweebg/ipwatcher/internal/config"
	"gopkg.in/gomail.v2"
	"log"
	"time"
)

// Recipient represents a email recipient defined as per the configuration file
type Recipient struct {
	// Name of the recipient
	Name string `mapstructure:"name"`
	// Address is the email address of the recipient
	Address string `mapstructure:"address"`
}

// Notifier allows for email mass notification
type Notifier struct {
	// From is the address to send from, obtained from the configuration file at 'watcher.smtp.*'
	From string
	// Recipients is the slice containing the recipients information such as email and name
	Recipients []Recipient

	// emailDialer represents the *gomail.Dialer object responsible by sending the email messages
	emailDialer *gomail.Dialer

	// doneCh is a channel that indicates when the email set is sent
	doneCh chan struct{}
}

func NewNotifier() *Notifier {

	c := config.GetConfig()

	dialer := gomail.NewDialer(
		c.GetString("watcher.smtp.smtp_server"),
		c.GetInt("watcher.smtp.smtp_port"),
		c.GetString("watcher.smtp.username"),
		c.GetString("watcher.smtp.password"),
	)

	var recipients []Recipient
	err := c.UnmarshalKey("watcher.smtp.recipients", &recipients)
	if err != nil {
		log.Fatalf("invalid 'watcher.smtp.recipients' configuration: %v\n", err.Error())
	}

	return &Notifier{
		From:        c.GetString("watcher.smtp.from_address"),
		Recipients:  recipients,
		emailDialer: dialer,
	}
}

func (n *Notifier) NotifyMail(ctx context.Context) error {

	s, err := n.emailDialer.Dial()
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	for _, r := range n.Recipients {

		ctx = context.WithValue(ctx, "name", r.Name)

		m.SetHeader("From", n.From)
		m.SetAddressHeader("To", r.Address, r.Name)
		m.SetHeader("Subject", "Update on your public address!")
		m.SetBody("text/html", generateMailBody(ctx))

		if err := gomail.Send(s, m); err != nil {
			log.Printf("could not send email to %q: %v", r.Address, err)
		}
		m.Reset()
	}

	err = s.Close()
	if err != nil {
		return err
	}

	return nil
}

type bodyGenerator func(context.Context) string

func generateMailBody(ctx context.Context) string {

	patterns := map[string]bodyGenerator{
		"on_change": generateOnChange,
		"on_match":  generateOnMatch,
		"on_error":  generateOnError,
	}

	event := ctx.Value("event").(string)
	generator, _ := patterns[event]

	return generator(ctx)
}

func generateOnChange(ctx context.Context) string {

	name := ctx.Value("name").(string)

	previousAddress := ctx.Value("previous_address").(string)
	currentAddress := ctx.Value("current_address").(string)

	timestamp := ctx.Value("timestamp").(time.Time)
	source := ctx.Value("source").(string)

	// todo: make email template dynamic by allowing its definition on the configuration file
	return fmt.Sprintf(`<html>
	<head>
		<title>Watcher Report</title>
	</head>
	<body style="font-family: Arial, sans-serif;">
		<div style="background-color: #f0f0f0; padding: 20px;">
			<h1 style="color: #333;">Your Public IP Address Has Updated</h1>
			<p style="font-size: 16px;">Hello <strong>%s</strong>, your public IP address has been changed. Here are the details:</p>
			<ul style="font-size: 16px;">
				<li><strong>Previous Address:</strong> %s</li>
				<li><strong>Current Address:</strong> %s</li>
				<li><strong>Updated at:</strong> %s</li>
				<li><strong>Information Source:</strong> %s</li>
			</ul>
		</div>
	</body>
	</html>`,
		name, previousAddress, currentAddress, timestamp.Format("2006-01-02 15:04:05"), source)

}

func generateOnMatch(ctx context.Context) string {

	name := ctx.Value("name").(string)
	timestamp := ctx.Value("timestamp").(time.Time)
	source := ctx.Value("source").(string)

	return fmt.Sprintf(`<html>
	<head>
		<title>Watcher Report</title>
	</head>
	<body style="font-family: Arial, sans-serif;">
		<div style="background-color: #f0f0f0; padding: 20px;">
			<h1 style="color: #333;">Your Public IP Address Has <strong>Not</strong> Changed</h1>
			<p style="font-size: 16px;">Hello <strong>%s</strong>, your public IP address is still the same. Here are the details:</p>
			<ul style="font-size: 16px;">
				<li><strong>At:</strong> %s</li>
				<li><strong>Information Source:</strong> %s</li>
			</ul>
		</div>
	</body>
	</html>`,
		name, timestamp.Format("2006-01-02 15:04:05"), source)

}

func generateOnError(ctx context.Context) string {

	name := ctx.Value("name").(string)

	timestamp := ctx.Value("timestamp").(time.Time)
	err := ctx.Value("error").(error)

	return fmt.Sprintf(`<html>
	<head>
		<title>Watcher Report</title>
	</head>
	<body style="font-family: Arial, sans-serif;">
		<div style="background-color: #f0f0f0; padding: 20px;">
			<h1 style="color: #333;">Watcher Error</h1>
			<p style="font-size: 16px;">Hello <strong>%s</strong>, an error occured while watching your address. Here are the details:</p>
			<ul style="font-size: 16px;">
				<li><strong>At:</strong> %s</li>
				<li><strong>Error:</strong> %s</li>
			</ul>
		</div>
	</body>
	</html>`,
		name, timestamp.Format("2006-01-02 15:04:05"), err.Error())

}
