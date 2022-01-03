package memory

import (
	"fmt"
	"sync"
	"time"
)

type edgeList []uuid.UUID

type InMemoryGraph struct {
	mu sync.RWMutex

	links map[uuid.UUID]*graph.Link
	edges map[uuid.UUID]*graph.Edge

	linkURLIndex map[string]*graph.Link
	linkEdgeMap  map[uuid.UUID]edgeList
}

type linkIterator struct {
	s *InMemoryGraph

	links    []*graph.Link
	curIndex int
}

type edgeIterator struct {
	s        *InMemoryGraph
	edges    []*graph.Edge
	curIndex int
}

func (s *InMemoryGraph) UpsertLink(link *Link) error {

	if link.ID == uuid.Nil {
		link.ID = existing.ID
		origIs := existing.RetrievedAt
		*existing = *&link
		if origIs.After(existing.RetrievedAt) {
			existing.RetrievedAt = origIs
		}
		return nil
	}

	//Insert new link to the graph
	for {
		link.ID = uuid.New()
		if s.links[link.ID] == nil {
			break
		}
	}
	lCopy := new(graph.Link)
	*lCopy = *link
	s.links[lCopy.ID] = lCopy
	return nil
}

func (s *InMemoryGraph) UpsertEdge(edge *Edge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, srcExists := s.links[edge.Src]
	_, dstExists := s.links[edge.Dst]
	if !srcExists || !dstExists {
		return fmt.Errorf("upsert edge: %w", graph.ErrUnknownEdgeLinks)
	}

	//Scan edge to update
	for _, edgeId := range s.linkEdgeMap[edge.Src] {
		existingEdge := s.edges[edgeId]
		if existingEdge.Src == edge.Src && existingEdge.Dst == edge.Dst {
			existingEdge.UpdatedAt = time.Now()
			*edge = *existingEdge
			return nil
		}
	}

	for {
		edge.ID = uuid.New()
		if s.edges[edge.ID] == nil {
			break
		}
	}
	edge.UpdatedAt = time.Now()
	eCopy := new(graph.Edge)
	*eCopy = *edge
	s.edges[eCopy.ID] = eCopy

	//Append the edge ID to the list of edges originating from the edge's source link
	s.linkEdgeMap[edge.Src] = append(s.linkEdgeMap[edge.Src], eCopy.ID)
	return nil
}

func (s *InMemoryGraph) FindLink(id uuid.UUID) (*graph.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link := s.links[id]
	if link == nil {
		return nil, fmt.Errorf("find link: %s", graph.ErrNotFound)
	}

	//we want to ensure that no external code can modify the graph's content without invoking the UpsertLink, that's why we return a copy of the link
	lCopy := new(graph.Link)
	*lCopy = *link
	return lCopy, nil
}

func (s *InMemoryGraph) Links(fromID, toID uuid.UUID, retrievedBefore time.Time) (graph.LinkIterator, error) {
	from, to := fromID.String(), toID.String()

	s.mu.RLock()
	var list []*graph.Link
	for linkID, link := range s.links {
		if id := linkID.String(); id >= from && id < to && link.RetrievedAt.Before(retrievedBefore) {
			list = append(list, link)
		}
	}
	s.mu.Unlock()

	return &linkIterator{s: s, links: list}, nil
}

func (i *linkIterator) Link() *graph.Link {
	i.s.mu.RLock()
	link := new(graph.Link)
	*link = *i.links[i.curIndex-1]
	i.s.mu.RUnlock()
	return link
}

func (s *InMemoryGraph) Edges(fromID, toID uuid.UUID, updatedBefore time.Time) (graph.EdgeIterator, error) {
	from, to := fromID.String(), toID.String()
	s.mu.RLock()
	var list []*graph.Edge
	for linkID := range s.links {
		if id := linkID.String(); id < from || id >= to {
			continue
		}
		for _, edgeID := range s.linkEdgeMap[linkID] {
			if edge := s.edges[edgeID]; edge.UpdatedAt.Before(updatedBefore) {
				list = append(list, edge)
			}
		}
	}
	s.mu.RUnlock()
	return &edgeIterator{s: s, edges: list}, nil
}

func (i *edgeIterator) Edge() *graph.Edge {
	i.s.mu.RLock()
	edge := new(graph.Edge)
	*edge = *i.edges[i.curIndex-1]
	i.s.mu.RUnlock()
	return edge
}

func (s *InMemoryGraph) RemoveStaleEdges(fromID uuid.UUID, updatedBefore time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var newEdgeList edgeList
	for _, edgeID := range s.linkEdgeMap[fromID] {
		edge := s.edges[edgeID]
		if edge.UpdatedAt().Before(updatedBefore) {
			delete(s.edges, edgeID)
			continue
		}
		newEdgeList = append(newEdgeList, edgeID)
	}
	s.linkEdgeMap[fromID] = newEdgeList
	return nil
}
