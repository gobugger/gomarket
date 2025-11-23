package view

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/gobugger/gomarket/internal/repo"
	"time"
)

type Ticket struct {
	Ticket       repo.Ticket
	Responses    []repo.TicketResponse
	PrevResponse time.Time
	AuthorName   string
}

type TicketView struct{}

func (tv TicketView) Get(ctx context.Context, q *repo.Queries, id uuid.UUID) (*Ticket, error) {
	ticket, err := q.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}

	author, err := q.GetUser(ctx, ticket.AuthorID)
	if err != nil {
		return nil, err
	}

	responses, err := q.GetTicketResponsesForTicket(ctx, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	t := &Ticket{
		Ticket:     ticket,
		Responses:  responses,
		AuthorName: author.Username,
	}

	if len(responses) > 0 {
		t.PrevResponse = responses[0].CreatedAt
	}

	return t, nil
}

func (tv TicketView) GetOpenTickets(ctx context.Context, q *repo.Queries) ([]Ticket, error) {
	tickets, err := q.GetOpenTickets(ctx)
	if err != nil {
		return nil, err
	}

	view := []Ticket{}
	for _, ticket := range tickets {
		author, err := q.GetUser(ctx, ticket.AuthorID)
		if err != nil {
			return nil, err
		}

		responses, err := q.GetTicketResponsesForTicket(ctx, ticket.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

		t := Ticket{
			Ticket:     ticket,
			AuthorName: author.Username,
			Responses:  responses,
		}

		if len(responses) > 0 {
			t.PrevResponse = responses[0].CreatedAt
		}

		view = append(view, t)
	}
	return view, nil
}
