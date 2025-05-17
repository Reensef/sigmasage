package domain

type BotInfo struct {
	ID int64
}

type SMACSignalDial struct {
	Deal   Deal
	Signal SMACSignal
}
