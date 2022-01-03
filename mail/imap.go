package mail

import (
	goimap "github.com/emersion/go-imap"
	goimapClient "github.com/emersion/go-imap/client"
	gomail "github.com/emersion/go-message/mail"
)

var (
	messageSection = goimap.BodySectionName{Peek: false}
	fetchMessages  = []goimap.FetchItem{messageSection.FetchItem()}
)

func (c *Client) connectIMAP() error {
	var err error
	c.log.Debug("connecting to IMAP...")
	dsn := c.cfg.IMAPhost + ":" + c.cfg.IMAPPort
	c.imap, err = goimapClient.DialTLS(dsn, nil)
	if err != nil {
		c.log.Error("cannot dial IMAP server: %v", err)
		return err
	}
	if err = c.imap.Login(c.cfg.Login, c.cfg.Password); err != nil {
		c.log.Error("cannot login to IMAP server: %v", err)
		return err
	}

	c.log.Info("connected to IMAP server")
	return nil
}

func (c *Client) GetMessages() {
	c.log.Debug("loading emails...")
	mbox, err := c.imap.Select(c.cfg.Mailbox, false)
	if err != nil {
		c.log.Error("cannot open mailbox %s: %v", c.cfg.Mailbox, err)
		return
	}
	c.log.Debug("total emails: %d", mbox.Messages)
	if mbox.Messages == 0 {
		c.log.Debug("no messages in the mailbox %s", c.cfg.Mailbox)
		return
	}

	// size := mbox.Messages - mbox.UnseenSeqNum
	size := mbox.Messages
	seqset := &goimap.SeqSet{}
	seqset.AddRange(uint32(1), mbox.Messages)
	// seqset.AddRange(mbox.UnseenSeqNum, mbox.Messages)
	messages := make(chan *goimap.Message, size)
	go func() {
		err := c.imap.Fetch(seqset, fetchMessages, messages)
		if err != nil {
			c.log.Error("cannot fetch message: %v", err)
		}
	}()

	c.parseMessages(messages)
}

func (c *Client) parseMessages(messages chan *goimap.Message) {
	for message := range messages {
		// server didn't return a message
		if message == nil {
			continue
		}
		body := message.GetBody(&messageSection)
		// server didn't return a body of that message
		if body == nil {
			continue
		}

		reader, err := gomail.CreateReader(body)
		if err != nil {
			c.log.Error("cannot read an email: %v", err)
			continue
		}
		c.log.Debug("headers: %+v", reader.Header)
	}
}
