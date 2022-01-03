package cockroach

import (
	"database/sql"
)

type CockroachDBGraph struct {
	db sql.DB
}

var upsertLinkQuery = `
INSERT INTO (url, retrieved_at) VALUES ($1, $2) ON CONFLICT (url) DO UPDATE SET retrieved_at=GREATEST(links.retrieved_at, $2)
RETURNING id, retrieved_at
`

func (c *CockroachDBGraph) UpsertLink(link *Link) error {
	asd := c.db.QueryRow(upsertLinkQuery, link.URL, link.RetrievedAt)

}
