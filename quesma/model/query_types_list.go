package model

// list of all aggregation types in Elasticsearch.
var AggregationQueryTypes = []string{
	// metrics:
	"avg",
	"boxplot",
	"cardinality",
	"extended_stats",
	"geo_bounds",
	"geo_centroid",
	"geo_line",
	"cartesian_bounds",
	"cartesian_centroid",
	"matrix_stats",
	"max",
	"median_absolute_deviation",
	"min",
	"percentile_ranks",
	"percentiles",
	"rate",
	"scripted_metric",
	"stats",
	"string_stats",
	"sum",
	"t_test",
	"top_hits",
	"top_metrics",
	"value_count",
	"weighted_avg",

	// bucket:
	"adjacency_matrix",
	"auto_date_histogram",
	"categorize_text",
	"children",
	"composite",
	"date_histogram",
	"date_range",
	"diversified_sampler",
	"filter",
	"filters",
	"frequent_item_sets",
	"geo_distance",
	"geohash_grid",
	"geohex_grid",
	"geotile_grid",
	"global",
	"histogram",
	"ip_prefix",
	"ip_range",
	"missing",
	"multi_terms",
	"nested",
	"parent",
	"random_sampler",
	"range",
	"rare_terms",
	"reverse_nested",
	"sampler",
	"significant_terms",
	"significant_text",
	"terms",
	"time_series",
	"variable_width_histogram",

	// pipeline:
	"avg_bucket",
	"bucket_script",
	"bucket_count_ks_test",
	"bucket_correlation",
	"bucket_selector",
	"bucket_sort",
	"change_point",
	"cumulative_cardinality",
	"cumulative_sum",
	"derivative",
	"extended_stats_bucket",
	"inference",
	"max_bucket",
	"min_bucket",
	"moving_avg",
	"moving_fn",
	"moving_percentiles",
	"normalize",
	"percentiles_bucket",
	"serial_diff",
	"stats_bucket",
	"sum_bucket",
}

// TODO list of all Query DSL types in Elasticsearch.