package schema

import (
	"fmt"
	"mitmproxy/quesma/clickhouse"
	"mitmproxy/quesma/concurrent"
	"mitmproxy/quesma/logger"
	"mitmproxy/quesma/quesma/config"
	"sync/atomic"
	"time"
)

type (
	Schema struct {
		Fields map[FieldName]Field
	}
	Field struct {
		Name FieldName
		Type Type
	}
	TableName string
	FieldName string
	Type      string
)

type (
	Registry interface {
		AllSchemas() map[TableName]Schema
		FindSchema(name TableName) (Schema, bool)
		Load() error
		Start()
	}
	schemaRegistry struct {
		started                atomic.Bool
		schemas                *concurrent.Map[TableName, Schema]
		configuration          config.QuesmaConfiguration
		clickhouseSchemaLoader *clickhouse.SchemaLoader
		ClickhouseTypeAdapter  ClickhouseTypeAdapter
	}
)

func (s *schemaRegistry) Start() {
	if s.started.CompareAndSwap(false, true) {
		s.loadTypeMappingsFromConfiguration()
	}

	// TODO remove
	go func() {
		for {
			<-time.After(5 * time.Second)
			_ = s.Load()
		}
	}()
}

func (s *schemaRegistry) loadTypeMappingsFromConfiguration() {
	for _, indexConfiguration := range s.configuration.IndexConfig {
		if !indexConfiguration.Enabled {
			continue
		}
		if indexConfiguration.SchemaConfiguration != nil {
			logger.Debug().Msgf("loading schema for index %s", indexConfiguration.Name)
			fields := make(map[FieldName]Field)
			for _, field := range indexConfiguration.SchemaConfiguration.Fields {
				fieldName := FieldName(field.Name)
				fields[fieldName] = Field{
					Name: fieldName,
					// TODO check if type is valid
					Type: Type(field.Type),
				}
			}
			s.schemas.Store(TableName(indexConfiguration.Name), Schema{Fields: fields})
		}
	}
}

func (s *schemaRegistry) Load() error {
	if !s.started.Load() {
		return fmt.Errorf("schema registry not started")
	}
	definitions := s.clickhouseSchemaLoader.TableDefinitions()
	schemas := s.schemas.Snapshot()
	definitions.Range(func(indexName string, value *clickhouse.Table) bool {
		logger.Debug().Msgf("loading schema for table %s", indexName)
		fields := make(map[FieldName]Field)
		if schema, found := schemas[TableName(indexName)]; found {
			fields = schema.Fields
		}
		for _, col := range value.Cols {
			indexConfig := s.configuration.IndexConfig[indexName]
			if explicitType, found := indexConfig.TypeMappings[col.Name]; found {
				logger.Debug().Msgf("found explicit type mapping for column %s: %s", col.Name, explicitType)
				fields[FieldName(col.Name)] = Field{
					Name: FieldName(col.Name),
					Type: Type(explicitType),
				}
				continue
			}
			if _, exists := fields[FieldName(col.Name)]; !exists {
				quesmaType, found := s.ClickhouseTypeAdapter.Adapt(col.Type.String())
				if !found {
					logger.Error().Msgf("type %s not supported", col.Type.String())
					continue
				} else {
					fields[FieldName(col.Name)] = Field{
						Name: FieldName(col.Name),
						Type: quesmaType, // TODO convert to our type
					}
				}
			}
		}
		s.schemas.Store(TableName(indexName), Schema{Fields: fields})
		return true
	})
	for name, schema := range s.schemas.Snapshot() {
		fmt.Printf("schema: %s\n", name)
		for fieldName, field := range schema.Fields {
			fmt.Printf("\tfield: %s, type: %s\n", fieldName, field.Type)
		}

		break
	}
	return nil
}

func (s *schemaRegistry) AllSchemas() map[TableName]Schema {
	return s.schemas.Snapshot()
}

func (s *schemaRegistry) FindSchema(name TableName) (Schema, bool) {
	schema, found := s.schemas.Load(name)
	return schema, found
}

func NewSchemaRegistry(schemaManagement *clickhouse.SchemaLoader, configuration config.QuesmaConfiguration) Registry {
	return &schemaRegistry{
		schemas:                concurrent.NewMap[TableName, Schema](),
		started:                atomic.Bool{},
		configuration:          configuration,
		clickhouseSchemaLoader: schemaManagement,
		ClickhouseTypeAdapter:  NewClickhouseTypeAdapter(),
	}
}
