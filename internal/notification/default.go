package notification

import "fmt"

type console struct {
}

func newDefault() (Notification, error) {
	return &console{}, nil
}

func (c *console) SendMarkdown(title, message string) {
	fmt.Println(title)
	fmt.Println(message)
}

func (c *console) Send(message interface{}) error {
	fmt.Println(message)
	return nil
}
