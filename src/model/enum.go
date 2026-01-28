package model

import (
	"fmt"
	"strings"
)

type EventType string

// const (
// 	LOGOUT          EventType = "LOGOUT"
// 	WITHDRAWAL      EventType = "WITHDRAWAL"
// 	KYC_UPDATE      EventType = "KYC_UPDATE"
// 	PROFILE_UPDATE  EventType = "PROFILE_UPDATE"
// 	PAYMENT         EventType = "PAYMENT"
// 	ACCOUNT_UPDATE  EventType = "ACCOUNT_UPDATE"
// 	CUSTOMER_EVENT  EventType = "CUSTOMER_EVENT"
// 	TRANSACTION     EventType = "TRANSACTION"
// 	KYC             EventType = "KYC"
// 	REGISTER        EventType = "REGISTER"
// 	PASSWORD_CHANGE EventType = "PASSWORD_CHANGE"
// 	LOGIN           EventType = "LOGIN"
// 	MANUAL          EventType = "MANUAL"
// )

// func AllEventTypes() []EventType {
// 	return []EventType{
// 		LOGOUT, WITHDRAWAL, KYC_UPDATE, PROFILE_UPDATE, PAYMENT, ACCOUNT_UPDATE,
// 		CUSTOMER_EVENT, TRANSACTION, KYC, REGISTER, PASSWORD_CHANGE, LOGIN, MANUAL,
// 	}
// }

func ParseEventType(s string) (EventType, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return "", fmt.Errorf("event_type required")
	}
	// Dynamic event types: we allow any string that looks reasonable (basic char check optional, or just pass through)
	// For backward compatibility and strictness where needed, we could check against AllEventTypes,
	// but the goal is dynamic support.
	return EventType(s), nil
}

func IsMoneyBearing(et EventType) bool {
	// Heuristic: Check for common financial keywords in the event type.
	s := strings.ToUpper(string(et))
	keywords := []string{"PAYMENT", "TRANSACTION", "WITHDRAWAL", "DEPOSIT", "TRANSFER", "PURCHASE", "REFUND"}
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
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
