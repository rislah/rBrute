package logger

type Status int

func (s Status) String() string {
	switch s {
	case BOT_RETRYING:
		return "BOT_RETRYING"
	case PROXY_RETRYING:
		return "PROXY_RETRYING"
	case BOT_SUCCESS:
		return "BOT_SUCCESS"
	case BOT_FAILED:
		return "BOT_FAILED"
	case VARIABLES_FOUND:
		return "VARIABLES_FOUND"
	case VARIABLES_MISSED:
		return "VARIABLES_MISSED"
	}
	return ""
}

const (
	BOT_RETRYING   Status = iota
	PROXY_RETRYING Status = iota
	BOT_SUCCESS
	BOT_FAILED
	VARIABLES_FOUND
	VARIABLES_MISSED
)
