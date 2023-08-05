package types

var notifies = map[any][]*notify{}

type notify struct {
	typ    string
	Text   string
	UserId any
}

func (notify *notify) Type() string {
	return notify.typ
}

func addNotify(typ, text string, userId any) *notify {
	notify := &notify{
		typ:    typ,
		Text:   text,
		UserId: userId,
	}
	notifies[userId] = append(notifies[userId], notify)
	return notify
}

func NotifyPrimary(text string, userId any) *notify {
	return addNotify("primary", text, userId)
}

func NotifySecondary(text string, userId any) *notify {
	return addNotify("secondary", text, userId)
}

func NotifySuccess(text string, userId any) *notify {
	return addNotify("success", text, userId)
}

func NotifyDanger(text string, userId any) *notify {
	return addNotify("danger", text, userId)
}

func NotifyWarning(text string, userId any) *notify {
	return addNotify("warning", text, userId)
}

func NotifyInfo(text string, userId any) *notify {
	return addNotify("info", text, userId)
}

func NotifyLight(text string, userId any) *notify {
	return addNotify("light", text, userId)
}

func NotifyDark(text string, userId any) *notify {
	return addNotify("dark", text, userId)
}

func Notifies(userId any, clear ...bool) []*notify {
	n := notifies[userId]
	if len(clear) > 0 && clear[0] {
		delete(notifies, userId)
	}
	return n
}
