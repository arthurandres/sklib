package sklib

type SearchAPI interface {
	Browse(request BrowseRoutesRequest) (FullQuotes, error)
}

type EngineSearchAPI struct {
	engine RequestEngine
}

func (m *EngineSearchAPI) Browse(request BrowseRoutesRequest) (FullQuotes, error) {
	return Browse(m.engine, request)
}
