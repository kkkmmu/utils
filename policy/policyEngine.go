//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

// policyEngine.go
package policy

import (
	//"reflect"
	"sort"
	"strconv"
	"strings"
	"utils/netUtils"
	"utils/patriciaDB"
	"utils/policy/policyCommonDefs"
	//	"utils/commonDefs"
	"net"
	//	"asicdServices"
	//	"asicd/asicdConstDefs"
	"bytes"
	"fmt"
	//  "database/sql"
)

func (db *PolicyEngineDB) ActionListHasAction(actionList []PolicyAction, actionType int, action string) (match bool) {
	db.Logger.Info(fmt.Sprintln("ActionListHasAction for action ", action))
	return match
}

func (db *PolicyEngineDB) ActionNameListHasAction(actionList []string, actionType int, action string) (match bool) {
	db.Logger.Info(fmt.Sprintln("ActionListHasAction for action ", action))
	return match
}

func (db *PolicyEngineDB) PolicyEngineCheckActionsForEntity(entity PolicyEngineFilterEntityParams, policyConditionType int) (actionList []string) {
	db.Logger.Info(fmt.Sprintln("PolicyEngineTest to see if there are any policies for condition ", policyConditionType))
	var policyStmtList []string
	switch policyConditionType {
	case policyCommonDefs.PolicyConditionTypeDstIpPrefixMatch:
		break
	case policyCommonDefs.PolicyConditionTypeProtocolMatch:
		policyStmtList = db.ProtocolPolicyListDB[entity.RouteProtocol]
		break
	default:
		db.Logger.Err(fmt.Sprintln("Unknown conditonType"))
		return nil
	}
	if policyStmtList == nil || len(policyStmtList) == 0 {
		db.Logger.Info(fmt.Sprintln("no policy statements configured for this protocol"))
		return nil
	}
	for i := 0; i < len(policyStmtList); i++ {
		db.Logger.Info(fmt.Sprintln("Found policy stmt ", policyStmtList[i], " for this entity"))
		policyList := db.PolicyStmtPolicyMapDB[policyStmtList[i]]
		if policyList == nil || len(policyList) == 0 {
			db.Logger.Info(fmt.Sprintln("No policies configured for this entity"))
			return nil
		}
		for j := 0; j < len(policyList); j++ {
			db.Logger.Info(fmt.Sprintln("Found policy ", policyList[j], "for this statement"))
			policyStmtInfo := db.PolicyStmtDB.Get(patriciaDB.Prefix(policyStmtList[i]))
			if policyStmtInfo == nil {
				db.Logger.Info(fmt.Sprintln("Did not find this stmt in the DB"))
				return nil
			}
			policyStmt := policyStmtInfo.(PolicyStmt)
			if db.ConditionCheckValid(entity, policyStmt.Conditions, policyStmt) {
				db.Logger.Info(fmt.Sprintln("All conditions valid for this route, so this policy will be potentially applied to this route"))
				return policyStmt.Actions
			}
		}
	}
	return actionList
}
func (db *PolicyEngineDB) PolicyEngineUndoActionsPolicyStmt(policy Policy, policyStmt PolicyStmt, params interface{}, conditionsAndActionsList ConditionsAndActionsList) {
	db.Logger.Info(fmt.Sprintln("policyEngineUndoActionsPolicyStmt"))
	if conditionsAndActionsList.ActionList == nil {
		db.Logger.Info(fmt.Sprintln("No actions"))
		return
	}
	var i int
	conditionInfoList := make([]interface{}, 0)
	for j := 0; j < len(conditionsAndActionsList.ConditionList); j++ {
		conditionInfoList = append(conditionInfoList, conditionsAndActionsList.ConditionList[j].ConditionInfo)
	}

	for i = 0; i < len(conditionsAndActionsList.ActionList); i++ {
		db.Logger.Info(fmt.Sprintln("Find policy action number ", i, " name ", conditionsAndActionsList.ActionList[i], " in the action database"))
		/*
			actionItem := db.PolicyActionsDB.Get(patriciaDB.Prefix(policyStmt.Actions[i]))
			if actionItem == nil {
				db.Logger.Info(fmt.Sprintln("Did not find action ", conditionsAndActionsList.ActionList[i], " in the action database")
				continue
			}
			actionInfo := actionItem.(PolicyAction)
		*/
		policyAction := conditionsAndActionsList.ActionList[i]
		if db.UndoActionfuncMap[policyAction.ActionType] != nil {
			db.UndoActionfuncMap[policyAction.ActionType](policyAction.ActionInfo, conditionInfoList, params, policyStmt)
		}
	}
}
func (db *PolicyEngineDB) PolicyEngineUndoPolicyForEntity(entity PolicyEngineFilterEntityParams, policy Policy, params interface{}) {
	db.Logger.Info(fmt.Sprintln("policyEngineUndoPolicyForRoute - policy name ", policy.Name, "  route: ", entity.DestNetIp, " type:", entity.RouteProtocol))
	if db.GetPolicyEntityMapIndex == nil {
		return
	}
	policyEntityIndex := db.GetPolicyEntityMapIndex(entity, policy.Name)
	if policyEntityIndex == nil {
		db.Logger.Info(fmt.Sprintln("policy entity map index nil"))
		return
	}
	policyStmtMap := db.PolicyEntityMap[policyEntityIndex]
	if policyStmtMap.PolicyStmtMap == nil {
		db.Logger.Info(fmt.Sprintln("Unexpected:None of the policy statements of this policy have been applied on this route"))
		return
	}
	for stmt, conditionsAndActionsList := range policyStmtMap.PolicyStmtMap {
		db.Logger.Info(fmt.Sprintln("Applied policyStmtName ", stmt))
		policyStmt := db.PolicyStmtDB.Get(patriciaDB.Prefix(stmt))
		if policyStmt == nil {
			db.Logger.Info(fmt.Sprintln("Invalid policyStmt"))
			continue
		}
		db.PolicyEngineUndoActionsPolicyStmt(policy, policyStmt.(PolicyStmt), params, conditionsAndActionsList)
		//check if the route still exists - it may have been deleted by the previous statement action
		if db.IsEntityPresentFunc != nil {
			if !(db.IsEntityPresentFunc(params)) {
				db.Logger.Info(fmt.Sprintln("This entity no longer exists"))
				break
			}
		}
	}
}
func (db *PolicyEngineDB) PolicyEngineImplementActions(entity PolicyEngineFilterEntityParams, action PolicyAction,
	conditionInfoList []interface{}, params interface{}, policyStmt PolicyStmt) (policyActionList []PolicyAction) {
	db.Logger.Info(fmt.Sprintln("policyEngineImplementActions"))
	policyActionList = make([]PolicyAction, 0)
	addActionToList := false
	switch action.ActionType {
	case policyCommonDefs.PolicyActionTypeRouteDisposition, policyCommonDefs.PolicyActionTypeRouteRedistribute,
		policyCommonDefs.PolicyActionTypeNetworkStatementAdvertise, policyCommonDefs.PolicyActionTypeAggregate:
		if entity.DeletePath == true {
			db.Logger.Info(fmt.Sprintln("action to be reversed", action.ActionType))
			if db.UndoActionfuncMap[action.ActionType] != nil {
				db.UndoActionfuncMap[action.ActionType](action.ActionInfo, conditionInfoList, params, policyStmt)
			}
			addActionToList = true
		} else { //if entity.CreatePath == true or neither create/delete is valid - in case this function is called a a part of policy create{
			db.Logger.Info(fmt.Sprintln("action to be applied", action.ActionType))
			if db.ActionfuncMap[action.ActionType] != nil {
				db.ActionfuncMap[action.ActionType](action.ActionInfo, conditionInfoList, params)
			}
			addActionToList = true
		}
	default:
		db.Logger.Err(fmt.Sprintln("UnknownInvalid type of action"))
		break
	}
	if addActionToList == true {
		policyActionList = append(policyActionList, action)
	}
	return policyActionList
}

/*
func (db *PolicyEngineDB) FindPrefixMatch(ipAddr string, ipPrefix patriciaDB.Prefix, policyName string) (match bool) {
	db.Logger.Info(fmt.Sprintln("Prefix match policy ", policyName))
	policyListItem := db.PrefixPolicyListDB.GetLongestPrefixNode(ipPrefix)
	if policyListItem == nil {
		db.Logger.Info(fmt.Sprintln("intf stored at prefix ", ipPrefix, " is nil"))
		return false
	}
	if policyListItem != nil && reflect.TypeOf(policyListItem).Kind() != reflect.Slice {
		db.Logger.Err(fmt.Sprintln("Incorrect data type for this prefix "))
		return false
	}
	policyListSlice := reflect.ValueOf(policyListItem)
	for idx := 0; idx < policyListSlice.Len(); idx++ {
		prefixPolicyListInfo := policyListSlice.Index(idx).Interface().(PrefixPolicyListInfo)
		if prefixPolicyListInfo.policyName != policyName {
			db.Logger.Info(fmt.Sprintln("Found a potential match for this prefix but the policy ", policyName, " is not what we are looking for"))
			continue
		}
		if prefixPolicyListInfo.lowRange == -1 && prefixPolicyListInfo.highRange == -1 {
			db.Logger.Info(fmt.Sprintln("Looking for exact match condition for prefix ", prefixPolicyListInfo.ipPrefix))
			if bytes.Equal(ipPrefix, prefixPolicyListInfo.ipPrefix) {
				db.Logger.Info(fmt.Sprintln(" Matched the prefix"))
				return true
			} else {
				db.Logger.Info(fmt.Sprintln(" Did not match the exact prefix"))
				return false
			}
		}
		tempSlice := strings.Split(ipAddr, "/")
		maskLen, err := strconv.Atoi(tempSlice[1])
		if err != nil {
			db.Logger.Err(fmt.Sprintln("err getting maskLen"))
			return false
		}
		db.Logger.Info(fmt.Sprintln("Mask len = ", maskLen))
		if maskLen < prefixPolicyListInfo.lowRange || maskLen > prefixPolicyListInfo.highRange {
			db.Logger.Info(fmt.Sprintln("Mask range of the route ", maskLen, " not within the required mask range:", prefixPolicyListInfo.lowRange, "..", prefixPolicyListInfo.highRange))
			return false
		} else {
			db.Logger.Info(fmt.Sprintln("Mask range of the route ", maskLen, " within the required mask range:", prefixPolicyListInfo.lowRange, "..", prefixPolicyListInfo.highRange))
			return true
		}
	}
	return match
}*/
func (db *PolicyEngineDB) FindPrefixMatch(ipAddr string, ipPrefix patriciaDB.Prefix, condition PolicyCondition) (match bool) {
	db.Logger.Info(fmt.Sprintln("ipAddr : ", ipAddr, " ipPrefix: ", ipPrefix, " condition.IpPrefix: ", condition.ConditionInfo.(MatchPrefixConditionInfo).IpPrefix, " conditionInfo,MaskLengthRange: ", condition.ConditionInfo.(MatchPrefixConditionInfo).Prefix.IpPrefix))
	conditionInfo := condition.ConditionInfo.(MatchPrefixConditionInfo)
	if conditionInfo.LowRange == -1 && conditionInfo.HighRange == -1 {
		_, ipNet, err := net.ParseCIDR(condition.ConditionInfo.(MatchPrefixConditionInfo).Prefix.IpPrefix)
		if err != nil {
			return false
		}
		if bytes.Equal(ipPrefix, conditionInfo.IpPrefix) {
			db.Logger.Info(fmt.Sprintln(" Matched the prefix"))
			return true
		}
		networkMask := ipNet.Mask
		vdestMask := net.IPv4Mask(networkMask[0], networkMask[1], networkMask[2], networkMask[3])
		destIp := (net.IP(ipPrefix)).Mask(vdestMask)
		db.Logger.Info(fmt.Sprintln("networkMask: ", networkMask, " vdestMask: ", vdestMask, " destIp: ", destIp, "Looking for exact match condition for prefix ", conditionInfo.IpPrefix, " and ", destIp))
		if bytes.Equal(destIp, conditionInfo.IpPrefix) {
			db.Logger.Info(fmt.Sprintln(" Matched the prefix"))
			return true
		} else {
			db.Logger.Info(fmt.Sprintln(" Did not match the exact prefix"))
			return false
		}
	}
	tempSlice := strings.Split(ipAddr, "/")
	maskLen, err := strconv.Atoi(tempSlice[1])
	if err != nil {
		db.Logger.Err(fmt.Sprintln("err getting maskLen"))
		return false
	}
	db.Logger.Info(fmt.Sprintln("Mask len = ", maskLen))
	if maskLen < conditionInfo.LowRange || maskLen > conditionInfo.HighRange {
		db.Logger.Info(fmt.Sprintln("Mask range of the route ", maskLen, " not within the required mask range:", conditionInfo.LowRange, "-", conditionInfo.HighRange))
		return false
	} else {
		db.Logger.Info(fmt.Sprintln("Mask range of the route ", maskLen, " within the required mask range:", conditionInfo.LowRange, "-", conditionInfo.HighRange))
		return true
	}
	return match
}
func (db *PolicyEngineDB) DstIpPrefixMatchConditionfunc(entity PolicyEngineFilterEntityParams, condition PolicyCondition) (match bool) {
	db.Logger.Info(fmt.Sprintln("dstIpPrefixMatchConditionfunc"))
	ipPrefix, err := netUtils.GetNetworkPrefixFromCIDR(entity.DestNetIp)
	if err != nil {
		db.Logger.Info(fmt.Sprintln("Invalid ipPrefix for the route ", entity.DestNetIp))
		return false
	}
	match = db.FindPrefixMatch(entity.DestNetIp, ipPrefix, condition)
	if match {
		db.Logger.Info(fmt.Sprintln("Found a match for this prefix"))
	}
	return match
}
func (db *PolicyEngineDB) ProtocolMatchConditionfunc(entity PolicyEngineFilterEntityParams, condition PolicyCondition) (match bool) {
	db.Logger.Info(fmt.Sprintln("protocolMatchConditionfunc: check if policy protocol: ", condition.ConditionInfo.(string), " matches entity protocol: ", entity.RouteProtocol))
	matchProto := condition.ConditionInfo.(string)
	if matchProto == entity.RouteProtocol {
		db.Logger.Info(fmt.Sprintln("Protocol condition matches"))
		match = true
	}
	return match
}
func (db *PolicyEngineDB) ConditionCheckValid(entity PolicyEngineFilterEntityParams, conditionsList []string, policyStmt PolicyStmt) (valid bool) {
	db.Logger.Info(fmt.Sprintln("conditionCheckValid"))
	valid = true
	if conditionsList == nil {
		db.Logger.Info(fmt.Sprintln("No conditions to match, so valid"))
		return true
	}
	for i := 0; i < len(conditionsList); i++ {
		db.Logger.Info(fmt.Sprintln("Find policy condition number ", i, " name ", policyStmt.Conditions[i], " in the condition database"))
		conditionItem := db.PolicyConditionsDB.Get(patriciaDB.Prefix(conditionsList[i]))
		if conditionItem == nil {
			db.Logger.Info(fmt.Sprintln("Did not find condition ", conditionsList[i], " in the condition database"))
			continue
		}
		condition := conditionItem.(PolicyCondition)
		db.Logger.Info(fmt.Sprintln("policy condition number ", i, " type ", condition.ConditionType))
		if db.ConditionCheckfuncMap[condition.ConditionType] != nil {
			match := db.ConditionCheckfuncMap[condition.ConditionType](entity, condition)
			if !match {
				db.Logger.Info(fmt.Sprintln("Condition does not match"))
				return false
			}
		}
	}
	db.Logger.Info(fmt.Sprintln("returning valid= ", valid))
	return valid
}
func (db *PolicyEngineDB) PolicyEngineMatchConditions(entity PolicyEngineFilterEntityParams, conditions []string, matchConditions string) (match bool, conditionsList []PolicyCondition) {
	db.Logger.Info(fmt.Sprintln("policyEngineMatchConditions"))
	var i int
	allConditionsMatch := true
	anyConditionsMatch := false
	addConditiontoList := false
	conditionsList = make([]PolicyCondition, 0)
	for i = 0; i < len(conditions); i++ {
		addConditiontoList = false
		db.Logger.Info(fmt.Sprintln("Find policy condition number ", i, " name ", conditions[i], " in the condition database"))
		conditionItem := db.PolicyConditionsDB.Get(patriciaDB.Prefix(conditions[i]))
		if conditionItem == nil {
			db.Logger.Info(fmt.Sprintln("Did not find condition ", conditions[i], " in the condition database"))
			continue
		}
		condition := conditionItem.(PolicyCondition)
		db.Logger.Info(fmt.Sprintln("policy condition number ", i, "  type ", condition.ConditionType))
		if db.ConditionCheckfuncMap[condition.ConditionType] != nil {
			match = db.ConditionCheckfuncMap[condition.ConditionType](entity, condition)
			if match {
				db.Logger.Info(fmt.Sprintln("Condition match found"))
				anyConditionsMatch = true
				addConditiontoList = true
			} else {
				allConditionsMatch = false
			}
		}
		if addConditiontoList == true {
			conditionsList = append(conditionsList, condition)
		}
	}
	if matchConditions == "all" && allConditionsMatch == true {
		return true, conditionsList
	}
	if matchConditions == "any" && anyConditionsMatch == true {
		return true, conditionsList
	}
	return match, conditionsList
}
func (db *PolicyEngineDB) PolicyEngineApplyPolicyStmt(entity *PolicyEngineFilterEntityParams, info ApplyPolicyInfo,
	policyStmt PolicyStmt, policyPath int, params interface{}, hit *bool, deleted *bool) {
	policy := info.ApplyPolicy
	db.Logger.Info(fmt.Sprintln("policyEngineApplyPolicyStmt - ", policyStmt.Name))
	var conditionList []PolicyCondition
	conditionInfoList := make([]interface{}, 0)
	var match bool
	if policyStmt.Conditions == nil && info.Conditions == nil {
		db.Logger.Info(fmt.Sprintln("No policy conditions"))
		*hit = true
	} else {
		//match, ret_conditionList := db.PolicyEngineMatchConditions(*entity, policyStmt)
		match, conditionList = db.PolicyEngineMatchConditions(*entity, policyStmt.Conditions, policyStmt.MatchConditions)
		db.Logger.Info(fmt.Sprintln("match = ", match))
		*hit = match
		if !match {
			db.Logger.Info(fmt.Sprintln("Stmt Conditions do not match"))
			return
		}
		for j := 0; j < len(conditionList); j++ {
			conditionInfoList = append(conditionInfoList, conditionList[j].ConditionInfo)
		}
		match, conditionList = db.PolicyEngineMatchConditions(*entity, info.Conditions, "all")
		db.Logger.Info(fmt.Sprintln("match = ", match))
		*hit = match
		if !match {
			db.Logger.Info(fmt.Sprintln("Extra Conditions do not match"))
			return
		}
		for j := 0; j < len(conditionList); j++ {
			conditionInfoList = append(conditionInfoList, conditionList[j].ConditionInfo)
		}
	}
	actionList := db.PolicyEngineImplementActions(*entity, info.Action, conditionInfoList, params, policyStmt)
	if db.ActionListHasAction(actionList, policyCommonDefs.PolicyActionTypeRouteDisposition, "Reject") {
		db.Logger.Info(fmt.Sprintln("Reject action was applied for this entity"))
		*deleted = true
	}
	//check if the route still exists - it may have been deleted by the previous statement action
	if db.IsEntityPresentFunc != nil {
		*deleted = !(db.IsEntityPresentFunc(params))
	}
	db.AddPolicyEntityMapEntry(*entity, policy.Name, policyStmt.Name, conditionList, actionList)
	if db.UpdateEntityDB != nil {
		policyDetails := PolicyDetails{Policy: policy.Name, PolicyStmt: policyStmt.Name, ConditionList: conditionList, ActionList: actionList, EntityDeleted: *deleted}
		db.UpdateEntityDB(policyDetails, params)
	}
}

func (db *PolicyEngineDB) PolicyEngineApplyPolicy(entity *PolicyEngineFilterEntityParams, info ApplyPolicyInfo, policyPath int, params interface{}, hit *bool) {
	db.Logger.Info(fmt.Sprintln("policyEngineApplyPolicy - ", info.ApplyPolicy.Name))
	policy := info.ApplyPolicy
	var policyStmtKeys []int
	deleted := false
	for k := range policy.PolicyStmtPrecedenceMap {
		db.Logger.Info(fmt.Sprintln("key k = ", k))
		policyStmtKeys = append(policyStmtKeys, k)
	}
	sort.Ints(policyStmtKeys)
	for i := 0; i < len(policyStmtKeys); i++ {
		db.Logger.Info(fmt.Sprintln("Key: ", policyStmtKeys[i], " policyStmtName ", policy.PolicyStmtPrecedenceMap[policyStmtKeys[i]]))
		policyStmt := db.PolicyStmtDB.Get((patriciaDB.Prefix(policy.PolicyStmtPrecedenceMap[policyStmtKeys[i]])))
		if policyStmt == nil {
			db.Logger.Info(fmt.Sprintln("Invalid policyStmt"))
			continue
		}
		db.PolicyEngineApplyPolicyStmt(entity, info, policyStmt.(PolicyStmt), policyPath, params, hit, &deleted)
		if deleted == true {
			db.Logger.Info(fmt.Sprintln("Entity was deleted as a part of the policyStmt ", policy.PolicyStmtPrecedenceMap[policyStmtKeys[i]]))
			break
		}
		if *hit == true {
			if policy.MatchType == "any" {
				db.Logger.Info(fmt.Sprintln("Match type for policy ", policy.Name, " is any and the policy stmt ", (policyStmt.(PolicyStmt)).Name, " is a hit, no more policy statements will be executed"))
				break
			}
		}
	}
}
func (db *PolicyEngineDB) PolicyEngineApplyForEntity(entity PolicyEngineFilterEntityParams, policyData interface{}, params interface{}) {
	db.Logger.Info(fmt.Sprintln("policyEngineApplyForEntity"))
	info := policyData.(ApplyPolicyInfo)
	policy := info.ApplyPolicy
	policyHit := false
	if len(entity.PolicyList) == 0 {
		db.Logger.Info(fmt.Sprintln("This route has no policy applied to it so far, just apply the new policy"))
		db.PolicyEngineApplyPolicy(&entity, info, policyCommonDefs.PolicyPath_All, params, &policyHit)
	} else {
		db.Logger.Info(fmt.Sprintln("This route already has policy applied to it - len(route.PolicyList) - ", len(entity.PolicyList)))

		for i := 0; i < len(entity.PolicyList); i++ {
			db.Logger.Info(fmt.Sprintln("policy at index ", i))
			policyInfo := db.PolicyDB.Get(patriciaDB.Prefix(entity.PolicyList[i]))
			if policyInfo == nil {
				db.Logger.Info(fmt.Sprintln("Unexpected: Invalid policy in the route policy list"))
			} else {
				oldPolicy := policyInfo.(Policy)
				if !isPolicyTypeSame(oldPolicy, policy) {
					db.Logger.Info(fmt.Sprintln("The policy type applied currently is not the same as new policy, so apply new policy"))
					db.PolicyEngineApplyPolicy(&entity, info, policyCommonDefs.PolicyPath_All, params, &policyHit)
				} else if oldPolicy.Precedence < policy.Precedence {
					db.Logger.Info(fmt.Sprintln("The policy types are same and precedence of the policy applied currently is lower than the new policy, so do nothing"))
					return
				} else {
					db.Logger.Info(fmt.Sprintln("The new policy's precedence is lower, so undo old policy's actions and apply the new policy"))
					//db.PolicyEngineUndoPolicyForEntity(entity, oldPolicy, params)
					db.PolicyEngineApplyPolicy(&entity, info, policyCommonDefs.PolicyPath_All, params, &policyHit)
				}
			}
		}
	}
}
func (db *PolicyEngineDB) PolicyEngineReverseGlobalPolicyStmt(policy Policy, policyStmt PolicyStmt) {
	db.Logger.Info(fmt.Sprintln("policyEngineApplyGlobalPolicyStmt - ", policyStmt.Name))
	var conditionItem interface{} = nil
	//global policies can only have statements with 1 condition and 1 action
	if policyStmt.Actions == nil {
		db.Logger.Info(fmt.Sprintln("No policy actions defined"))
		return
	}
	if policyStmt.Conditions == nil {
		db.Logger.Info(fmt.Sprintln("No policy conditions"))
	} else {
		if len(policyStmt.Conditions) > 1 {
			db.Logger.Info(fmt.Sprintln("only 1 condition allowed for global policy stmt"))
			return
		}
		conditionItem = db.PolicyConditionsDB.Get(patriciaDB.Prefix(policyStmt.Conditions[0]))
		if conditionItem == nil {
			db.Logger.Info(fmt.Sprintln("Condition ", policyStmt.Conditions[0], " not found"))
			return
		}
		actionItem := db.PolicyActionsDB.Get(patriciaDB.Prefix(policyStmt.Actions[0]))
		if actionItem == nil {
			db.Logger.Info(fmt.Sprintln("Action ", policyStmt.Actions[0], " not found"))
			return
		}
		actionInfo := actionItem.(PolicyAction)
		if db.UndoActionfuncMap[actionInfo.ActionType] != nil {
			//since global policies have just 1 condition, we can pass that as the params to the undo call
			db.UndoActionfuncMap[actionInfo.ActionType](actionItem, nil, conditionItem, policyStmt)
		}
	}
}
func (db *PolicyEngineDB) PolicyEngineApplyGlobalPolicyStmt(policy Policy, policyStmt PolicyStmt) {
	db.Logger.Info(fmt.Sprintln("policyEngineApplyGlobalPolicyStmt - ", policyStmt.Name))
	var conditionItem interface{} = nil
	//global policies can only have statements with 1 condition and 1 action
	if policyStmt.Actions == nil {
		db.Logger.Info(fmt.Sprintln("No policy actions defined"))
		return
	}
	if policyStmt.Conditions == nil {
		db.Logger.Info(fmt.Sprintln("No policy conditions"))
	} else {
		if len(policyStmt.Conditions) > 1 {
			db.Logger.Info(fmt.Sprintln("only 1 condition allowed for global policy stmt"))
			return
		}
		conditionItem = db.PolicyConditionsDB.Get(patriciaDB.Prefix(policyStmt.Conditions[0]))
		if conditionItem == nil {
			db.Logger.Info(fmt.Sprintln("Condition ", policyStmt.Conditions[0], " not found"))
			return
		}
		policyCondition := conditionItem.(PolicyCondition)
		conditionInfoList := make([]interface{}, 0)
		conditionInfoList = append(conditionInfoList, policyCondition.ConditionInfo)

		actionItem := db.PolicyActionsDB.Get(patriciaDB.Prefix(policyStmt.Actions[0]))
		if actionItem == nil {
			db.Logger.Info(fmt.Sprintln("Action ", policyStmt.Actions[0], " not found"))
			return
		}
		actionInfo := actionItem.(PolicyAction)
		if db.ActionfuncMap[actionInfo.ActionType] != nil {
			db.ActionfuncMap[actionInfo.ActionType](actionInfo.ActionInfo, conditionInfoList, nil)
		}
	}
}
func (db *PolicyEngineDB) PolicyEngineReverseGlobalPolicy(policy Policy) {
	db.Logger.Info(fmt.Sprintln("policyEngineReverseGlobalPolicy"))
	var policyStmtKeys []int
	for k := range policy.PolicyStmtPrecedenceMap {
		db.Logger.Info(fmt.Sprintln("key k = ", k))
		policyStmtKeys = append(policyStmtKeys, k)
	}
	sort.Ints(policyStmtKeys)
	for i := 0; i < len(policyStmtKeys); i++ {
		db.Logger.Info(fmt.Sprintln("Key: ", policyStmtKeys[i], " policyStmtName ", policy.PolicyStmtPrecedenceMap[policyStmtKeys[i]]))
		policyStmt := db.PolicyStmtDB.Get((patriciaDB.Prefix(policy.PolicyStmtPrecedenceMap[policyStmtKeys[i]])))
		if policyStmt == nil {
			db.Logger.Info(fmt.Sprintln("Invalid policyStmt"))
			continue
		}
		db.PolicyEngineReverseGlobalPolicyStmt(policy, policyStmt.(PolicyStmt))
	}
}
func (db *PolicyEngineDB) PolicyEngineApplyGlobalPolicy(policy Policy) {
	db.Logger.Info(fmt.Sprintln("policyEngineApplyGlobalPolicy"))
	var policyStmtKeys []int
	for k := range policy.PolicyStmtPrecedenceMap {
		db.Logger.Info(fmt.Sprintln("key k = ", k))
		policyStmtKeys = append(policyStmtKeys, k)
	}
	sort.Ints(policyStmtKeys)
	for i := 0; i < len(policyStmtKeys); i++ {
		db.Logger.Info(fmt.Sprintln("Key: ", policyStmtKeys[i], " policyStmtName ", policy.PolicyStmtPrecedenceMap[policyStmtKeys[i]]))
		policyStmt := db.PolicyStmtDB.Get((patriciaDB.Prefix(policy.PolicyStmtPrecedenceMap[policyStmtKeys[i]])))
		if policyStmt == nil {
			db.Logger.Info(fmt.Sprintln("Invalid policyStmt"))
			continue
		}
		db.PolicyEngineApplyGlobalPolicyStmt(policy, policyStmt.(PolicyStmt))
	}
}

func (db *PolicyEngineDB) PolicyEngineTraverseAndApplyPolicy(info ApplyPolicyInfo) {
	db.Logger.Info(fmt.Sprintln("PolicyEngineTraverseAndApplyPolicy -  apply policy ", info.ApplyPolicy.Name))
	if db.TraverseAndApplyPolicyFunc != nil {
		db.Logger.Info(fmt.Sprintln("Calling TraverseAndApplyPolicyFunc function"))
		db.TraverseAndApplyPolicyFunc(info, db.PolicyEngineApplyForEntity)
	}
	/*	if policy.ExportPolicy || policy.ImportPolicy {
			db.Logger.Info(fmt.Sprintln("Applying import/export policy to all routes"))
			if db.TraverseAndApplyPolicyFunc != nil {
				db.Logger.Info(fmt.Sprintln("Calling TraverseAndApplyPolicyFunc function"))
				db.TraverseAndApplyPolicyFunc(policy, db.PolicyEngineApplyForEntity)
			}
		} else if policy.GlobalPolicy {
			db.Logger.Info(fmt.Sprintln("Need to apply global policy"))
			db.PolicyEngineApplyGlobalPolicy(policy)
		}*/
}

func (db *PolicyEngineDB) PolicyEngineTraverseAndReversePolicy(policy Policy) {
	db.Logger.Info(fmt.Sprintln("PolicyEngineTraverseAndReversePolicy -  reverse policy ", policy.Name))
	if policy.ExportPolicy || policy.ImportPolicy {
		db.Logger.Info(fmt.Sprintln("Reversing import/export policy "))
		db.TraverseAndReversePolicyFunc(policy)
	} else if policy.GlobalPolicy {
		db.Logger.Info(fmt.Sprintln("Need to reverse global policy"))
		db.PolicyEngineReverseGlobalPolicy(policy)
	}
}

func (db *PolicyEngineDB) PolicyEngineFilter(entity PolicyEngineFilterEntityParams, policyPath int, params interface{}) {
	/*db.Logger.Info(fmt.Sprintln("PolicyEngineFilter"))
	var policyPath_Str string
	if policyPath == policyCommonDefs.PolicyPath_Import {
		policyPath_Str = "Import"
	} else if policyPath == policyCommonDefs.PolicyPath_Export {
		policyPath_Str = "Export"
	} else if policyPath == policyCommonDefs.PolicyPath_All {
		policyPath_Str = "ALL"
		db.Logger.Info(fmt.Sprintln("policy path ", policyPath_Str, " unexpected in this function"))
		return
	}
	db.Logger.Info(fmt.Sprintln("PolicyEngineFilter for policypath ", policyPath_Str, "create = ", entity.CreatePath, " delete = ", entity.DeletePath, " route: ", entity.DestNetIp, " protocol type: ", entity.RouteProtocol))*/
	var policyKeys []int
	var policyHit bool
	idx := 0
	var policyInfo interface{}
	if policyPath == policyCommonDefs.PolicyPath_Import {
		for k := range db.ImportPolicyPrecedenceMap {
			policyKeys = append(policyKeys, k)
		}
	} else if policyPath == policyCommonDefs.PolicyPath_Export {
		for k := range db.ExportPolicyPrecedenceMap {
			policyKeys = append(policyKeys, k)
		}
	}
	sort.Ints(policyKeys)
	for {
		if entity.DeletePath == true { //policyEngineFilter called during delete
			if entity.PolicyList != nil {
				if idx >= len(entity.PolicyList) {
					break
				}
				//db.Logger.Info(fmt.Sprintln("getting policy ", idx, " from entity.PolicyList"))
				policyInfo = db.PolicyDB.Get(patriciaDB.Prefix(entity.PolicyList[idx]))
				idx++
				if policyInfo.(Policy).ExportPolicy && policyPath == policyCommonDefs.PolicyPath_Import || policyInfo.(Policy).ImportPolicy && policyPath == policyCommonDefs.PolicyPath_Export {
					//		db.Logger.Info(fmt.Sprintln("policy ", policyInfo.(Policy).Name, " not the same type as the policypath -", policyPath_Str))
					continue
				}
			} else {
				//db.Logger.Info(fmt.Sprintln("PolicyList empty and this is a delete operation, so break"))
				break
			}
		} else if entity.CreatePath == true { //policyEngine filter called during create
			//db.Logger.Info(fmt.Sprintln("idx = ", idx, " len(policyKeys):", len(policyKeys)))
			if idx >= len(policyKeys) {
				break
			}
			policyName := ""
			if policyPath == policyCommonDefs.PolicyPath_Import {
				policyName = db.ImportPolicyPrecedenceMap[policyKeys[idx]]
			} else if policyPath == policyCommonDefs.PolicyPath_Export {
				policyName = db.ExportPolicyPrecedenceMap[policyKeys[idx]]
			}
			//db.Logger.Info(fmt.Sprintln("getting policy  ", idx, " policyKeys[idx] = ", policyKeys[idx], " ", policyName, " from PolicyDB"))
			policyInfo = db.PolicyDB.Get((patriciaDB.Prefix(policyName)))
			idx++
		}
		if policyInfo == nil {
			db.Logger.Info(fmt.Sprintln("Nil policy"))
			break
		}
		policy := policyInfo.(Policy)
		localPolicyDB := *db.LocalPolicyDB
		if localPolicyDB != nil && localPolicyDB[policy.LocalDBSliceIdx].IsValid == false {
			//db.Logger.Info(fmt.Sprintln("Invalid policy at localDB slice idx ", policy.LocalDBSliceIdx))
			continue
		}
		applyList := db.ApplyPolicyMap[policy.Name]
		if applyList == nil {
			//db.Logger.Info(fmt.Sprintln("no application for this policy ", policy.Name))
			continue
		}
		for j := 0; j < len(applyList); j++ {
			db.PolicyEngineApplyPolicy(&entity, applyList[j], policyPath, params, &policyHit)
			if policyHit {
				//db.Logger.Info(fmt.Sprintln("Policy ", policy.Name, " applied to the route"))
				break
			}
		}
	}
	if entity.PolicyHitCounter == 0 {
		//db.Logger.Info(fmt.Sprintln("Need to apply default policy, policyPath = ", policyPath, "policyPath_Str= ", policyPath_Str))
		if policyPath == policyCommonDefs.PolicyPath_Import {
			//db.Logger.Info(fmt.Sprintln("Applying default import policy"))
			if db.DefaultImportPolicyActionFunc != nil {
				db.DefaultImportPolicyActionFunc(nil, nil, params)
			}
		} else if policyPath == policyCommonDefs.PolicyPath_Export {
			//db.Logger.Info(fmt.Sprintln("Applying default export policy"))
			if db.DefaultExportPolicyActionFunc != nil {
				db.DefaultExportPolicyActionFunc(nil, nil, params)
			}
		}
	}
	if entity.DeletePath == true {
		db.DeletePolicyEntityMapEntry(entity, "")
	}
}
