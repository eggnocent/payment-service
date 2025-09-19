package constants

type PaymentStatus int
type PaymentStatusString string

const (
	Initial    PaymentStatus = 0
	Pending    PaymentStatus = 100
	Settlement PaymentStatus = 200
	Expire     PaymentStatus = 300

	InitialString    PaymentStatusString = "Initial"
	PendingString    PaymentStatusString = "Pending"
	SettlementString PaymentStatusString = "Settlement"
	ExpireString     PaymentStatusString = "Expire"
)

var mapStatusStringtoINT = map[PaymentStatusString]PaymentStatus{
	InitialString:    Initial,
	PendingString:    Pending,
	SettlementString: Settlement,
	ExpireString:     Expire,
}

var mapStatusIntToString = map[PaymentStatus]PaymentStatusString{
	Initial:    InitialString,
	Pending:    PendingString,
	Settlement: SettlementString,
	Expire:     ExpireString,
}

func (p PaymentStatus) GetStatusString() PaymentStatusString {
	return mapStatusIntToString[p]
}

func (p PaymentStatusString) GetStatusInt() PaymentStatus {
	return mapStatusStringtoINT[p]
}
