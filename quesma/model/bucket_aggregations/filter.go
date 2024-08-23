// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0
package bucket_aggregations

import (
	"context"
	"quesma/logger"
	"quesma/model"
)

type FilterAgg struct {
	ctx         context.Context
	WhereClause model.Expr
}

func NewFilterAgg(ctx context.Context, whereClause model.Expr) FilterAgg {
	return FilterAgg{ctx: ctx, WhereClause: whereClause}
}

func (query FilterAgg) AggregationType() model.AggregationType {
	return model.BucketAggregation
}

func (query FilterAgg) TranslateSqlResponseToJson(rows []model.QueryResultRow, level int) model.JsonMap {
	if len(rows) == 0 {
		logger.WarnWithCtx(query.ctx).Msg("no rows returned for filter aggregation")
		return make(model.JsonMap, 0)
	}
	return model.JsonMap{"doc_count": rows[0].Cols[level].Value}
}

func (query FilterAgg) String() string {
	return "count"
}

func (query FilterAgg) DoesNotHaveGroupBy() bool {
	return true
}