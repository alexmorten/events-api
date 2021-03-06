package search

import (
	"context"
	"fmt"
	"reflect"

	"github.com/olivere/elastic"
)

const nodeIndexName = "neo4j-index-node"

//Client wraps the elasticsearch client for ease of use
type Client struct {
	address string
	*elastic.Client
}

//NewClient connects to elasticsearch and returns a handler on success
// error will be returned if no connection could be eastablished
func NewClient(address string, lazy bool) (*Client, error) {
	client := &Client{
		address: address,
	}
	if lazy {
		return client, nil
	}
	return client, client.ensureConnectionExists()
}

//FuzzyNameSearch through nodes with given label
func (c *Client) FuzzyNameSearch(label, searchTerm string, iterator func(props map[string]interface{})) error {
	if c.Client == nil {
		err := c.ensureConnectionExists()
		if err != nil {
			return err
		}
	}

	query := elastic.NewBoolQuery().Should(
		elastic.NewMatchQuery("labels", label),
		elastic.NewFuzzyQuery("name", searchTerm),
	)

	searchResult, err := c.Search().Index(nodeIndexName).Query(query).Do(context.Background())
	if err != nil {
		return err
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	// Each is a utility function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization. If you want full control
	// over iterating the hits, see below.
	var ttyp map[string]interface{}
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		props := item.(map[string]interface{})
		iterator(props)
	}

	return nil
}

func (c *Client) ensureConnectionExists() error {
	client, err := elastic.NewClient(
		elastic.SetURL(c.address),
		elastic.SetSniff(false),
	)
	if err == nil {
		c.Client = client
	}
	return err
}
