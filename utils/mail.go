package utils

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net/mail"
	"net/smtp"
	"strconv"
)

type MailSender struct {
	smtpAddr string
	port     int
	mailUser string
	password string
}

var (
	mailContentType = "text/html; charset=UTF-8"
	crlf            = "\r\n"
)

func NewMailSender(smtpAddr string, port int, mailUser, password string) MailSender {
	return MailSender{smtpAddr: smtpAddr, port: port, mailUser: mailUser, password: password}
}

// SendEmail send email
func (ms MailSender) SendEmail(fromName, to, subject, body string) error {
	smtpAddr := ms.smtpAddr
	port := ms.port
	mailUer := ms.mailUser
	password := ms.password

	if err := ms.validateEmails(mailUer, to); err != nil {
		return errors.New("mail check error: " + err.Error())
	}

	addr := fmt.Sprintf("%s:%s", smtpAddr, strconv.Itoa(port))
	auth := smtp.PlainAuth("", mailUer, password, smtpAddr)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpAddr,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, smtpAddr)
	if err != nil {
		return err
	}
	defer c.Quit()

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(mailUer); err != nil {
		return err
	}

	if err = c.Rcpt(to); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	message := ms.buildMessage(fromName, to, subject, body)
	_, err = w.Write([]byte(message))
	return err
}

func (ms *MailSender) validateEmails(emails ...string) error {
	for _, email := range emails {
		if _, err := mail.ParseAddress(email); err != nil {
			return err
		}
	}
	return nil
}
func (ms *MailSender) buildMessage(fromName, to, subject, body string) string {
	var buf bytes.Buffer

	// 写入头部
	writeHeader(&buf, "From", fmt.Sprintf("%s <%s>", ms.encodeRFC2047(fromName), ms.mailUser))
	writeHeader(&buf, "To", to)
	writeHeader(&buf, "Subject", ms.encodeRFC2047(subject))
	writeHeader(&buf, "Content-Type", mailContentType)
	writeHeader(&buf, "MIME-Version", "1.0")

	// 添加空行分隔头部和正文
	buf.WriteString(crlf)

	// 写入正文
	buf.WriteString(body)

	return buf.String()
}

func (ms *MailSender) encodeRFC2047(s string) string {
	// 如果字符串只包含 ASCII 字符，则不需要编码
	if isASCII(s) {
		return s
	}
	return mime.QEncoding.Encode("utf-8", s)
}

func writeHeader(buf *bytes.Buffer, key, value string) {
	buf.WriteString(key)
	buf.WriteString(": ")
	buf.WriteString(value)
	buf.WriteString(crlf)
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}
