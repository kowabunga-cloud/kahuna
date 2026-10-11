/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"context"
	"fmt"

	"github.com/kowabunga-cloud/kahuna/internal/sdk"
)

func NewKwarantineRouter() sdk.Router {
	return sdk.NewKwarantineAPIController(&KwarantineService{}, sdk.WithKwarantineAPIErrorHandler(ApiErrorHandler))
}

type KwarantineService struct{}

func (s *KwarantineService) AddKwarantineInstance(ctx context.Context, kwarantineId string, instanceId string) (sdk.ImplResponse, error) {
	k, err := FindKwarantineByID(kwarantineId)
	if err != nil {
		return HttpNotFound(err)
	}

	// ensure instance exists
	_, err = FindInstanceByID(instanceId)
	if err != nil {
		return HttpNotFound(err)
	}

	err = k.AddInstance(instanceId)
	if err != nil {
		return HttpServerError(err)
	}

	payload := k.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}

func (s *KwarantineService) AddKwarantineKompute(ctx context.Context, kwarantineId string, komputeId string) (sdk.ImplResponse, error) {
	k, err := FindKwarantineByID(kwarantineId)
	if err != nil {
		return HttpNotFound(err)
	}

	// ensure Kompute exists
	_, err = FindKomputeByID(komputeId)
	if err != nil {
		return HttpNotFound(err)
	}

	err = k.AddKompute(komputeId)
	if err != nil {
		return HttpServerError(err)
	}

	payload := k.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}

func (s *KwarantineService) DeleteKwarantine(ctx context.Context, kwarantineId string) (sdk.ImplResponse, error) {
	k, err := FindKwarantineByID(kwarantineId)
	if err != nil {
		return HttpNotFound(err)
	}

	err = k.Delete()
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *KwarantineService) ListKwarantines(ctx context.Context) (sdk.ImplResponse, error) {
	return HttpOK(FindResourceIDs(MongoCollectionKwarantineName))
}

func (s *KwarantineService) ReadKwarantine(ctx context.Context, kwarantineId string) (sdk.ImplResponse, error) {
	k, err := FindKwarantineByID(kwarantineId)
	if err != nil {
		return HttpNotFound(err)
	}

	payload := k.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}

func (s *KwarantineService) RemoveKwarantineInstance(ctx context.Context, kwarantineId string, instanceId string) (sdk.ImplResponse, error) {
	k, err := FindKwarantineByID(kwarantineId)
	if err != nil {
		return HttpNotFound(err)
	}

	err = k.RemoveInstance(instanceId)
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *KwarantineService) RemoveKwarantineKompute(ctx context.Context, kwarantineId string, komputeId string) (sdk.ImplResponse, error) {
	k, err := FindKwarantineByID(kwarantineId)
	if err != nil {
		return HttpNotFound(err)
	}

	err = k.RemoveKompute(komputeId)
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *KwarantineService) UpdateKwarantine(ctx context.Context, kwarantineId string, kwarantine sdk.KwarantineUpdate) (sdk.ImplResponse, error) {
	LogHttpRequest(RA("kwarantineId", kwarantineId), RA("kwarantine", kwarantine))

	// ensure Kwarantine exists
	k, err := FindKwarantineByID(kwarantineId)
	if err != nil {
		return HttpNotFound(err)
	}

	// unspecified policy is left unchanged
	policy := kwarantine.Policy
	if policy == "" {
		policy = k.Policy
	}
	if policy != KwarantinePolicyHost && policy != KwarantinePolicyZone {
		return HttpBadParams(fmt.Errorf("invalid Kwarantine policy: %s", policy))
	}

	err = k.Update(kwarantine.Name, kwarantine.Description, policy)
	if err != nil {
		return HttpServerError(err)
	}

	payload := k.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}
