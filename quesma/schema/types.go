// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0
package schema

import "slices"

var (
	// TODO add more and review existing
	TypeText         = Type{Name: "text", Properties: []TypeProperty{Searchable, FullText}}
	TypeKeyword      = Type{Name: "keyword", Properties: []TypeProperty{Searchable, Aggregatable}}
	TypeInteger      = Type{Name: "integer", Properties: []TypeProperty{Searchable, Aggregatable}}
	TypeLong         = Type{Name: "long", Properties: []TypeProperty{Searchable, Aggregatable}}
	TypeUnsignedLong = Type{Name: "unsigned_long", Properties: []TypeProperty{Searchable, Aggregatable}}
	TypeTimestamp    = Type{Name: "timestamp", Properties: []TypeProperty{Searchable, Aggregatable}}
	TypeDate         = Type{Name: "date", Properties: []TypeProperty{Searchable, Aggregatable}}
	TypeFloat        = Type{Name: "float", Properties: []TypeProperty{Searchable, Aggregatable}}
	TypeBoolean      = Type{Name: "boolean", Properties: []TypeProperty{Searchable, Aggregatable}}
	TypeObject       = Type{Name: "object", Properties: []TypeProperty{Searchable}}
	TypeArray        = Type{Name: "array", Properties: []TypeProperty{Searchable}}
	TypeMap          = Type{Name: "map", Properties: []TypeProperty{Searchable}}
	TypeIp           = Type{Name: "ip", Properties: []TypeProperty{Searchable, Aggregatable}}
	TypePoint        = Type{Name: "point", Properties: []TypeProperty{Searchable, Aggregatable}}
	TypeUnknown      = Type{Name: "unknown", Properties: []TypeProperty{Searchable}}
)

const (
	Aggregatable TypeProperty = "aggregatable"
	Searchable   TypeProperty = "searchable"
	FullText     TypeProperty = "full_text"
)

func (t Type) Equal(t2 Type) bool {
	return t.Name == t2.Name
}

func (t Type) IsAggregatable() bool {
	return slices.Contains(t.Properties, Aggregatable)
}

func (t Type) IsSearchable() bool {
	return slices.Contains(t.Properties, Searchable)
}

func (t Type) IsFullText() bool {
	return slices.Contains(t.Properties, FullText)
}

type (
	Type struct {
		Name       string
		Properties []TypeProperty
	}
	TypeProperty string
)

func (t Type) String() string {
	return t.Name
}

func ParseQuesmaType(t string) (Type, bool) {
	switch t {
	case TypeText.Name:
		return TypeText, true
	case TypeKeyword.Name:
		return TypeKeyword, true
	case TypeLong.Name:
		return TypeLong, true
	case TypeTimestamp.Name:
		return TypeTimestamp, true
	case TypeDate.Name:
		return TypeDate, true
	case TypeFloat.Name:
		return TypeFloat, true
	case TypeBoolean.Name, "bool":
		return TypeBoolean, true
	case TypeObject.Name, "json":
		return TypeObject, true
	case TypeArray.Name:
		return TypeArray, true
	case TypeMap.Name:
		return TypeMap, true
	case TypeIp.Name:
		return TypeIp, true
	case TypePoint.Name, "geo_point":
		return TypePoint, true
	default:
		return TypeUnknown, false
	}
}
