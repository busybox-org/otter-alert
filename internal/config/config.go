package config

import "time"

type Config struct {
	Interval     time.Duration
	Zookeeper    []string
	Notification *Notification
	Manager      *Manager
}

type Notification struct {
	Type   string
	Url    string
	Secret string
}

type Manager struct {
	DatabaseUrl string
	Endpoint    string
	Username    string
	Password    string
}

var App = &Config{
	Notification: new(Notification),
	Manager:      new(Manager),
}
