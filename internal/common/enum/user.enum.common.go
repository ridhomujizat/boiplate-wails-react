package enum

type UserTypeEnum string

const (
	AGENT  UserTypeEnum = "agent"
	CLIENT UserTypeEnum = "client"
	BOT    UserTypeEnum = "bot"
)

func (e UserTypeEnum) ToString() string {
	switch e {
	case AGENT:
		return "agent"
	case CLIENT:
		return "general"
	case BOT:
		return "bot"
	default:
		return ""
	}
}

func (e UserTypeEnum) IsValid() bool {
	switch e {
	case AGENT, CLIENT, BOT:
		return true
	}
	return false
}
