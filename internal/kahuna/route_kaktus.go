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

func NewKaktusRouter() sdk.Router {
	return sdk.NewKaktusAPIController(&KaktusService{}, sdk.WithKaktusAPIErrorHandler(ApiErrorHandler))
}

type KaktusService struct{}

func (s *KaktusService) DeleteKaktus(ctx context.Context, kaktusId string) (sdk.ImplResponse, error) {
	// ensure kaktus exists
	h, err := FindKaktusByID(kaktusId)
	if err != nil {
		return HttpNotFound(err)
	}

	// check if kaktus still has children referenced
	if h.HasChildren() {
		return HttpConflict(nil)
	}

	// remove kaktus
	err = h.Delete()
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *KaktusService) DisableKaktusMaintenance(ctx context.Context, kaktusId string) (sdk.ImplResponse, error) {
	h, err := FindKaktusByID(kaktusId)
	if err != nil {
		return HttpNotFound(err)
	}

	err = h.DisableMaintenance()
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *KaktusService) EnableKaktusMaintenance(ctx context.Context, kaktusId string) (sdk.ImplResponse, error) {
	h, err := FindKaktusByID(kaktusId)
	if err != nil {
		return HttpNotFound(err)
	}

	err = h.EnableMaintenance()
	if err != nil {
		return HttpServerError(err)
	}

	return HttpOK(nil)
}

func (s *KaktusService) ListKaktusInstances(ctx context.Context, kaktusId string) (sdk.ImplResponse, error) {
	h, err := FindKaktusByID(kaktusId)
	if err != nil {
		return HttpNotFound(err)
	}

	payload := h.Instances()
	return HttpOK(payload)
}

func (s *KaktusService) ListKaktuses(ctx context.Context) (sdk.ImplResponse, error) {
	return HttpOK(FindResourceIDs(MongoCollectionKaktusName))
}

func (s *KaktusService) ReadKaktus(ctx context.Context, kaktusId string) (sdk.ImplResponse, error) {
	h, err := FindKaktusByID(kaktusId)
	if err != nil {
		return HttpNotFound(err)
	}

	payload := h.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}

func (s *KaktusService) ReadKaktusCaps(ctx context.Context, kaktusId string) (sdk.ImplResponse, error) {
	h, err := FindKaktusByID(kaktusId)
	if err != nil {
		return HttpNotFound(err)
	}

	payload := h.Capabilities.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}

func (s *KaktusService) UpdateKaktus(ctx context.Context, kaktusId string, kaktus sdk.KaktusUpdate) (sdk.ImplResponse, error) {
	LogHttpRequest(RA("kaktusId", kaktusId), RA("kaktus", kaktus))

	// ensure kaktus exists
	h, err := FindKaktusByID(kaktusId)
	if err != nil {
		return HttpNotFound(err)
	}

	// unspecified settings are left unchanged
	cpu_price := h.Costs.CPU.Price
	cpu_currency := h.Costs.CPU.Currency
	if kaktus.CpuCost != nil && kaktus.CpuCost.Price != 0 {
		cpu_price = kaktus.CpuCost.Price
		cpu_currency = kaktus.CpuCost.Currency
	}

	memory_price := h.Costs.Memory.Price
	memory_currency := h.Costs.Memory.Currency
	if kaktus.MemoryCost != nil && kaktus.MemoryCost.Price != 0 {
		memory_price = kaktus.MemoryCost.Price
		memory_currency = kaktus.MemoryCost.Currency
	}

	overcommit_cpu := h.OverCommit.CPU
	if kaktus.OvercommitCpuRatio != 0 {
		overcommit_cpu = kaktus.OvercommitCpuRatio
	}
	overcommit_memory := h.OverCommit.Memory
	if kaktus.OvercommitMemoryRatio != 0 {
		overcommit_memory = kaktus.OvercommitMemoryRatio
	}

	agents := kaktus.Agents
	if agents == nil {
		agents = h.AgentIDs
	}

	err = h.Update(kaktus.Name, kaktus.Description, cpu_price, cpu_currency, memory_price, memory_currency, overcommit_cpu, overcommit_memory, agents)
	if err != nil {
		return HttpServerError(err)
	}

	payload := h.Model()
	LogHttpResponse(payload)
	return HttpOK(payload)
}
