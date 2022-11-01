package notification

type Notification interface {
	SendMarkdown(title, message string)
	Send(message interface{}) error
}

func New(mode, addr, secret string) (Notification, error) {
	switch mode {
	case "dingtalk":
		return newDingTalk(addr, secret)
	default:
		return newDefault()
	}
}
