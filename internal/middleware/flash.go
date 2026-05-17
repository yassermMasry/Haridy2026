package middleware

import "github.com/gin-contrib/sessions"

const flashKey = "flash_message"

func SetFlash(session sessions.Session, message string) {
	session.Set(flashKey, message)
	_ = session.Save()
}

func PopFlash(session sessions.Session) string {
	value := session.Get(flashKey)
	if value == nil {
		return ""
	}
	session.Delete(flashKey)
	_ = session.Save()
	if msg, ok := value.(string); ok {
		return msg
	}
	return ""
}
