// policyUtils.go
package policy

import (
	"bytes"
	"errors"
	"log"
	"log/syslog"
	"os"
	"utils/patriciaDB"
	"utils/policy/policyCommonDefs"
)

const (
	add = iota
	del
	delAll
	invalidate
)
const (
	Invalid = -1
	Valid   = 0
)

type ConditionsAndActionsList struct {
	ConditionList []PolicyCondition
	ActionList    []PolicyAction
}

type PolicyStmtMap struct {
	PolicyStmtMap map[string]ConditionsAndActionsList
}

type PolicyEngineFilterEntityParams struct {
	DestNetIp        string //CIDR format
	NextHopIp        string
	RouteProtocol    string
	CreatePath       bool
	DeletePath       bool
	PolicyList       []string
	PolicyHitCounter int
}

//struct sent to the application for updating its local maps/DBs
type PolicyDetails struct {
	Policy        string
	PolicyStmt    string
	ConditionList []PolicyCondition
	ActionList    []PolicyAction
	EntityDeleted bool //whether this policy/stmt resulted in deleting the entity
}

type LocalDB struct {
	Prefix     patriciaDB.Prefix
	IsValid    bool
	Precedence int
}
type LocalDBSlice []LocalDB

func (slice *LocalDBSlice) updateLocalDB(prefix patriciaDB.Prefix, op int) {
	if slice == nil {
		return
	}
	tempSlice := *slice
	if op == add {
		localDBRecord := LocalDB{Prefix: prefix, IsValid: true}
		tempSlice = append(tempSlice, localDBRecord)
	} else if op == del {
		found := false
		var i int
		for i = 0; i < len(tempSlice); i++ {
			if bytes.Equal(tempSlice[i].Prefix, prefix) {
				found = true
				break
			}
		}
		if found == true {
			if len(tempSlice) <= i+1 {
				tempSlice = tempSlice[:i]
			} else {
				tempSlice = append(tempSlice[:i], tempSlice[i+1:]...)
			}
		}
	}
	*slice = tempSlice
}

type Policyfunc func(actionInfo interface{}, conditionInfo []interface{}, params interface{})
type PolicyConditionCheckfunc func(entity PolicyEngineFilterEntityParams, condition PolicyCondition, policyStmt PolicyStmt) bool
type UndoActionfunc func(actionInfo interface{}, conditionList []interface{}, params interface{}, policyStmt PolicyStmt)
type PolicyCheckfunc func(params interface{}) bool
type EntityUpdatefunc func(details PolicyDetails, params interface{})
type PolicyApplyfunc func(entity PolicyEngineFilterEntityParams, policyData interface{}, params interface{})
type EntityTraverseAndApplyPolicyfunc func(data interface{}, updatefunc PolicyApplyfunc)
type EntityTraverseAndReversePolicyfunc func(data interface{})
type PolicyEntityMapIndex interface{}
type GetPolicyEnityMapIndexFunc func(entity PolicyEngineFilterEntityParams, policy string) PolicyEntityMapIndex

type PolicyEngineDB struct {
	Logger                        *log.Logger
	PolicyConditionsDB            *patriciaDB.Trie
	LocalPolicyConditionsDB       *LocalDBSlice
	PolicyActionsDB               *patriciaDB.Trie
	LocalPolicyActionsDB          *LocalDBSlice
	PolicyStmtDB                  *patriciaDB.Trie
	LocalPolicyStmtDB             *LocalDBSlice
	PolicyDB                      *patriciaDB.Trie
	LocalPolicyDB                 *LocalDBSlice
	PolicyStmtPolicyMapDB         map[string][]string //policies using this statement
	PrefixPolicyListDB            *patriciaDB.Trie
	ProtocolPolicyListDB          map[string][]string //policystmt names assoociated with every protocol type
	ImportPolicyPrecedenceMap     map[int]string
	ExportPolicyPrecedenceMap     map[int]string
	PolicyEntityMap               map[PolicyEntityMapIndex]PolicyStmtMap
	DefaultImportPolicyActionFunc Policyfunc
	DefaultExportPolicyActionFunc Policyfunc
	IsEntityPresentFunc           PolicyCheckfunc
	GetPolicyEntityMapIndex       GetPolicyEnityMapIndexFunc
	UpdateEntityDB                EntityUpdatefunc
	ConditionCheckfuncMap         map[int]PolicyConditionCheckfunc
	ActionfuncMap                 map[int]Policyfunc
	UndoActionfuncMap             map[int]UndoActionfunc
	TraverseAndApplyPolicyFunc    EntityTraverseAndApplyPolicyfunc
	TraverseAndReversePolicyFunc  EntityTraverseAndReversePolicyfunc
}

func (db *PolicyEngineDB) buildPolicyConditionCheckfuncMap() {
	db.Logger.Println("buildPolicyConditionCheckfuncMap")
	db.ConditionCheckfuncMap[policyCommonDefs.PolicyConditionTypeDstIpPrefixMatch] = db.DstIpPrefixMatchConditionfunc
	db.ConditionCheckfuncMap[policyCommonDefs.PolicyConditionTypeProtocolMatch] = db.ProtocolMatchConditionfunc
}
func NewPolicyEngineDB() (policyEngineDB *PolicyEngineDB) {
	policyEngineDB = &PolicyEngineDB{}
	if policyEngineDB.Logger == nil {
		policyEngineDB.Logger = log.New(os.Stdout, "PolicyEngine :", log.Ldate|log.Ltime|log.Lshortfile)

		syslogger, err := syslog.New(syslog.LOG_NOTICE|syslog.LOG_INFO|syslog.LOG_DAEMON, "PolicyEngine")
		if err == nil {
			syslogger.Info("### PolicyEngineDB initailized")
			policyEngineDB.Logger.SetOutput(syslogger)
		}
	}
	policyEngineDB.PolicyActionsDB = patriciaDB.NewTrie()
	LocalPolicyActionsDB := make([]LocalDB, 0)
	localActionSlice := LocalDBSlice(LocalPolicyActionsDB)
	policyEngineDB.LocalPolicyActionsDB = &localActionSlice

	policyEngineDB.PolicyConditionsDB = patriciaDB.NewTrie()
	LocalPolicyConditionsDB := make([]LocalDB, 0)
	localConditionSlice := LocalDBSlice(LocalPolicyConditionsDB)
	policyEngineDB.LocalPolicyConditionsDB = &localConditionSlice

	policyEngineDB.PolicyStmtDB = patriciaDB.NewTrie()
	LocalPolicyStmtDB := make([]LocalDB, 0)
	localStmtSlice := LocalDBSlice(LocalPolicyStmtDB)
	policyEngineDB.LocalPolicyStmtDB = &localStmtSlice

	policyEngineDB.PolicyDB = patriciaDB.NewTrie()
	LocalPolicyDB := make([]LocalDB, 0)
	localPolicySlice := LocalDBSlice(LocalPolicyDB)
	policyEngineDB.LocalPolicyDB = &localPolicySlice

	policyEngineDB.PolicyStmtPolicyMapDB = make(map[string][]string)
	policyEngineDB.PolicyEntityMap = make(map[PolicyEntityMapIndex]PolicyStmtMap)
	policyEngineDB.PrefixPolicyListDB = patriciaDB.NewTrie()
	policyEngineDB.ProtocolPolicyListDB = make(map[string][]string)
	policyEngineDB.ImportPolicyPrecedenceMap = make(map[int]string)
	policyEngineDB.ExportPolicyPrecedenceMap = make(map[int]string)
	policyEngineDB.ConditionCheckfuncMap = make(map[int]PolicyConditionCheckfunc)
	policyEngineDB.buildPolicyConditionCheckfuncMap()
	policyEngineDB.ActionfuncMap = make(map[int]Policyfunc)
	policyEngineDB.UndoActionfuncMap = make(map[int]UndoActionfunc)
	return policyEngineDB
}

func (db *PolicyEngineDB) SetDefaultImportPolicyActionFunc(defaultfunc Policyfunc) {
	db.DefaultImportPolicyActionFunc = defaultfunc
}
func (db *PolicyEngineDB) SetDefaultExportPolicyActionFunc(defaultfunc Policyfunc) {
	db.DefaultExportPolicyActionFunc = defaultfunc
}
func (db *PolicyEngineDB) SetIsEntityPresentFunc(IsPresent PolicyCheckfunc) {
	db.IsEntityPresentFunc = IsPresent
}
func (db *PolicyEngineDB) SetEntityUpdateFunc(updatefunc EntityUpdatefunc) {
	db.UpdateEntityDB = updatefunc
}
func (db *PolicyEngineDB) SetActionFunc(action int, setfunc Policyfunc) {
	db.ActionfuncMap[action] = setfunc
}
func (db *PolicyEngineDB) SetUndoActionFunc(action int, setfunc UndoActionfunc) {
	db.UndoActionfuncMap[action] = setfunc
}
func (db *PolicyEngineDB) SetTraverseAndApplyPolicyFunc(updatefunc EntityTraverseAndApplyPolicyfunc) {
	db.TraverseAndApplyPolicyFunc = updatefunc
}
func (db *PolicyEngineDB) SetTraverseAndReversePolicyFunc(updatefunc func(policyItem interface{})) {
	db.TraverseAndReversePolicyFunc = updatefunc
}
func (db *PolicyEngineDB) SetGetPolicyEntityMapIndexFunc(getfunc func(entity PolicyEngineFilterEntityParams, policy string) PolicyEntityMapIndex) {
	db.GetPolicyEntityMapIndex = getfunc
}
func isPolicyTypeSame(oldPolicy Policy, policy Policy) (same bool) {
	if oldPolicy.ExportPolicy == policy.ExportPolicy && oldPolicy.ImportPolicy == policy.ImportPolicy {
		same = true
	}
	return same
}
func (db *PolicyEngineDB) AddPolicyEntityMapEntry(entity PolicyEngineFilterEntityParams, policy string,
	policyStmt string, conditionList []PolicyCondition, actionList []PolicyAction) {
	db.Logger.Println("AddPolicyEntityMapEntry")
	var policyStmtMap PolicyStmtMap
	var conditionsAndActionsList ConditionsAndActionsList
	if db.PolicyEntityMap == nil {
		db.PolicyEntityMap = make(map[PolicyEntityMapIndex]PolicyStmtMap)
	}
	if db.GetPolicyEntityMapIndex == nil {
		return
	}
	policyEntityMapIndex := db.GetPolicyEntityMapIndex(entity, policy)
	if policyEntityMapIndex == nil {
		db.Logger.Println("policyEntityMapKey nil")
		return
	}
	policyStmtMap, ok := db.PolicyEntityMap[policyEntityMapIndex]
	if !ok {
		policyStmtMap.PolicyStmtMap = make(map[string]ConditionsAndActionsList)
	}
	_, ok = policyStmtMap.PolicyStmtMap[policyStmt]
	if ok {
		db.Logger.Println("policy statement map for statement ", policyStmt, " already in place for policy ", policy)
		return
	}
	conditionsAndActionsList.ConditionList = make([]PolicyCondition, 0)
	conditionsAndActionsList.ActionList = make([]PolicyAction, 0)
	for i := 0; conditionList != nil && i < len(conditionList); i++ {
		conditionsAndActionsList.ConditionList = append(conditionsAndActionsList.ConditionList, conditionList[i])
	}
	for i := 0; actionList != nil && i < len(actionList); i++ {
		conditionsAndActionsList.ActionList = append(conditionsAndActionsList.ActionList, actionList[i])
	}
	policyStmtMap.PolicyStmtMap[policyStmt] = conditionsAndActionsList
	db.PolicyEntityMap[policyEntityMapIndex] = policyStmtMap
}
func (db *PolicyEngineDB) DeletePolicyEntityMapEntry(entity PolicyEngineFilterEntityParams, policy string) {
	db.Logger.Println("DeletePolicyEntityMapEntry for policy ", policy)
	if db.PolicyEntityMap == nil {
		db.Logger.Println("PolicyEntityMap empty")
		return
	}
	if db.GetPolicyEntityMapIndex == nil {
		return
	}
	policyEntityMapIndex := db.GetPolicyEntityMapIndex(entity, policy)
	if policyEntityMapIndex == nil {
		db.Logger.Println("policyEntityMapIndex nil")
		return
	}
	//PolicyRouteMap[policyRouteIndex].policyStmtMap=nil
	delete(db.PolicyEntityMap, policyEntityMapIndex)
}
func (db *PolicyEngineDB) PolicyActionType(actionType int) (exportTypeAction bool, importTypeAction bool, globalTypeAction bool) {
	db.Logger.Println("PolicyActionType for type ", actionType)
	switch actionType {
	case policyCommonDefs.PoilcyActionTypeSetAdminDistance:
		globalTypeAction = true
		db.Logger.Println("PoilcyActionTypeSetAdminDistance, setting globalTypeAction true")
		break
	case policyCommonDefs.PolicyActionTypeAggregate:
		exportTypeAction = true
		db.Logger.Println("PolicyActionTypeAggregate: setting exportTypeAction true")
		break
	case policyCommonDefs.PolicyActionTypeRouteRedistribute:
		exportTypeAction = true
		db.Logger.Println("PolicyActionTypeRouteRedistribute: setting exportTypeAction true")
		break
	case policyCommonDefs.PolicyActionTypeNetworkStatementAdvertise:
		exportTypeAction = true
		db.Logger.Println("PolicyActionTypeNetworkStatementAdvertise: setting exportTypeAction true")
		break
	case policyCommonDefs.PolicyActionTypeRouteDisposition:
		importTypeAction = true
		db.Logger.Println("setting importTypeAction true")
		break
	default:
		db.Logger.Println("Unknown action type")
		break
	}
	return exportTypeAction, importTypeAction, globalTypeAction
}
func (db *PolicyEngineDB) SetAndValidatePolicyType(policy *Policy, stmt PolicyStmt) (err error) {
	db.Logger.Println("SetPolicyTypeFromPolicyStmt")
	if policy.ExportPolicy == false && policy.ImportPolicy == false && policy.GlobalPolicy == false {
		db.Logger.Println("Policy is still not associated with a type, set it from stmt")
		policy.ExportPolicy = stmt.ExportStmt
		policy.ImportPolicy = stmt.ImportStmt
		policy.GlobalPolicy = stmt.GlobalStmt

		if policy.ImportPolicy && db.ImportPolicyPrecedenceMap != nil {
			_, ok := db.ImportPolicyPrecedenceMap[int(policy.Precedence)]
			if ok {
				db.Logger.Println("There is already a import policy with this precedence.")
				err = errors.New("There is already a import policy with this precedence.")
				return err
			}
		} else if policy.ExportPolicy && db.ExportPolicyPrecedenceMap != nil {
			_, ok := db.ExportPolicyPrecedenceMap[int(policy.Precedence)]
			if ok {
				db.Logger.Println("There is already a export policy with this precedence.")
				err = errors.New("There is already a export policy with this precedence.")
				return err
			}
		} else if policy.GlobalPolicy {
			db.Logger.Println("This is a global policy")
		}
		return err
	}
	if policy.ExportPolicy != stmt.ExportStmt ||
		policy.ImportPolicy != stmt.ImportStmt ||
		policy.GlobalPolicy != stmt.GlobalStmt {
		db.Logger.Println("Policy type settings, export/import/global :", policy.ExportPolicy, "/", policy.ImportPolicy, "/", policy.GlobalPolicy, " does not match the export/import/global settings on the stmt: ", stmt.ExportStmt, "/", stmt.ImportStmt, "/", stmt.GlobalStmt)
		err = errors.New("Mismatch on policy type")
		return err
	}
	return err
}
