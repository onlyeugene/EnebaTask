// Package atlas holds the Atlas Search index definition. The JSON file is the one to paste
// into the Atlas UI if the seed can't create the index; embedding it means the Go code and
// the file can never drift apart.
package atlas

import _ "embed"

//go:embed products_search.json
var ProductsSearch []byte
