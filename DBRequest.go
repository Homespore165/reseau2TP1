package main

type DBRequest struct {
	QueryType  string        // e.g. "create", "insert", "select"
	Parameters []interface{} // Query parameters
	Response   chan DBResponse
}
