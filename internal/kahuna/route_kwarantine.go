/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"context"

	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

func NewKwarantineRouter() sdk.Router {
	return sdk.NewKwarantineAPIController(&KwarantineService{})
}

type KwarantineService struct{}

func (s *KwarantineService) AddKwarantineInstance(ctx context.Context, kwarantineId string, instanceId string) (sdk.ImplResponse, error) {
	return HttpNotImplemented(nil)
}

func (s *KwarantineService) AddKwarantineKompute(ctx context.Context, kwarantineId string, komputeId string) (sdk.ImplResponse, error) {
	return HttpNotImplemented(nil)
}

func (s *KwarantineService) DeleteKwarantine(ctx context.Context, kwarantineId string) (sdk.ImplResponse, error) {
	return HttpNotImplemented(nil)
}

func (s *KwarantineService) ListKwarantines(ctx context.Context) (sdk.ImplResponse, error) {
	return HttpNotImplemented(nil)
}

func (s *KwarantineService) ReadKwarantine(ctx context.Context, kwarantineId string) (sdk.ImplResponse, error) {
	return HttpNotImplemented(nil)
}

func (s *KwarantineService) RemoveKwarantineInstance(ctx context.Context, kwarantineId string, instanceId string) (sdk.ImplResponse, error) {
	return HttpNotImplemented(nil)
}

func (s *KwarantineService) RemoveKwarantineKompute(ctx context.Context, kwarantineId string, komputeId string) (sdk.ImplResponse, error) {
	return HttpNotImplemented(nil)
}

func (s *KwarantineService) UpdateKwarantine(ctx context.Context, kwarantineId string, kwarantine sdk.Kwarantine) (sdk.ImplResponse, error) {
	return HttpNotImplemented(nil)
}
