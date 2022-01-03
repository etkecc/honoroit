package mail

import (
	goimapClient "github.com/emersion/go-imap/client"

	"gitlab.com/etke.cc/honoroit/logger"
)

// Config of the mail client
type Config struct {
	// IMAPhost of an email server
	IMAPhost string
	// IMAPPort of an email server
	IMAPPort string
	// SMTPhost of an email server
	SMTPhost string
	// SMTPport of an email server
	SMTPport string
	// Login of a user
	Login string
	// Password of a user
	Password string
	// Email of the logged in user
	Email string
	// Mailbox to check mail, usually INBOX
	Mailbox string
	// Sentbox is a mailbox for sent mail, usually Sent
	Sentbox string
	// LogLevel of the email logger
	LogLevel string
}

// Client object
type Client struct {
	cfg  *Config
	log  *logger.Logger
	imap *goimapClient.Client
}

// New mail client
func New(cfg *Config) *Client {
	client := &Client{
		cfg: cfg,
		log: logger.New("mail.", cfg.LogLevel),
	}

	return client
}

// Start email client
func (c *Client) Start() error {
	idlechan := make(chan struct{}, 1)
	defer close(idlechan)

	if err := c.connectIMAP(); err != nil {
		return err
	}

	c.GetMessages()
	c.log.Info("IMAP client initialized, enabling IDLE mode")
	return c.imap.Idle(idlechan, nil)
}

// Stop email client
func (c *Client) Stop() {
	if err := c.imap.Logout(); err != nil {
		c.log.Error("cannot logout IMAP: %v", err)
	}
}
