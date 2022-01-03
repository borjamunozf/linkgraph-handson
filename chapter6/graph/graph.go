package graph

import "time"

type Graph interface {
	UpsertLink(link *Link) error
	FindLink(uuid uuid.UUID) (*Link, error)

	UpsertEdge(edge *Edge) error
	RemoveStaleEdge(fromID uuid.UUID, updatedBefore time.Time) error

	Links(fromID, toID uuid.UUID, retrievedBefore time.Time) (LinkIterator, error)
	Edges(fromID, toID uuid.UUID, updatedBefore time.Time) (EdgeIterator, error)
}

type Link struct {
	ID          uuid.UUID
	URL         string
	RetrievedAt time.Time
}

type Edge struct {
	ID        uuid.UUID
	Src       uuid.UUID
	Dst       uuid.UUID
	UpdatedAt time.Time
}

type LinkIterator interface {
	Iterator
	Link() *Link
}

type EdgeIterator interface {
	Iterator
	Edge() *Edge
}

type Iterator interface {
	Nex() bool
	Error() error
	Close() error
}
