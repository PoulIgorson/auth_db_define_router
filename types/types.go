package types

var notifies = map[uint][]*notify{}

type notify struct {
	typ    string
	Text   string
	UserId uint
}

func (notify *notify) Type() string {
	return notify.typ
}

func addNotify(typ, text string, userId uint) *notify {
	notify := &notify{
		typ:    typ,
		Text:   text,
		UserId: userId,
	}
	notifies[userId] = append(notifies[userId], notify)
	return notify
}

func NotifyPrimary(text string, userId uint) *notify {
	return addNotify("primary", text, userId)
}

func NotifySecondary(text string, userId uint) *notify {
	return addNotify("secondary", text, userId)
}

func NotifySuccess(text string, userId uint) *notify {
	return addNotify("success", text, userId)
}

func NotifyDanger(text string, userId uint) *notify {
	return addNotify("danger", text, userId)
}

func NotifyWarning(text string, userId uint) *notify {
	return addNotify("warning", text, userId)
}

func NotifyInfo(text string, userId uint) *notify {
	return addNotify("info", text, userId)
}

func NotifyLight(text string, userId uint) *notify {
	return addNotify("light", text, userId)
}

func NotifyDark(text string, userId uint) *notify {
	return addNotify("dark", text, userId)
}

func Notifies(userId uint, clear ...bool) []*notify {
	n := notifies[userId]
	if len(clear) > 0 && clear[0] {
		delete(notifies, userId)
	}
	return n
}
