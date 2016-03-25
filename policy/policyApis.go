// policyApis.go
package policy

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"utils/netUtils"
	"utils/patriciaDB"
	"utils/policy/policyCommonDefs"
)

type PolicyStmt struct { //policy engine uses this
	Name            string
	Precedence      int
	MatchConditions string
	Conditions      []string
	Actions         []string
	PolicyList      []string
	LocalDBSliceIdx int8
	ImportStmt      bool
	ExportStmt      bool
	GlobalStmt      bool
}
type PolicyStmtConfig struct {
	Name            string
	AdminState      string
	MatchConditions string
	Conditions      []string
	Actions         []string
}

type Policy struct {
	Name                    string
	Precedence              int
	MatchType               string
	PolicyStmtPrecedenceMap map[int]string
	LocalDBSliceIdx         int8
	ImportPolicy            bool
	ExportPolicy            bool
	GlobalPolicy            bool
	Extensions              interface{}
}

type PolicyDefinitionStmtPrecedence struct {
	Precedence int
	Statement  string
}
type PolicyDefinitionConfig struct {
	Name                       string
	Precedence                 int
	MatchType                  string
	PolicyDefinitionStatements []PolicyDefinitionStmtPrecedence
	Export                     bool
	Import                     bool
	Global                     bool
	Extensions                 interface{}
}

type PrefixPolicyListInfo struct {
	ipPrefix   patriciaDB.Prefix
	policyName string
	lowRange   int
	highRange  int
}

func validMatchConditions(matchConditionStr string) (valid bool) {
	fmt.Println("validMatchConditions for string ", matchConditionStr)
	if matchConditionStr == "any" || matchConditionStr == "all" {
		valid = true
	}
	return valid
}
func (db *PolicyEngineDB) UpdateProtocolPolicyTable(protoType string, name string, op int) {
	db.Logger.Printf("updateProtocolPolicyTable for protocol %d policy name %s op %d\n", protoType, name, op)
	var i int
	policyList := db.ProtocolPolicyListDB[protoType]
	if policyList == nil {
		if op == del {
			db.Logger.Println("Cannot find the policy map for this protocol, so cannot delete")
			return
		}
		policyList = make([]string, 0)
	}
	if op == add {
		policyList = append(policyList, name)
	}
	found := false
	if op == del {
		for i = 0; i < len(policyList); i++ {
			if policyList[i] == name {
				db.Logger.Println("Found the policy in the protocol policy table, deleting it")
				found = true
				break
			}
		}
		if found {
			policyList = append(policyList[:i], policyList[i+1:]...)
		}
	}
	db.ProtocolPolicyListDB[protoType] = policyList
}
func (db *PolicyEngineDB) UpdatePrefixPolicyTableWithPrefix(ipAddr string, name string, op int, lowRange int, highRange int) {
	db.Logger.Println("updatePrefixPolicyTableWithPrefix ", ipAddr)
	var i int
	ipPrefix, err := netUtils.GetNetworkPrefixFromCIDR(ipAddr)
	if err != nil {
		db.Logger.Println("ipPrefix invalid ")
		return
	}
	var policyList []PrefixPolicyListInfo
	var prefixPolicyListInfo PrefixPolicyListInfo
	policyListItem := db.PrefixPolicyListDB.Get(ipPrefix)
	if policyListItem != nil && reflect.TypeOf(policyListItem).Kind() != reflect.Slice {
		db.Logger.Println("Incorrect data type for this prefix ")
		return
	}
	if policyListItem == nil {
		if op == del {
			db.Logger.Println("Cannot find the policy map for this prefix, so cannot delete")
			return
		}
		policyList = make([]PrefixPolicyListInfo, 0)
	} else {
		policyListSlice := reflect.ValueOf(policyListItem)
		policyList = make([]PrefixPolicyListInfo, 0)
		for i = 0; i < policyListSlice.Len(); i++ {
			policyList = append(policyList, policyListSlice.Index(i).Interface().(PrefixPolicyListInfo))
		}
	}
	if op == add {
		prefixPolicyListInfo.ipPrefix = ipPrefix
		prefixPolicyListInfo.policyName = name
		prefixPolicyListInfo.lowRange = lowRange
		prefixPolicyListInfo.highRange = highRange
		policyList = append(policyList, prefixPolicyListInfo)
	}
	found := false
	if op == del {
		for i = 0; i < len(policyList); i++ {
			if policyList[i].policyName == name {
				db.Logger.Println("Found the policy in the prefix policy table, deleting it")
				break
			}
		}
		if found {
			policyList = append(policyList[:i], policyList[i+1:]...)
		}
	}
	db.PrefixPolicyListDB.Set(ipPrefix, policyList)
}
func (db *PolicyEngineDB) UpdatePrefixPolicyTableWithMaskRange(ipAddr string, masklength string, name string, op int) {
	db.Logger.Println("updatePrefixPolicyTableWithMaskRange")
	maskList := strings.Split(masklength, "..")
	if len(maskList) != 2 {
		db.Logger.Println("Invalid masklength range")
		return
	}
	lowRange, err := strconv.Atoi(maskList[0])
	if err != nil {
		db.Logger.Println("maskList[0] not valid")
		return
	}
	highRange, err := strconv.Atoi(maskList[1])
	if err != nil {
		db.Logger.Println("maskList[1] not valid")
		return
	}
	db.Logger.Println("lowRange = ", lowRange, " highrange = ", highRange)
	db.UpdatePrefixPolicyTableWithPrefix(ipAddr, name, op, lowRange, highRange)
	/*		for idx := lowRange;idx<highRange;idx ++ {
			ipMask:= net.CIDRMask(idx, 32)
			ipMaskStr := net.IP(ipMask).String()
			db.Logger.Println("idx ", idx, "ipMaskStr = ", ipMaskStr)
			ipPrefix, err := getNetowrkPrefixFromStrings(ipAddrStr, ipMaskStr)
			if err != nil {
				db.Logger.Println("Invalid prefix")
				return
			}
			updatePrefixPolicyTableWithPrefix(ipPrefix, name, op,lowRange,highRange)
		}*/
}
func (db *PolicyEngineDB) UpdatePrefixPolicyTableWithPrefixSet(prefixSet string, name string, op int) {
	db.Logger.Println("updatePrefixPolicyTableWithPrefixSet")
}
func (db *PolicyEngineDB) UpdatePrefixPolicyTable(conditionInfo interface{}, name string, op int) {
	condition := conditionInfo.(MatchPrefixConditionInfo)
	db.Logger.Printf("updatePrefixPolicyTable for prefixSet %s prefix %s policy name %s op %d\n", condition.PrefixSet, condition.Prefix, name, op)
	if condition.UsePrefixSet {
		db.Logger.Println("Need to look up Prefix set to get the prefixes")
		db.UpdatePrefixPolicyTableWithPrefixSet(condition.PrefixSet, name, op)
	} else {
		if condition.Prefix.MasklengthRange == "exact" {
			/*ipPrefix, err := getNetworkPrefixFromCIDR(condition.prefix.IpPrefix)
			   if err != nil {
				db.Logger.Println("ipPrefix invalid ")
				return
			   }*/
			db.UpdatePrefixPolicyTableWithPrefix(condition.Prefix.IpPrefix, name, op, -1, -1)
		} else {
			db.Logger.Println("Masklength= ", condition.Prefix.MasklengthRange)
			db.UpdatePrefixPolicyTableWithMaskRange(condition.Prefix.IpPrefix, condition.Prefix.MasklengthRange, name, op)
		}
	}
}
func (db *PolicyEngineDB) UpdateStatements(policy Policy, stmt PolicyStmt, op int) (err error) {
	db.Logger.Println("UpdateStatements for stmt ", stmt.Name)
	var i int
	if stmt.PolicyList == nil {
		if op == del {
			db.Logger.Println("stmt.PolicyList nil")
			return err
		}
		stmt.PolicyList = make([]string, 0)
	}
	if op == add {
		stmt.PolicyList = append(stmt.PolicyList, policy.Name)
	}
	found := false
	if op == del {
		for i = 0; i < len(stmt.PolicyList); i++ {
			if stmt.PolicyList[i] == policy.Name {
				db.Logger.Println("Found the policy in the policy stmt table, deleting it")
				found = true
				break
			}
		}
		if found {
			stmt.PolicyList = append(stmt.PolicyList[:i], stmt.PolicyList[i+1:]...)
		}
	}
	db.PolicyStmtDB.Set(patriciaDB.Prefix(stmt.Name), stmt)
	return err
}

func (db *PolicyEngineDB) UpdateGlobalStatementTable(policy string, stmt string, op int) (err error) {
	db.Logger.Println("updateGlobalStatementTablestmt ", stmt, " with policy ", policy)
	var i int
	policyList := db.PolicyStmtPolicyMapDB[stmt]
	if policyList == nil {
		if op == del {
			db.Logger.Println("Cannot find the policy map for this stmt, so cannot delete")
			err = errors.New("Cannot find the policy map for this stmt, so cannot delete")
			return err
		}
		policyList = make([]string, 0)
	}
	if op == add {
		policyList = append(policyList, policy)
	}
	found := false
	if op == del {
		for i = 0; i < len(policyList); i++ {
			if policyList[i] == policy {
				db.Logger.Println("Found the policy in the policy stmt table, deleting it")
				found = true
				break
			}
		}
		if found {
			policyList = append(policyList[:i], policyList[i+1:]...)
		}
	}
	db.PolicyStmtPolicyMapDB[stmt] = policyList
	return err
}
func (db *PolicyEngineDB) UpdateConditions(policyStmt PolicyStmt, conditionName string, op int) (err error) {
	db.Logger.Println("updateConditions for condition ", conditionName)
	conditionItem := db.PolicyConditionsDB.Get(patriciaDB.Prefix(conditionName))
	if conditionItem != nil {
		condition := conditionItem.(PolicyCondition)
		switch condition.ConditionType {
		case policyCommonDefs.PolicyConditionTypeProtocolMatch:
			db.Logger.Println("PolicyConditionTypeProtocolMatch")
			db.UpdateProtocolPolicyTable(condition.ConditionInfo.(string), policyStmt.Name, op)
			break
		case policyCommonDefs.PolicyConditionTypeDstIpPrefixMatch:
			db.Logger.Println("PolicyConditionTypeDstIpPrefixMatch")
			db.UpdatePrefixPolicyTable(condition.ConditionInfo, policyStmt.Name, op)
			break
		}
		if condition.PolicyStmtList == nil {
			condition.PolicyStmtList = make([]string, 0)
		}
		condition.PolicyStmtList = append(condition.PolicyStmtList, policyStmt.Name)
		db.Logger.Println("Adding policy ", policyStmt.Name, "to condition ", conditionName)
		db.PolicyConditionsDB.Set(patriciaDB.Prefix(conditionName), condition)
	} else {
		db.Logger.Println("Condition name ", conditionName, " not defined")
		err = errors.New("Condition name not defined")
	}
	return err
}

func (db *PolicyEngineDB) UpdateActions(policyStmt PolicyStmt, action PolicyAction, op int) (err error) {
	db.Logger.Println("updateActions for action ", action.Name)
	var i int
	if action.PolicyStmtList == nil {
		if op == del {
			db.Logger.Println("action.PolicyStmtList empty")
			err = errors.New("action.PolicyStmtLisy Empty")
			return err
		}
		action.PolicyStmtList = make([]string, 0)
	}
	if op == add {
		action.PolicyStmtList = append(action.PolicyStmtList, policyStmt.Name)
	}
	found := false
	if op == del {
		for i = 0; i < len(action.PolicyStmtList); i++ {
			if action.PolicyStmtList[i] == policyStmt.Name {
				db.Logger.Println("Found the policyStmt in the action's list, deleting it")
				found = true
				break
			}
		}
		if found {
			action.PolicyStmtList = append(action.PolicyStmtList[:i], action.PolicyStmtList[i+1:]...)
		}
	}

	db.PolicyActionsDB.Set(patriciaDB.Prefix(action.Name), action)
	return err
}

func (db *PolicyEngineDB) CreatePolicyStatement(cfg PolicyStmtConfig) (err error) {
	db.Logger.Println("CreatePolicyStatement")
	policyStmt := db.PolicyStmtDB.Get(patriciaDB.Prefix(cfg.Name))
	var i int
	if policyStmt == nil {
		db.Logger.Println("Defining a new policy statement with name ", cfg.Name)
		var newPolicyStmt PolicyStmt
		newPolicyStmt.Name = cfg.Name
		if !validMatchConditions(cfg.MatchConditions) {
			db.Logger.Println("Invalid match conditions - try any/all")
			err = errors.New("Invalid match conditions - try any/all")
			return err
		}
		newPolicyStmt.MatchConditions = cfg.MatchConditions
		if len(cfg.Conditions) > 0 {
			db.Logger.Println("Policy Statement has %d ", len(cfg.Conditions), " number of conditions")
			newPolicyStmt.Conditions = make([]string, 0)
			for i = 0; i < len(cfg.Conditions); i++ {
				newPolicyStmt.Conditions = append(newPolicyStmt.Conditions, cfg.Conditions[i])
				err = db.UpdateConditions(newPolicyStmt, cfg.Conditions[i], add)
				if err != nil {
					db.Logger.Println("updateConditions returned err ", err)
					err = errors.New("Error with updateConditions")
					return err
				}
			}
		}
		if len(cfg.Actions) > 0 {
			db.Logger.Println("Policy Statement has %d ", len(cfg.Actions), " number of actions")
			if len(cfg.Actions) > 1 {
				db.Logger.Println("Cannot have more than 1 action in a policy")
				err = errors.New("Cannot have more than 1 action in a policy")
				return err
			}
			var action PolicyAction
			newPolicyStmt.Actions = make([]string, 0)
			for i = 0; i < len(cfg.Actions); i++ {
				actionItem := db.PolicyActionsDB.Get(patriciaDB.Prefix(cfg.Actions[i]))
				if actionItem != nil {
					action = actionItem.(PolicyAction)
				} else {
					db.Logger.Println("action name ", cfg.Actions[i], " not defined")
					err = errors.New("action name not defined")
				}
				newPolicyStmt.ExportStmt, newPolicyStmt.ImportStmt, newPolicyStmt.GlobalStmt = db.PolicyActionType(action.ActionType)
				newPolicyStmt.Actions = append(newPolicyStmt.Actions, cfg.Actions[i])
				err = db.UpdateActions(newPolicyStmt, action, add)
				if err != nil {
					db.Logger.Println("updateActions returned err ", err)
					err = errors.New("Update actions returned err")
					return err
				}
			}
		}
		newPolicyStmt.LocalDBSliceIdx = int8(len(*db.LocalPolicyStmtDB))
		if ok := db.PolicyStmtDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicyStmt); ok != true {
			db.Logger.Println(" return value not ok")
			err = errors.New("error inserting into policy stmt DB")
			return err
		}
		db.LocalPolicyStmtDB.updateLocalDB(patriciaDB.Prefix(cfg.Name), add)
	} else {
		db.Logger.Println("Duplicate Policy definition name")
		err = errors.New("Duplicate policy definition")
		return err
	}
	return err
}

func (db *PolicyEngineDB) DeletePolicyStatement(cfg PolicyStmtConfig) (err error) {
	db.Logger.Println("DeletePolicyStatement for name ", cfg.Name)
	ok := db.PolicyStmtDB.Match(patriciaDB.Prefix(cfg.Name))
	if !ok {
		err = errors.New("No policy statement with this name found")
		return err
	}
	policyStmtInfoGet := db.PolicyStmtDB.Get(patriciaDB.Prefix(cfg.Name))
	if policyStmtInfoGet != nil {
		policyStmtInfo := policyStmtInfoGet.(PolicyStmt)
		if len(policyStmtInfo.PolicyList) != 0 {
			db.Logger.Println("This policy stmt is being used by one or more policies. Delete the policies before deleting the stmt")
			err = errors.New("This policy stmt is being used by one or more policies. Delete the policies before deleting the stmt")
			return err
		}
		//invalidate localPolicyStmt
		/*	   if policyStmtInfo.LocalDBSliceIdx < int8(len(*db.LocalPolicyStmtDB)) {
		          db.Logger.Println("local DB slice index for this policy stmt is ", policyStmtInfo.LocalDBSliceIdx)
				  LocalPolicyStmtDB := LocalDBSlice (*db.LocalPolicyStmtDB)
				  LocalPolicyStmtDB[policyStmtInfo.LocalDBSliceIdx].IsValid = false
			   }*/
		// PolicyEngineTraverseAndReverse(policyStmtInfo)
		db.Logger.Println("Deleting policy statement with name ", cfg.Name)
		if ok := db.PolicyStmtDB.Delete(patriciaDB.Prefix(cfg.Name)); ok != true {
			db.Logger.Println(" return value not ok for delete PolicyStmtDB")
			err = errors.New("error with delteing policy stmt")
			return err
		}
		db.LocalPolicyStmtDB.updateLocalDB(patriciaDB.Prefix(cfg.Name), del)
		//update other tables
		if len(policyStmtInfo.Conditions) > 0 {
			for i := 0; i < len(policyStmtInfo.Conditions); i++ {
				db.UpdateConditions(policyStmtInfo, policyStmtInfo.Conditions[i], del)
			}
		}
		if len(policyStmtInfo.Actions) > 0 {
			var action PolicyAction
			for i := 0; i < len(policyStmtInfo.Actions); i++ {
				actionItem := db.PolicyActionsDB.Get(patriciaDB.Prefix(policyStmtInfo.Actions[i]))
				if actionItem != nil {
					action = actionItem.(PolicyAction)
				} else {
					db.Logger.Println("action name ", policyStmtInfo.Actions[i], " not defined")
					err = errors.New("action name not defined")
				}
				db.UpdateActions(policyStmtInfo, action, del)
			}
		}
	}
	return err
}

func (db *PolicyEngineDB) CreatePolicyDefinition(cfg PolicyDefinitionConfig) (err error) {
	db.Logger.Println("CreatePolicyDefinition")
	policy := db.PolicyDB.Get(patriciaDB.Prefix(cfg.Name))
	var i int
	if policy == nil {
		db.Logger.Println("Defining a new policy with name ", cfg.Name)
		var newPolicy Policy
		newPolicy.Name = cfg.Name
		newPolicy.Precedence = cfg.Precedence
		newPolicy.MatchType = cfg.MatchType
		db.Logger.Println("Policy has %d ", len(cfg.PolicyDefinitionStatements), " number of statements")
		newPolicy.PolicyStmtPrecedenceMap = make(map[int]string)
		for i = 0; i < len(cfg.PolicyDefinitionStatements); i++ {
			var stmt PolicyStmt
			db.Logger.Println("Adding statement ", cfg.PolicyDefinitionStatements[i].Statement, " at precedence id ", cfg.PolicyDefinitionStatements[i].Precedence)
			newPolicy.PolicyStmtPrecedenceMap[int(cfg.PolicyDefinitionStatements[i].Precedence)] = cfg.PolicyDefinitionStatements[i].Statement
			Item := db.PolicyStmtDB.Get(patriciaDB.Prefix(cfg.PolicyDefinitionStatements[i].Statement))
			if Item != nil {
				stmt = Item.(PolicyStmt)
			} else {
				db.Logger.Println("action name ", cfg.PolicyDefinitionStatements[i].Statement, " not defined")
				err = errors.New("action name not defined")
			}
			err = db.SetAndValidatePolicyType(&newPolicy, stmt)
			if err != nil {
				db.Logger.Println("Error in setting policy type")
				err = errors.New("Error setting policy type")
				return err
			}
			err = db.UpdateGlobalStatementTable(newPolicy.Name, cfg.PolicyDefinitionStatements[i].Statement, add)
			if err != nil {
				db.Logger.Println("UpdateGlobalStatementTable returned err ", err)
				err = errors.New("Error with UpdateGlobalStatementTable")
				return err
			}
			err = db.UpdateStatements(newPolicy, stmt, add)
			if err != nil {
				db.Logger.Println("updateStatements returned err ", err)
				err = errors.New("error with updateStatements")
				return err
			}
		}
		for k := range newPolicy.PolicyStmtPrecedenceMap {
			db.Logger.Println("key k = ", k)
		}
		newPolicy.LocalDBSliceIdx = int8(len(*db.LocalPolicyDB))
		newPolicy.Extensions = cfg.Extensions
		if ok := db.PolicyDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicy); ok != true {
			db.Logger.Println(" return value not ok")
			err = errors.New("error inserting into policyDB")
			return err
		}
		db.LocalPolicyDB.updateLocalDB(patriciaDB.Prefix(cfg.Name), add)
		if newPolicy.ImportPolicy {
			db.Logger.Println("Adding ", newPolicy.Name, " as import policy")
			if db.ImportPolicyPrecedenceMap == nil {
				db.ImportPolicyPrecedenceMap = make(map[int]string)
			}
			db.ImportPolicyPrecedenceMap[int(cfg.Precedence)] = cfg.Name
		} else if newPolicy.ExportPolicy {
			db.Logger.Println("Adding ", newPolicy.Name, " as export policy")
			if db.ExportPolicyPrecedenceMap == nil {
				db.ExportPolicyPrecedenceMap = make(map[int]string)
			}
			db.ExportPolicyPrecedenceMap[int(cfg.Precedence)] = cfg.Name
		}
		db.PolicyEngineTraverseAndApplyPolicy(newPolicy)
	} else {
		db.Logger.Println("Duplicate Policy definition name")
		err = errors.New("Duplicate policy definition")
		return err
	}
	return err
}

func (db *PolicyEngineDB) DeletePolicyDefinition(cfg PolicyDefinitionConfig) (err error) {
	db.Logger.Println("DeletePolicyDefinition for name ", cfg.Name)
	ok := db.PolicyDB.Match(patriciaDB.Prefix(cfg.Name))
	if !ok {
		err = errors.New("No policy with this name found")
		return err
	}
	policyInfoGet := db.PolicyDB.Get(patriciaDB.Prefix(cfg.Name))
	if policyInfoGet != nil {
		policyInfo := policyInfoGet.(Policy)
		//invalidate localPolicy
		/*  if policyInfo.LocalDBSliceIdx < int8(len(*db.LocalPolicyDB)) {
		          db.Logger.Println("local DB slice index for this policy is ", policyInfo.LocalDBSliceIdx)
				  LocalPolicyDB := LocalDBSlice (*db.LocalPolicyDB)
				  LocalPolicyDB[policyInfo.LocalDBSliceIdx].IsValid = false
			   }*/
		db.PolicyEngineTraverseAndReversePolicy(policyInfo)
		db.Logger.Println("Deleting policy with name ", cfg.Name)
		if ok := db.PolicyDB.Delete(patriciaDB.Prefix(cfg.Name)); ok != true {
			db.Logger.Println(" return value not ok for delete PolicyDB")
			err = errors.New("error deleting from policyDB")
			return err
		}
		db.LocalPolicyDB.updateLocalDB(patriciaDB.Prefix(cfg.Name), del)
		var stmt PolicyStmt
		for _, v := range policyInfo.PolicyStmtPrecedenceMap {
			err = db.UpdateGlobalStatementTable(policyInfo.Name, v, del)
			if err != nil {
				db.Logger.Println("UpdateGlobalStatementTable returned err ", err)
				err = errors.New("UpdateGlobalStatementTable returned err")
				return err
			}
			Item := db.PolicyStmtDB.Get(patriciaDB.Prefix(v))
			if Item != nil {
				stmt = Item.(PolicyStmt)
			} else {
				db.Logger.Println("action name ", v, " not defined")
				err = errors.New("action name not defined")
			}
			err = db.UpdateStatements(policyInfo, stmt, del)
			if err != nil {
				db.Logger.Println("updateStatements returned err ", err)
				err = errors.New("UpdateStatements returned err")
				return err
			}
		}
		if policyInfo.ExportPolicy {
			if db.ExportPolicyPrecedenceMap != nil {
				delete(db.ExportPolicyPrecedenceMap, int(policyInfo.Precedence))
			}
		}
		if policyInfo.ImportPolicy {
			if db.ImportPolicyPrecedenceMap != nil {
				delete(db.ImportPolicyPrecedenceMap, int(policyInfo.Precedence))
			}
		}
	}
	return err
}
