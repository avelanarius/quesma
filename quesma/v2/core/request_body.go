// Copyright Quesma, licensed under the Elastic License 2.0.
// SPDX-License-Identifier: Elastic-2.0
package quesma_api

type RequestBody interface {
	IsParsedRequestBody() // this is a marker method
}
