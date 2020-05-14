package common

type ResolveOptions struct {
	Connection string `form:"id"`
	User       string `form:"user"`
	Session    string `form:"session"`
	Channel    string `form:"channel"`
}
