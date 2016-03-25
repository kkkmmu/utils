// ribdPolicyActionApis.go
package policy

import (
	"errors"
	"strconv"
	"utils/patriciaDB"
	"utils/policy/policyCommonDefs"
)

type RedistributeActionInfo struct {
	Redistribute               bool
	RedistributeTargetProtocol string
}

type PolicyAggregateActionInfo struct {
	GenerateASSet   bool
	SendSummaryOnly bool
}

type PolicyAction struct {
	Name              string
	ActionType        int
	ActionInfo        interface{}
	PolicyStmtList    []string
	ActionGetBulkInfo string
	LocalDBSliceIdx   int
}

type PolicyActionConfig struct {
	Name                           string
	ActionType                     string
	SetAdminDistanceValue          int
	Accept                         bool
	Reject                         bool
	RedistributeAction             string
	RedistributeTargetProtocol     string
	NetworkStatementTargetProtocol string
	GenerateASSet                  bool
	SendSummaryOnly                bool
}

func (db *PolicyEngineDB) CreatePolicyRouteDispositionAction(cfg PolicyActionConfig) (val bool, err error) {
	db.Logger.Println("CreateRouteDispositionAction")
	policyAction := db.PolicyActionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if policyAction == nil {
		db.Logger.Println("Defining a new policy action with name ", cfg.Name)
		routeDispositionAction := ""
		if cfg.Accept == true {
			routeDispositionAction = "Accept"
		} else if cfg.Reject == true {
			routeDispositionAction = "Reject"
		} else {
			db.Logger.Println("User should set either one of accept/reject to true for this action type")
			err = errors.New("User should set either one of accept/reject to true for this action type")
			return val, err
		}
		newPolicyAction := PolicyAction{Name: cfg.Name, ActionType: policyCommonDefs.PolicyActionTypeRouteDisposition, ActionInfo: routeDispositionAction, LocalDBSliceIdx: (len(*db.LocalPolicyActionsDB))}
		newPolicyAction.ActionGetBulkInfo = routeDispositionAction
		if ok := db.PolicyActionsDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicyAction); ok != true {
			db.Logger.Println(" return value not ok")
			err = errors.New("Error inserting action in DB")
			return val, err
		}
		db.LocalPolicyActionsDB.updateLocalDB(patriciaDB.Prefix(cfg.Name), add)
	} else {
		db.Logger.Println("Duplicate action name")
		err = errors.New("Duplicate policy action definition")
		return val, err
	}
	return val, err
}

func (db *PolicyEngineDB) CreatePolicyAdminDistanceAction(cfg PolicyActionConfig) (val bool, err error) {
	db.Logger.Println("CreatePolicyAdminDistanceAction")
	policyAction := db.PolicyActionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if policyAction == nil {
		db.Logger.Println("Defining a new policy action with name ", cfg.Name, "Setting admin distance value to ", cfg.SetAdminDistanceValue)
		newPolicyAction := PolicyAction{Name: cfg.Name, ActionType: policyCommonDefs.PoilcyActionTypeSetAdminDistance, ActionInfo: cfg.SetAdminDistanceValue, LocalDBSliceIdx: (len(*db.LocalPolicyActionsDB))}
		newPolicyAction.ActionGetBulkInfo = "Set admin distance to value " + strconv.Itoa(int(cfg.SetAdminDistanceValue))
		if ok := db.PolicyActionsDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicyAction); ok != true {
			db.Logger.Println(" return value not ok")
			err = errors.New("Error inserting action in DB")
			return val, err
		}
		db.LocalPolicyActionsDB.updateLocalDB(patriciaDB.Prefix(cfg.Name), add)
	} else {
		db.Logger.Println("Duplicate action name")
		err = errors.New("Duplicate policy action definition")
		return val, err
	}
	return val, err
}
func (db *PolicyEngineDB) CreatePolicyNetworkStatementAdvertiseAction(cfg PolicyActionConfig) (val bool, err error) {
	db.Logger.Println("CreatePolicyNetworkStatementAdvertiseAction")
	policyAction := db.PolicyActionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if policyAction == nil {
		db.Logger.Println("Defining a new policy action with name ", cfg.Name)
		newPolicyAction := PolicyAction{Name: cfg.Name, ActionType: policyCommonDefs.PolicyActionTypeNetworkStatementAdvertise, ActionInfo: cfg.NetworkStatementTargetProtocol, LocalDBSliceIdx: (len(*db.LocalPolicyActionsDB))}
		newPolicyAction.ActionGetBulkInfo = "Advertise network statement to " + cfg.NetworkStatementTargetProtocol
		if ok := db.PolicyActionsDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicyAction); ok != true {
			db.Logger.Println(" return value not ok")
			err = errors.New("Error inserting action in DB")
			return val, err
		}
		db.LocalPolicyActionsDB.updateLocalDB(patriciaDB.Prefix(cfg.Name), add)
	} else {
		db.Logger.Println("Duplicate action name")
		err = errors.New("Duplicate policy action definition")
		return val, err
	}
	return val, err
}
func (db *PolicyEngineDB) CreatePolicyRedistributionAction(cfg PolicyActionConfig) (val bool, err error) {
	db.Logger.Println("CreatePolicyRedistributionAction")

	policyAction := db.PolicyActionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if policyAction == nil {
		db.Logger.Println("Defining a new policy action with name ", cfg.Name)
		redistributeActionInfo := RedistributeActionInfo{RedistributeTargetProtocol: cfg.RedistributeTargetProtocol}
		if cfg.RedistributeAction == "Allow" {
			redistributeActionInfo.Redistribute = true
		} else if cfg.RedistributeAction == "Block" {
			redistributeActionInfo.Redistribute = false
		} else {
			db.Logger.Println("Invalid redistribute option ", cfg.RedistributeAction, " - should be either Allow/Block")
			err = errors.New("Invalid redistribute option")
			return val, err
		}
		newPolicyAction := PolicyAction{Name: cfg.Name, ActionType: policyCommonDefs.PolicyActionTypeRouteRedistribute, ActionInfo: redistributeActionInfo, LocalDBSliceIdx: (len(*db.LocalPolicyActionsDB))}
		newPolicyAction.ActionGetBulkInfo = cfg.RedistributeAction + " Redistribute to Target Protocol " + cfg.RedistributeTargetProtocol
		if ok := db.PolicyActionsDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicyAction); ok != true {
			db.Logger.Println(" return value not ok")
			err = errors.New("Error inserting action in DB")
			return val, err
		}
		db.LocalPolicyActionsDB.updateLocalDB(patriciaDB.Prefix(cfg.Name), add)
	} else {
		db.Logger.Println("Duplicate action name")
		err = errors.New("Duplicate policy action definition")
		return val, err
	}
	return val, err
}

func (db *PolicyEngineDB) CreatePolicyAggregateAction(cfg PolicyActionConfig) (val bool, err error) {
	db.Logger.Println("CreatePolicyAggregateAction")

	policyAction := db.PolicyActionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if policyAction == nil {
		db.Logger.Println("Defining a new policy action with name ", cfg.Name)
		aggregateActionInfo := PolicyAggregateActionInfo{GenerateASSet: cfg.GenerateASSet, SendSummaryOnly: cfg.SendSummaryOnly}
		newPolicyAction := PolicyAction{Name: cfg.Name, ActionType: policyCommonDefs.PolicyActionTypeAggregate, ActionInfo: aggregateActionInfo, LocalDBSliceIdx: (len(*db.LocalPolicyActionsDB))}
		newPolicyAction.ActionGetBulkInfo = "Aggregate action set GenerateASSet to " +
			strconv.FormatBool(cfg.GenerateASSet) + " set SendSummaryOnly to " + strconv.FormatBool(cfg.SendSummaryOnly)
		if ok := db.PolicyActionsDB.Insert(patriciaDB.Prefix(cfg.Name), newPolicyAction); ok != true {
			db.Logger.Println(" return value not ok")
			err = errors.New("Error inserting action in DB")
			return val, err
		}
		db.LocalPolicyActionsDB.updateLocalDB(patriciaDB.Prefix(cfg.Name), add)
	} else {
		db.Logger.Println("Duplicate action name")
		err = errors.New("Duplicate policy action definition")
		return val, err
	}
	return val, err
}

func (db *PolicyEngineDB) CreatePolicyAction(cfg PolicyActionConfig) (err error) {
	db.Logger.Println("CreatePolicyAction")
	switch cfg.ActionType {
	case "RouteDisposition":
		_,err = db.CreatePolicyRouteDispositionAction(cfg)
		break
	case "Redistribution":
		_,err = db.CreatePolicyRedistributionAction(cfg)
		break
	case "SetAdminDistance":
		_,err = db.CreatePolicyAdminDistanceAction(cfg)
		break
	case "NetworkStatementAdvertise":
		_,err = db.CreatePolicyNetworkStatementAdvertiseAction(cfg)
		break
	case "Aggregate":
		_,err = db.CreatePolicyAggregateAction(cfg)
		break
	default:
		db.Logger.Println("Unknown action type ", cfg.ActionType)
		err = errors.New("Unknown action type")
	}
	return err
}

func (db *PolicyEngineDB) DeletePolicyAction(cfg PolicyActionConfig) (err error) {
	db.Logger.Println("DeletePolicyAction")
	actionItem := db.PolicyActionsDB.Get(patriciaDB.Prefix(cfg.Name))
	if actionItem == nil {
		db.Logger.Println("action ", cfg.Name, "not found in the DB")
		err = errors.New("action not found")
		return err
	}
	action := actionItem.(PolicyAction)
	if len(action.PolicyStmtList) != 0 {
		db.Logger.Println("This action is currently being used by one or more policy statements. Try deleting the stmt before deleting the action")
		err = errors.New("This action is currently being used by one or more policy statements. Try deleting the stmt before deleting the action")
		return err
	}
	deleted := db.PolicyActionsDB.Delete(patriciaDB.Prefix(cfg.Name))
	if deleted {
		db.Logger.Println("Found and deleted actions ", cfg.Name)
		db.LocalPolicyActionsDB.updateLocalDB(patriciaDB.Prefix(cfg.Name), del)
	}
	return err
}
