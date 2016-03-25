// PolicyConditionApis.go
package policy

import (
	"errors"
	"utils/policy/policyCommonDefs"
	"utils/patriciaDB"
)
type PolicyPrefix struct{
	IpPrefix string //CIDR eg: 1.1.1.2/24
	MasklengthRange string //exact or a specific range 21..24
}
type PolicyDstIpMatchPrefixSetCondition struct{
	PrefixSet string
	Prefix PolicyPrefix
}

type MatchPrefixConditionInfo struct {
	UsePrefixSet bool
	PrefixSet string
	DstIpMatch     bool
	SrcIpMatch     bool
	Prefix PolicyPrefix
}
type PolicyConditionConfig struct {
	Name string
	ConditionType string
	MatchProtocolConditionInfo string          
    MatchDstIpPrefixConditionInfo PolicyDstIpMatchPrefixSetCondition       
    //MatchNeighborConditionInfo   PolicyMatchNeighborSetCondition        
	//MatchTagConditionInfo   PolicyMatchTagSetCondition             
}

type PolicyCondition struct {
	Name          string
	ConditionType int
	ConditionInfo interface {}
	PolicyStmtList    [] string
	ConditionGetBulkInfo string
	LocalDBSliceIdx int
}

func (db * PolicyEngineDB) CreatePolicyDstIpMatchPrefixSetCondition(inCfg PolicyConditionConfig) (val bool, err error) {
	db.Logger.Println("CreatePolicyDstIpMatchPrefixSetCondition")
	cfg := inCfg.MatchDstIpPrefixConditionInfo
	var conditionInfo MatchPrefixConditionInfo
	var conditionGetBulkInfo string
    if len(cfg.PrefixSet) == 0 && len(cfg.Prefix.IpPrefix) == 0 {
		db.Logger.Println("Empty prefix set/nil prefix")
		err = errors.New("Empty prefix set/nil prefix")
		return val, err
	}
    if len(cfg.PrefixSet) != 0 && len(cfg.Prefix.IpPrefix) != 0 {
		db.Logger.Println("Cannot provide both prefix set and individual prefix")
		err = errors.New("Cannot provide both prefix set and individual prefix")
		return val, err
	}
    if len(cfg.Prefix.IpPrefix) != 0 {
	   conditionInfo.UsePrefixSet = false
       conditionInfo.Prefix.IpPrefix = cfg.Prefix.IpPrefix
	   conditionInfo.Prefix.MasklengthRange = cfg.Prefix.MasklengthRange
	   conditionGetBulkInfo = "match destination Prefix " + cfg.Prefix.IpPrefix + "MasklengthRange " + cfg.Prefix.MasklengthRange
	} else if len(cfg.PrefixSet) != 0 {
		conditionInfo.UsePrefixSet = true
		conditionInfo.PrefixSet = cfg.PrefixSet
	    conditionGetBulkInfo = "match destination Prefix " + cfg.PrefixSet
	}
	conditionInfo.DstIpMatch = true
	policyCondition := db.PolicyConditionsDB.Get(patriciaDB.Prefix(inCfg.Name))
	if(policyCondition == nil) {
	   db.Logger.Println("Defining a new policy condition with name ", inCfg.Name)
	   newPolicyCondition := PolicyCondition{Name:inCfg.Name,ConditionType:policyCommonDefs.PolicyConditionTypeDstIpPrefixMatch,ConditionInfo:conditionInfo ,LocalDBSliceIdx:(len(*db.LocalPolicyConditionsDB))}
       newPolicyCondition.ConditionGetBulkInfo = conditionGetBulkInfo 
	   if ok := db.PolicyConditionsDB.Insert(patriciaDB.Prefix(inCfg.Name), newPolicyCondition); ok != true {
	   db.Logger.Println(" return value not ok")
	   err = errors.New("Error creating condition in the DB")
	   return val, err
	}
	db.LocalPolicyConditionsDB.updateLocalDB(patriciaDB.Prefix(inCfg.Name),add)
    } else {
		db.Logger.Println("Duplicate Condition name")
		err = errors.New("Duplicate policy condition definition")
		return val, err
	}	
	return val, err
}

func (db * PolicyEngineDB)CreatePolicyMatchProtocolCondition(cfg PolicyConditionConfig) (val bool, err error) {
	db.Logger.Println("CreatePolicyMatchProtocolCondition")

	policyCondition := db.PolicyConditionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if(policyCondition == nil) {
	   db.Logger.Println("Defining a new policy condition with name ", cfg.Name, " to match on protocol ", cfg.MatchProtocolConditionInfo)
	   matchProto := cfg.MatchProtocolConditionInfo
	   newPolicyCondition := PolicyCondition{Name:cfg.Name,ConditionType:policyCommonDefs.PolicyConditionTypeProtocolMatch,ConditionInfo:matchProto ,LocalDBSliceIdx:(len(*db.LocalPolicyConditionsDB))}
       newPolicyCondition.ConditionGetBulkInfo = "match Protocol " + matchProto
		if ok := db.PolicyConditionsDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicyCondition); ok != true {
			db.Logger.Println(" return value not ok")
	        err = errors.New("Error creating condition in the DB")
			return val, err
		}
	    db.LocalPolicyConditionsDB.updateLocalDB(patriciaDB.Prefix(cfg.Name),add)
	} else {
		db.Logger.Println("Duplicate Condition name")
		err = errors.New("Duplicate policy condition definition")
		return val, err
	}
	return val, err
}
func (db * PolicyEngineDB)CreatePolicyCondition(cfg PolicyConditionConfig) (err error) {
	db.Logger.Println("CreatePolicyCondition")
	switch cfg.ConditionType {
		case "MatchDstIpPrefix":
		   _, err = db.CreatePolicyDstIpMatchPrefixSetCondition(cfg)
		   break
		case "MatchProtocol":
		    _, err = db.CreatePolicyMatchProtocolCondition(cfg)
		   break
		default:
		   db.Logger.Println("Unknown condition type ", cfg.ConditionType)
		   err = errors.New("Unknown condition type")
	}
	return err
}
func (db * PolicyEngineDB) DeletePolicyCondition(cfg PolicyConditionConfig) (err error) {
	db.Logger.Println("DeletePolicyCondition")
	conditionItem := db.PolicyConditionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if conditionItem == nil {
		db.Logger.Println("Condition ", cfg.Name, "not found in the DB")
		err = errors.New("Condition not found")
		return err
	}
	condition := conditionItem.(PolicyCondition)
	if len(condition.PolicyStmtList) != 0 {
		db.Logger.Println("This condition is currently being used by a policy statement. Try deleting the stmt before deleting the condition")
		err = errors.New("This condition is currently being used by a policy statement. Try deleting the stmt before deleting the condition")
		return err
	}
	deleted := db.PolicyConditionsDB.Delete(patriciaDB.Prefix(cfg.Name))
	if deleted {
		db.Logger.Println("Found and deleted condition ", cfg.Name)
		db.LocalPolicyConditionsDB.updateLocalDB(patriciaDB.Prefix(cfg.Name),del)
	}
	return err
}
