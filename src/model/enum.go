package model

import (
	"fmt"
	"strings"
)

type EventType string

const (
	LOGOUT          EventType = "LOGOUT"
	WITHDRAWAL      EventType = "WITHDRAWAL"
	KYC_UPDATE      EventType = "KYC_UPDATE"
	PROFILE_UPDATE  EventType = "PROFILE_UPDATE"
	PAYMENT         EventType = "PAYMENT"
	ACCOUNT_UPDATE  EventType = "ACCOUNT_UPDATE"
	CUSTOMER_EVENT  EventType = "CUSTOMER_EVENT"
	TRANSACTION     EventType = "TRANSACTION"
	KYC             EventType = "KYC"
	REGISTER        EventType = "REGISTER"
	PASSWORD_CHANGE EventType = "PASSWORD_CHANGE"
	LOGIN           EventType = "LOGIN"
	MANUAL          EventType = "MANUAL"
)

func AllEventTypes() []EventType {
	return []EventType{
		LOGOUT, WITHDRAWAL, KYC_UPDATE, PROFILE_UPDATE, PAYMENT, ACCOUNT_UPDATE,
		CUSTOMER_EVENT, TRANSACTION, KYC, REGISTER, PASSWORD_CHANGE, LOGIN, MANUAL,
	}
}

func ParseEventType(s string) (EventType, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	for _, v := range AllEventTypes() {
		if string(v) == s {
			return v, nil
		}
	}
	return "", fmt.Errorf("invalid event_type: %s", s)
}

func IsMoneyBearing(et EventType) bool {
	return et == PAYMENT || et == TRANSACTION || et == WITHDRAWAL
}

type NodeType string

const (
	NodeUser          NodeType = "USER"
	NodeMerchant      NodeType = "MERCHANT"
	NodeDevice        NodeType = "DEVICE"
	NodePaymentMethod NodeType = "PAYMENT_METHOD"
	NodeBank          NodeType = "BANK"
	NodeWallet        NodeType = "WALLET"
	NodeExchange      NodeType = "EXCHANGE"
)
