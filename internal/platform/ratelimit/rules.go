package ratelimit

import "time"

type Rule struct {
	Key    string
	Limit  int
	Window time.Duration
}

var RuleAuth = Rule{
	Key:    "auth",
	Limit:  10,
	Window: time.Minute,
}

var RuleModify = Rule{
	Key:    "modify",
	Limit:  20,
	Window: time.Minute,
}

var RuleGet = Rule{
	Key:    "get",
	Limit:  100,
	Window: time.Minute,
}

var RuleRefresh = Rule{
	Key:    "refresh",
	Limit:  5,
	Window: time.Minute,
}

var RuleAdmin = Rule{
	Key:    "admin",
	Limit:  42,
	Window: time.Minute,
}
