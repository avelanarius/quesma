package model

type Query struct {
	Sql       string
	TableName string
	CanParse  bool
}

type AsyncSearchQueryType int

const (
	Histogram AsyncSearchQueryType = iota
	AggsByField
	ListByField
	ListAllFields
	None
)

type QueryInfo struct {
	Typ       AsyncSearchQueryType
	FieldName string
	I1        int
	I2        int
}