/*
 * Copyright (c) The Kowabunga Project
 * Apache License, Version 2.0 (see LICENSE or https://www.apache.org/licenses/LICENSE-2.0.txt)
 * SPDX-License-Identifier: Apache-2.0
 */

package kahuna

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"slices"
	"sync"

	"github.com/kowabunga-cloud/common"
	"github.com/kowabunga-cloud/common/agents"
	"github.com/kowabunga-cloud/common/klog"
	"github.com/kowabunga-cloud/common/wsrpc"
)

type KowabungaAgent struct {
	ID          string
	Client      *wsrpc.WsRpcClient
	Type        string
	IsConnected bool
	Version     string
	Methods     []string
}

// WatchKeepalive waits for the agent's WebSocket session to terminate and
// unregisters the agent, unless it has meanwhile been replaced by a newer
// session (agent reconnection) or disconnected on purpose.
func (agent *KowabungaAgent) WatchKeepalive() {
	<-agent.Client.Done()
	klog.Infof("Disconnection of %s agent %s has been detected. Unregistering ...", agent.Type, agent.ID)
	unregisterAgentSession(agent)
}

// agents singleton
var agentsLock = &sync.RWMutex{}
var agentsOnce sync.Once
var kAgents map[string]*KowabungaAgent

func GetAgents() map[string]*KowabungaAgent {
	agentsOnce.Do(func() {
		klog.Debugf("Creating Kowabunga Agents map")
		kAgents = map[string]*KowabungaAgent{}
	})

	return kAgents
}

// DisconnectAgent forcibly closes the agent's current WebSocket session, if
// any, e.g. when the agent is deleted or its API key changes. The agent has to
// reconnect and authenticate again.
func DisconnectAgent(agentId string) {
	agentsLock.Lock()
	ag := GetAgents()[agentId]
	delete(GetAgents(), agentId)
	agentsLock.Unlock()

	if ag == nil {
		return
	}
	klog.Infof("Disconnecting %s agent %s ...", ag.Type, agentId)
	err := ag.Client.Close()
	if err != nil {
		klog.Error(err)
	}
}

func GetAgent(agentId string) *KowabungaAgent {
	agentsLock.RLock()
	defer agentsLock.RUnlock()
	return GetAgents()[agentId]
}

func GetEligibleAgent(candidateAgents []string, method string) *KowabungaAgent {
	candidates := []*KowabungaAgent{}

	// build list of agents capable of addressing requested method
	agentsLock.RLock()
	for _, ag := range GetAgents() {
		// discard agents not associated with the RPC caller
		if !slices.Contains(candidateAgents, ag.ID) {
			continue
		}

		// discard agent with non-implemented RPC method
		if !slices.Contains(ag.Methods, method) {
			continue
		}

		// discard offline agents
		if !ag.IsConnected || ag.Client == nil {
			continue
		}

		candidates = append(candidates, ag)
	}
	agentsLock.RUnlock()

	if len(candidates) == 0 {
		return nil
	}

	// randomly address one agent
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(candidates))))
	if err != nil {
		return candidates[0]
	}
	return candidates[n.Int64()]
}

// RegisterAgent registers a newly connected agent WebSocket session,
// replacing (and closing) any previous session from the same agent. An error
// means the session couldn't be registered, and is to be closed by the caller.
func RegisterAgent(agentType, agentId string, client *wsrpc.WsRpcClient) error {
	ag := &KowabungaAgent{
		ID:          agentId,
		Client:      client,
		Type:        agentType,
		IsConnected: true,
	}

	// discover agent's RPC capabilities
	args := agents.CapabilitiesArgs{}
	var reply agents.CapabilitiesReply
	method := fmt.Sprintf("%s.Capabilities", ag.Type)
	err := ag.Client.Call(method, args, &reply)
	if err != nil {
		klog.Errorf("Unable to call remote Capabilities() RPC. Is it a legit agent ??")
		return err
	}
	ag.Version = reply.Version
	ag.Methods = reply.Methods

	klog.Infof("Registering %s agent %s WebSocket connection ...", agentType, agentId)
	agentsLock.Lock()
	previous := GetAgents()[agentId]
	GetAgents()[agentId] = ag
	agentsLock.Unlock()
	go ag.WatchKeepalive()

	if previous != nil {
		// close previous session, its watcher won't unregister the new one
		klog.Infof("A previous agent WebSocket connection from %s was referenced. Closing ...", agentId)
		err := previous.Client.Close()
		if err != nil {
			klog.Error(err)
		}
	}

	// the session is registered: post-registration failures are not fatal
	ag.onRegistered()
	return nil
}

// onRegistered (re)synchronizes the resources managed by a newly connected agent.
func (ag *KowabungaAgent) onRegistered() {
	agentId := ag.ID
	switch ag.Type {
	case common.KowabungaKiwiAgent:
		kiwis := FindKiwis()
		for _, k := range kiwis {
			if slices.Contains(k.AgentIDs, agentId) {
				err := k.Reload()
				if err != nil {
					klog.Errorf("unable to reload Kiwi %s: %v", k.String(), err)
				}
			}
		}
	case common.KowabungaKaktusAgent:
		kaktuses := FindKaktuses()
		for _, k := range kaktuses {
			if slices.Contains(k.Agents(), agentId) {
				go func() {
					err := k.Scan()
					if err != nil {
						klog.Error(err)
					}
				}()
				break
			}
		}

		pools := FindStoragePools()
		for _, p := range pools {
			if slices.Contains(p.Agents(), agentId) {
				go func() {
					err := p.Scan()
					if err != nil {
						klog.Error(err)
					}
				}()
				break
			}
		}
	case common.KowabungaControllerAgent:
		args := agents.KontrollerReloadArgs{}
		var reply agents.KontrollerReloadReply
		err := RPC([]string{agentId}, "Reload", args, &reply)
		if err != nil {
			klog.Errorf("unable to reload controller agent %s: %v", agentId, err)
		}
	}
}

// unregisterAgentSession removes the agent from registered ones, only if
// that very session is still the registered one.
func unregisterAgentSession(ag *KowabungaAgent) {
	agentsLock.Lock()
	defer agentsLock.Unlock()
	if current := GetAgents()[ag.ID]; current != nil && current.Client == ag.Client {
		delete(GetAgents(), ag.ID)
	}
}

func VerifyAgents(candidates []string, kind string) []string {
	agents := []string{}

	// add eligible agents
	for _, agentId := range candidates {
		a, err := FindAgentByID(agentId)
		if err != nil {
			klog.Errorf("invalid agent ID: %s (%v)", agentId, err)
			continue
		}
		if a.Type != kind {
			klog.Errorf("invalid agent type")
			continue
		}
		agents = append(agents, agentId)
	}

	return agents
}

func RPC(candidateAgents []string, method string, args, reply any) error {
	ag := GetEligibleAgent(candidateAgents, method)
	if ag == nil {
		return fmt.Errorf("RPC: can't find any eligible agent to perform such request")
	}

	methodName := fmt.Sprintf("%s.%s", ag.Type, method)
	return ag.Client.Call(methodName, args, reply)
}
