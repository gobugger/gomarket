package order

import (
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/stretchr/testify/require"
	"testing"
)

// Find all status pairs that are not in valid transitions
func findInvalidTransitions(valid [][]repo.OrderStatus) [][]repo.OrderStatus {
	res := [][]repo.OrderStatus{}
	ss := []repo.OrderStatus{
		repo.OrderStatusPending,
		repo.OrderStatusPaid,
		repo.OrderStatusAccepted,
		repo.OrderStatusDeclined,
		repo.OrderStatusDispatched,
		repo.OrderStatusFinalized,
		repo.OrderStatusDisputed,
		repo.OrderStatusSettled,
	}

	exist := func(a repo.OrderStatus, b repo.OrderStatus) bool {
		for _, flow := range valid {
			for i := range flow[:len(flow)-1] {
				if a == flow[i] && b == flow[i+1] {
					return true
				}
			}
		}
		return false
	}

	for i := range ss {
		for j := range ss {
			if i != j && !exist(ss[i], ss[j]) {
				res = append(res, []repo.OrderStatus{ss[i], ss[j]})
			}
		}
	}

	return res
}

func TestTransitions(t *testing.T) {
	validTransitions := [][]repo.OrderStatus{
		{ // normal
			repo.OrderStatusPending,
			repo.OrderStatusPaid,
			repo.OrderStatusAccepted,
			repo.OrderStatusDispatched,
			repo.OrderStatusFinalized,
		},
		{ // dispute settled
			repo.OrderStatusDispatched,
			repo.OrderStatusDisputed,
			repo.OrderStatusSettled,
		},
		{ // invoice expired or cancelled
			repo.OrderStatusPending,
			repo.OrderStatusCancelled,
		},
		{ // vendor declined
			repo.OrderStatusPaid,
			repo.OrderStatusDeclined,
		},
		{ // vendor forgot to dispatch
			repo.OrderStatusAccepted,
			repo.OrderStatusDeclined,
		},
	}

	for _, transitions := range validTransitions {
		for i := range transitions[:len(transitions)-1] {
			from := transitions[i]
			to := transitions[i+1]
			require.Truef(t, validTransition(from, to), "%v -> %v shoud be valid transition", from, to)
		}
	}

	invalidTransitions := findInvalidTransitions(validTransitions)

	for _, transitions := range invalidTransitions {
		isValid := true
		for i := range transitions[:len(transitions)-1] {
			from := transitions[i]
			to := transitions[i+1]

			if !validTransition(from, to) {
				isValid = false
				break
			}
		}

		require.Falsef(t, isValid, "transitions %v should be invalid", transitions)
	}
}
