package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8"
)

type Product struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ProductSearch struct {
	client *elasticsearch.Client
}

func NewProductSearch(esURL string) (*ProductSearch, error) {
	cfg := elasticsearch.Config{Addresses: []string{esURL}}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ProductSearch{client: client}, nil
}

// Index a product
func (ps *ProductSearch) IndexProduct(ctx context.Context, p Product) error {
	data, _ := json.Marshal(p)
	_, err := ps.client.Index("products", bytes.NewReader(data))
	return err
}

// Search by description
func (ps *ProductSearch) SearchByDescription(ctx context.Context, desc string) ([]Product, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"description": desc,
			},
		},
	}
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(query)
	res, err := ps.client.Search(
		ps.client.Search.WithContext(ctx),
		ps.client.Search.WithIndex("products"),
		ps.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var searchResult struct {
		Hits struct {
			Hits []struct {
				Source Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	products := make([]Product, len(searchResult.Hits.Hits))
	for i, hit := range searchResult.Hits.Hits {
		products[i] = hit.Source
	}

	return products, nil
}
