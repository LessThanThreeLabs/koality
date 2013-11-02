package events

const (
	exchangeName       = "model:events" // TODO: change to "events"
	exchangeType       = "fanout"
	exchangeDurable    = false
	exchangeAutoDelete = false
	exchangeInternal   = false
	exchangeNoWait     = false
	exchangeMandatory  = true
	exchangeImmediate  = false
)
