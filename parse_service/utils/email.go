package utils

import (
	"net/smtp"
	"os"
	"strings"
)

var (
	user = os.Getenv("score_query_email_name")
	password = os.Getenv("score_query_email_password")
	host = os.Getenv("score_query_email_host")
)

func SendToMail(subject, body, mailType, to string) error {
	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var contentType string
	if mailType == "html" {
		contentType = "Content-Type: text/" + mailType + "; charset=UTF-8"
	} else {
		contentType = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: " + user + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	sendTo := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, sendTo, msg)
	return err
}