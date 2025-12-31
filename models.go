package sdk

import (
	"encoding/json"
	"fmt"
)

// This file contains all type definitions copied from catalog_service dependency.
// All types are organized by their original packages for clarity.

// ============ Infra: Filter types ============

type CommonCondition struct {
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Order    string         `json:"order"`
	OrderBy  string         `json:"order_by"`
	Filters  []CommonFilter `json:"filters"`
}

type CommonFilter struct {
	Name         string        `json:"name"`
	Values       []string      `json:"values"`
	Fuzzy        bool          `json:"fuzzy"`
	FilterValues []interface{} `json:"-"`
}

// ============ Models: Common types and IDs ============

type DatabaseID int64
type TableID int64
type CatalogID int64
type VolumeID string
type FileID string
type UserID uint
type RoleID uint
type PrivID uint
type PrivCode string
type PrivObjectID string
type ObjType uint16
type PrivType uint

const DatabaseIDNotFound = int64(9223372036854775807) // math.MaxInt64
const CatalogIDNotFound = int64(9223372036854775807)  // math.MaxInt64
const RoleIDNotFound = uint(4294967295)               // math.MaxUint
const UserIDNotFound UserID = 4294967295              // math.MaxUint

var TableIDInSubDatabase TableID = -1 //订阅库下的表都没有表id,用一个特殊值

type FullPath struct {
	IDList   []string `json:"id_list"`
	NameList []string `json:"name_list"`
}

// ============ Models: File types ============

// FileType represents the type of a file in the system.
type FileType int

const (
	FileTypeUnknown  FileType = 0
	FileTypeTXT      FileType = 1
	FileTypePDF      FileType = 2
	FileTypePPT      FileType = 4
	FileTypeDOC      FileType = 5
	FileTypeMarkdown FileType = 6
	FileTypeCSV      FileType = 7
	FileTypeParquet  FileType = 8
	FileTypeSQLFiles FileType = 9
	FileTypeDir      FileType = 10
	FileTypeDOCX     FileType = 11
	FileTypePPTX     FileType = 12
	FileTypeWAV      FileType = 13
	FileTypeMP3      FileType = 14
	FileTypeAAC      FileType = 15
	FileTypeFLAC     FileType = 16
	FileTypeMP4      FileType = 17
	FileTypeMOV      FileType = 18
	FileTypeMKV      FileType = 19
	FileTypePNG      FileType = 20
	FileTypeJPG      FileType = 21
	FileTypeJPEG     FileType = 22
	FileTypeBMP      FileType = 23
	FileTypeXLS      FileType = 24
	FileTypeXLSX     FileType = 25
	FileTypeHTM      FileType = 27
	FileTypeHTML     FileType = 28
	FileTypeEML      FileType = 29
	FileTypeMSG      FileType = 30
	FileTypeP7S      FileType = 31
	FileTypeDWG      FileType = 32
	FileTypeDXF      FileType = 33
	FileTypeFAS      FileType = 34
	FileTypeIMAGE    FileType = 3
)

// ============ Models: Priv types ============

const (
	ObjTypeNone ObjType = iota
	ObjTypeConnector
	ObjTypeLoadTask
	ObjTypeWorkFlow
	ObjTypeVolume
	ObjTypeDataSet
	ObjTypeAlarm
	ObjTypeUser
	ObjTypeRole
	ObjTypeExportTask
	ObjTypeDataCenter
	ObjTypeCatalog
	ObjTypeDatabase
	ObjTypeTable
	ObjTypeKnowledge    //知识库
	ObjTypePublication  //发布操作
	ObjTypeSubscription //订阅操作
)

func (objType ObjType) String() string {
	switch objType {
	case ObjTypeConnector:
		return "connector"
	case ObjTypeLoadTask:
		return "load_task"
	case ObjTypeWorkFlow:
		return "workflow"
	case ObjTypeVolume:
		return "volume"
	case ObjTypeDataSet:
		return "dataset"
	case ObjTypeAlarm:
		return "alarm"
	case ObjTypeUser:
		return "user"
	case ObjTypeRole:
		return "role"
	case ObjTypeExportTask:
		return "export_task"
	case ObjTypeDataCenter:
		return "data_center"
	case ObjTypeCatalog:
		return "catalog"
	case ObjTypeDatabase:
		return "database"
	case ObjTypeTable:
		return "table"
	case ObjTypeKnowledge:
		return "knowledge"
	case ObjTypePublication:
		return "publication"
	case ObjTypeSubscription:
		return "subscription"
	default:
		return "none"
	}
}

func (pc PrivCode) String() string {
	return string(pc)
}

// ============ PrivID constants ============

const (
	//用户
	PrivID_CreateUser       PrivID = 1
	PrivID_QueryUser        PrivID = 2
	PrivID_UpdatePassword   PrivID = 3
	PrivID_UpdateUserRole   PrivID = 4
	PrivID_UpdateUserInfo   PrivID = 5
	PrivID_UpdateUserStatus PrivID = 6
	PrivID_DeleteUser       PrivID = 7
	PrivID_QueryUserLog     PrivID = 8

	//角色
	PrivID_CreateRole       PrivID = 9
	PrivID_QueryRole        PrivID = 10
	PrivID_UpdateRoleInfo   PrivID = 11
	PrivID_UpdateRoleStatus PrivID = 12
	PrivID_DeleteRole       PrivID = 13
	PrivID_QueryRoleLog     PrivID = 14

	//连接器
	PrivID_CreateConnector PrivID = 15
	PrivID_QueryConnector  PrivID = 16 //查询连接器的列表
	PrivID_UpdateConnector PrivID = 17
	PrivID_DeleteConnector PrivID = 18

	//数据载入任务
	PrivID_CreateLoadTask PrivID = 19
	PrivID_QueryLoadTask  PrivID = 20 //查询数据载入任务的详情
	PrivID_UpdateLoadTask PrivID = 21
	PrivID_DeleteLoadTask PrivID = 22

	//工作流
	PrivID_CreateWorkflow PrivID = 23
	PrivID_RunWorkflow    PrivID = 24
	PrivID_QueryWorkflow  PrivID = 25
	PrivID_StopWorkflow   PrivID = 26
	PrivID_UpdateWorkflow PrivID = 27
	PrivID_DeleteWorkflow PrivID = 28

	//已废弃
	PrivID_CreateVolume_OLD PrivID = 29
	PrivID_QueryVolume_OLD  PrivID = 30
	PrivID_UpdateVolume_OLD PrivID = 31
	PrivID_DeleteVolume_OLD PrivID = 32
	PrivID_ExportVolume_OLD PrivID = 33

	//目录
	PrivID_CreateCatalog PrivID = 34
	PrivID_QueryCatalog  PrivID = 35
	PrivID_UpdateCatalog PrivID = 36
	PrivID_DeleteCatalog PrivID = 37

	//数据库
	PrivID_CreateDatabase PrivID = 38
	PrivID_QueryDatabase  PrivID = 39
	PrivID_UpdateDatabase PrivID = 52
	PrivID_DeleteDatabase PrivID = 53

	//告警
	PrivID_CreateAlterRule     PrivID = 40
	PrivID_QueryAlterRule      PrivID = 41
	PrivID_UpdateAlterRule     PrivID = 42
	PrivID_DeleteAlterRule     PrivID = 43
	PrivID_CreateAlterReceiver PrivID = 44
	PrivID_QueryAlterReceiver  PrivID = 45
	PrivID_UpdateAlterReceiver PrivID = 46
	PrivID_DeleteAlterReceiver PrivID = 47
	PrivID_QueryAlterLog       PrivID = 48

	//数据导出任务
	PrivID_CreateExportTask PrivID = 49
	PrivID_QueryExportTask  PrivID = 50 //查询数据导出任务的详情
	PrivID_DeleteExportTask PrivID = 51

	//数据卷
	PrivID_CreateVolume PrivID = 54
	PrivID_QueryVolume  PrivID = 55
	PrivID_UpdateVolume PrivID = 56
	PrivID_DeleteVolume PrivID = 57

	//以下为4.1新增的ID
	PrivID_GetConnector PrivID = 58 //查看某个连接器的详情
	PrivID_UseConnector PrivID = 59 //使用连接器做数据载入或导出

	PrivID_GetLoadTask PrivID = 60 //查看某个数据载入任务的详情

	PrivID_UpdateExportTask PrivID = 61 //更新数据导出任务的状态
	PrivID_GetExportTask    PrivID = 62 //查看某个数据导出任务的详情

	PrivID_GetWorkflow PrivID = 63 //查看某个工作流的详情(含分支对比、作业)

	PrivID_VolumeRead  PrivID = 64 //读取某个数据卷的内容
	PrivID_VolumeWrite PrivID = 65 //写入某个数据卷的内容

	PrivID_CreateTable    PrivID = 200 //新建表
	PrivID_ShowTables     PrivID = 201 //查询数据库下的所有表
	PrivID_AlterTable     PrivID = 202 //修改表的结构
	PrivID_DropTable      PrivID = 203 //删除表
	PrivID_CreateView     PrivID = 204 //创建视图
	PrivID_AlterView      PrivID = 205 //修改视图的结构
	PrivID_DropView       PrivID = 206 //删除视图
	PrivID_TableSelect    PrivID = 207 //查询表中的数据
	PrivID_TableInsert    PrivID = 208 //向表中插入数据
	PrivID_TableUpdate    PrivID = 209 //更新表中的数据
	PrivID_TableDelete    PrivID = 210 //删除表中的数据
	PrivID_TableTruncate  PrivID = 211 //截断表
	PrivID_TableReference PrivID = 212 //查询表的引用
	PrivID_TableIndex     PrivID = 213 //查询表的索引

	//知识库
	PrivID_CreateKnowledge PrivID = 300 //新建知识库
	PrivID_QueryKnowledge  PrivID = 301 //查询知识库的详情
	PrivID_UpdateKnowledge PrivID = 302 //更新知识库的内容
	PrivID_DeleteKnowledge PrivID = 303 //删除知识库
	PrivID_UseKnowledge    PrivID = 304 //使用知识库

	//发布订阅
	PrivID_CreatePublication  PrivID = 400 //新建发布
	PrivID_QueryPublication   PrivID = 401 //查询发布的详情
	PrivID_UpdatePublication  PrivID = 402 //更新发布的内容
	PrivID_DeletePublication  PrivID = 403 //删除发布
	PrivID_CreateSubscription PrivID = 404 //新建订阅
	PrivID_QuerySubscription  PrivID = 405 //查询订阅的详情
	PrivID_UpdateSubscription PrivID = 406 //更新订阅的内容
	PrivID_DeleteSubscription PrivID = 407 //删除订阅
	PrivID_UseSubscription    PrivID = 408 //使用订阅

	//4.1之后如果有新的类别，那么PrivID从500开始，每个类别100个值
)

// ============ PrivCode constants ============

const (
	//用户
	PrivCode_CreateUser       PrivCode = "U1"
	PrivCode_QueryUser        PrivCode = "U2"
	PrivCode_UpdatePassword   PrivCode = "U3"
	PrivCode_UpdateUserRole   PrivCode = "U4"
	PrivCode_UpdateUserInfo   PrivCode = "U5"
	PrivCode_UpdateUserStatus PrivCode = "U6"
	PrivCode_DeleteUser       PrivCode = "U7"
	PrivCode_QueryUserLog     PrivCode = "U8"

	//角色
	PrivCode_CreateRole       PrivCode = "R1"
	PrivCode_QueryRole        PrivCode = "R2"
	PrivCode_UpdateRoleInfo   PrivCode = "R3"
	PrivCode_UpdateRoleStatus PrivCode = "R4"
	PrivCode_DeleteRole       PrivCode = "R5"
	PrivCode_QueryRoleLog     PrivCode = "R6"

	//连接器
	PrivCode_CreateConnector PrivCode = "C1"
	PrivCode_QueryConnector  PrivCode = "C2" //查看连接器的列表
	PrivCode_UpdateConnector PrivCode = "C4" //4.1 从C3迁移到C4
	PrivCode_DeleteConnector PrivCode = "C5" //4.1 从C4迁移到C5

	//数据载入任务
	PrivCode_CreateLoadTask PrivCode = "L1"
	PrivCode_QueryLoadTask  PrivCode = "L2" //查看某个数据载入任务的列表
	PrivCode_UpdateLoadTask PrivCode = "L4" //4.1 从L3迁移到L4
	PrivCode_DeleteLoadTask PrivCode = "L5" //4.1 从L4迁移到L5

	//工作流
	PrivCode_CreateWorkflow PrivCode = "W1"
	PrivCode_RunWorkflow    PrivCode = "W2"
	PrivCode_QueryWorkflow  PrivCode = "W3" //查看工作流列表
	PrivCode_StopWorkflow   PrivCode = "W5" //4.1 从W4迁移到W5
	PrivCode_UpdateWorkflow PrivCode = "W6" //4.1 从W5迁移到W6
	PrivCode_DeleteWorkflow PrivCode = "W7" //4.1 从W6迁移到W7

	//已废弃
	PrivCode_CreateVolume_OLD PrivCode = "V1"
	PrivCode_QueryVolume_OLD  PrivCode = "V2"
	PrivCode_UpdateVolume_OLD PrivCode = "V3"
	PrivCode_DeleteVolume_OLD PrivCode = "V4"
	PrivCode_ExportVolume_OLD PrivCode = "V5"

	//目录
	PrivCode_CreateCatalog PrivCode = "DC1" //4.1 从D1迁移到DC1
	PrivCode_QueryCatalog  PrivCode = "DC2" //4.1 从D2迁移到DC2
	PrivCode_UpdateCatalog PrivCode = "DC3" //4.1 从D3迁移到DC3
	PrivCode_DeleteCatalog PrivCode = "DC4" //4.1 从D4迁移到DC4

	//数据库
	PrivCode_CreateDatabase PrivCode = "DB1" //4.1 从D5迁移到DB1
	PrivCode_QueryDatabase  PrivCode = "DB2" //4.1 从D6迁移到DB2
	PrivCode_UpdateDatabase PrivCode = "DB3" //4.1 从D7迁移到DB3
	PrivCode_DeleteDatabase PrivCode = "DB4" //4.1 从D8迁移到DB4

	//告警
	PrivCode_CreateAlterRule     PrivCode = "A1"
	PrivCode_QueryAlterRule      PrivCode = "A2"
	PrivCode_UpdateAlterRule     PrivCode = "A3"
	PrivCode_DeleteAlterRule     PrivCode = "A4"
	PrivCode_CreateAlterReceiver PrivCode = "A5"
	PrivCode_QueryAlterReceiver  PrivCode = "A6"
	PrivCode_UpdateAlterReceiver PrivCode = "A7"
	PrivCode_DeleteAlterReceiver PrivCode = "A8"
	PrivCode_QueryAlterLog       PrivCode = "A9"

	//数据导出任务
	PrivCode_CreateExportTask PrivCode = "E1"
	PrivCode_QueryExportTask  PrivCode = "E2"
	PrivCode_DeleteExportTask PrivCode = "E5" //4.1 从E3迁移到E5

	//数据卷
	PrivCode_CreateVolume PrivCode = "DV1" //4.1 从D9迁移到DV1
	PrivCode_QueryVolume  PrivCode = "DV2" //4.1 从D10迁移到DV2
	PrivCode_UpdateVolume PrivCode = "DV3" //4.1 从D11迁移到DV3
	PrivCode_DeleteVolume PrivCode = "DV4" //4.1 从D12迁移到DV4

	//以下内容为4.1新增的权限码,其中部分权限码是从4.0迁移到4.1的（就是码被新的含义占了）

	//连接器
	PrivCode_GetConnector PrivCode = "C3" //查看某个连接器的详情
	PrivCode_UseConnector PrivCode = "C6" //使用连接器做数据载入或导出

	//数据载入任务
	PrivCode_GetLoadTask PrivCode = "L3" //查看某个数据载入任务的详情

	//数据导出任务
	PrivCode_GetExportTask    PrivCode = "E3" //查看某个数据导出任务的详情
	PrivCode_UpdateExportTask PrivCode = "E4" //更新数据导出任务的状态

	//工作流
	PrivCode_GetWorkflow PrivCode = "W4" //查看某个工作流的详情(含对比，作业)

	//数据卷
	PrivCode_VolumeRead  PrivCode = "DV5" //读取某个数据卷的内容
	PrivCode_VolumeWrite PrivCode = "DV6" //写入某个数据卷的内容

	//表
	PrivCode_CreateTable    PrivCode = "DT1"
	PrivCode_ShowTables     PrivCode = "DT2"
	PrivCode_AlterTable     PrivCode = "DT3"
	PrivCode_DropTable      PrivCode = "DT4"
	PrivCode_CreateView     PrivCode = "DT5"
	PrivCode_AlterView      PrivCode = "DT6"
	PrivCode_DropView       PrivCode = "DT7"
	PrivCode_TableSelect    PrivCode = "DT8"
	PrivCode_TableInsert    PrivCode = "DT9"
	PrivCode_TableUpdate    PrivCode = "DT10"
	PrivCode_TableDelete    PrivCode = "DT11"
	PrivCode_TableTruncate  PrivCode = "DT12"
	PrivCode_TableReference PrivCode = "DT13"
	PrivCode_TableIndex     PrivCode = "DT14"

	//知识库
	PrivCode_CreateKnowledge PrivCode = "K1"
	PrivCode_QueryKnowledge  PrivCode = "K2"
	PrivCode_UpdateKnowledge PrivCode = "K3"
	PrivCode_DeleteKnowledge PrivCode = "K4"
	PrivCode_UseKnowledge    PrivCode = "K5"

	//发布订阅
	PrivCode_CreatePublication  PrivCode = "PS1"
	PrivCode_QueryPublication   PrivCode = "PS2"
	PrivCode_UpdatePublication  PrivCode = "PS3"
	PrivCode_DeletePublication  PrivCode = "PS4"
	PrivCode_QuerySubscription  PrivCode = "PS5"
	PrivCode_CreateSubscription PrivCode = "PS6"
	PrivCode_UpdateSubscription PrivCode = "PS7"
	PrivCode_DeleteSubscription PrivCode = "PS8"
)

var ObjTypeToPrivIDMap = map[ObjType][]PrivID{
	ObjTypeConnector:  {PrivID_UpdateConnector, PrivID_DeleteConnector, PrivID_GetConnector, PrivID_UseConnector},
	ObjTypeLoadTask:   {PrivID_UpdateLoadTask, PrivID_DeleteLoadTask, PrivID_GetLoadTask},
	ObjTypeExportTask: {PrivID_UpdateExportTask, PrivID_DeleteExportTask, PrivID_GetExportTask},
	ObjTypeWorkFlow:   {PrivID_UpdateWorkflow, PrivID_DeleteWorkflow, PrivID_GetWorkflow, PrivID_RunWorkflow, PrivID_StopWorkflow},
	ObjTypeAlarm:      {PrivID_QueryAlterRule, PrivID_UpdateAlterRule, PrivID_DeleteAlterRule, PrivID_QueryAlterReceiver, PrivID_UpdateAlterReceiver, PrivID_DeleteAlterReceiver, PrivID_QueryAlterLog},
	ObjTypeCatalog:    {PrivID_QueryCatalog, PrivID_UpdateCatalog, PrivID_DeleteCatalog, PrivID_CreateDatabase, PrivID_QueryDatabase},
	ObjTypeDatabase:   {PrivID_UpdateDatabase, PrivID_DeleteDatabase, PrivID_CreateVolume, PrivID_CreateTable, PrivID_ShowTables, PrivID_CreateView},
	ObjTypeVolume:     {PrivID_UpdateVolume, PrivID_DeleteVolume, PrivID_VolumeRead, PrivID_VolumeWrite},
	ObjTypeTable:      {PrivID_AlterTable, PrivID_DropTable, PrivID_AlterView, PrivID_DropView, PrivID_TableSelect, PrivID_TableInsert, PrivID_TableUpdate, PrivID_TableDelete, PrivID_TableTruncate, PrivID_TableReference, PrivID_TableIndex},
}

type CheckPriv struct {
	PrivID   PrivID       `json:"priv_id"`
	ObjectID PrivObjectID `json:"obj_id"`
}

// String returns the string representation of the PrivObjectID.
func (po PrivObjectID) String() string {
	return string(po)
}

// IntToPrivObjectID converts an int64 to a PrivObjectID.
//
// This is a convenience function for creating PrivObjectID values from integer IDs.
//
// Example:
//
//	objID := sdk.IntToPrivObjectID(123)
func IntToPrivObjectID(id int64) PrivObjectID {
	return PrivObjectID(fmt.Sprintf("%d", id))
}

// AuthorityCodeAndRule represents a privilege code with its associated rules.
type AuthorityCodeAndRule struct {
	Code            string             `json:"code"`
	BlackColumnList []string           `json:"black_column_list"`
	RuleList        []*TableRowColRule `json:"rule_list"`
}

// TableRowColRule represents a table row/column rule with expressions.
type TableRowColRule struct {
	Column         string                   `json:"column"`
	Relation       string                   `json:"relation"` // and or
	ExpressionList []*TableRowColExpression `json:"expression_list"`
}

// TableRowColExpression represents a single expression in a table row/column rule.
type TableRowColExpression struct {
	Operator   string   `json:"operator"` // = != like > >= < <= regexp_like
	Expression []string `json:"expression"`
	MatchType  string   `json:"match_type"` // c,i,m,n,u
}

type ObjPrivResponse struct {
	ObjID             string                  `json:"id"`
	ObjType           string                  `json:"category"`
	ObjName           string                  `json:"name"`
	AuthorityCodeList []*AuthorityCodeAndRule `json:"authority_code_list"`
}

type PrivObjectIDAndName struct {
	ObjectID   string                 `json:"id"`
	ObjectName string                 `json:"name"`
	DatabaseID string                 `json:"database_id"`
	ObjectType string                 `json:"type"`
	NodeList   []*PrivObjectIDAndName `json:"node_list"`
}

// ============ Models: Catalog types ============

type CatalogResponse struct {
	CatalogID     CatalogID `json:"id"`
	CatalogName   string    `json:"name"`
	Comment       string    `json:"description"`
	DatabaseCount int       `json:"database_count"`
	TableCount    int       `json:"table_count"`
	VolumeCount   int       `json:"volume_count"`
	FileCount     int       `json:"file_count"`
	Reserved      bool      `json:"reserved"`
	CreatedAt     string    `json:"created_at"`
	CreatedBy     string    `json:"created_by"`
	UpdatedAt     string    `json:"updated_at"`
	UpdatedBy     string    `json:"updated_by"`
}

type TreeNode struct {
	Typ                  string      `json:"type"`
	ID                   string      `json:"id"`
	Name                 string      `json:"name"`
	Description          string      `json:"description"`
	Reserved             bool        `json:"reserved"`
	HasWorkflowTargetRef bool        `json:"has_workflow_target_ref"`
	NodeList             []*TreeNode `json:"node_list"`
}

// ============ Models: Database types ============

type DatabaseResponse struct {
	DatabaseID   DatabaseID `json:"id"`
	DatabaseName string     `json:"name"`
	Comment      string     `json:"description"`
	TableCount   int        `json:"table_count"`
	VolumeCount  int        `json:"volume_count"`
	FileCount    int        `json:"file_count"`
	Reserved     bool       `json:"reserved"`
	CreatedAt    string     `json:"created_at"`
	CreatedBy    string     `json:"created_by"`
	UpdatedAt    string     `json:"updated_at"`
	UpdatedBy    string     `json:"updated_by"`
}

type DatabaseChildrenResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Typ           string `json:"type"`
	ChildrenCount int    `json:"children_count"`
	Size          int64  `json:"size"`
	Comment       string `json:"description"`
	Reserved      bool   `json:"reserved"`
	CreatedAt     string `json:"created_at"`
	CreatedBy     string `json:"created_by"`
	UpdatedAt     string `json:"updated_at"`
	UpdatedBy     string `json:"updated_by"`
}

// ============ Models: Volume types ============

type VolumeRefResp struct {
	VolumeID   VolumeID `json:"volume_id"`
	VolumeName string   `json:"volume_name"`
	RefType    string   `json:"ref_type"`
	RefID      string   `json:"ref_id"`
}

type VolumeChildrenResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	FileType       string `json:"file_type"`
	ShowType       string `json:"show_type"`
	FileExt        string `json:"file_ext"`
	OriginFileExt  string `json:"origin_file_ext"`
	RefFileID      string `json:"ref_file_id"`
	Size           int64  `json:"size"`
	VolumeID       string `json:"volume_id"`
	VolumeName     string `json:"volume_name"`
	VolumeReserved bool   `json:"volume_reserved"`
	RefWorkFlowID  string `json:"ref_workflow_id"`
	ParentID       string `json:"parent_id"`
	ShowPath       string `json:"show_path"`
	SavePath       string `json:"save_path"`
	CreatedAt      string `json:"created_at"`
	CreatedBy      string `json:"created_by"`
	UpdatedAt      string `json:"updated_at"`
}

// ============ Models: Table types ============

type TableRefResp struct {
	TableID   TableID `json:"table_id"`
	TableName string  `json:"table_name"`
	RefType   string  `json:"ref_type"`
	RefID     string  `json:"ref_id"`
}

// ============ Models: Role types ============

type RoleIDName struct {
	ID     RoleID   `json:"id"`
	Name   string   `json:"name"`
	Status string   `json:"status"`
	Codes  []string `json:"codes"`
}

// ============ Models: User types ============

type UserResponse struct {
	ID          UserID        `json:"id"`
	Name        string        `json:"name"`
	Status      string        `json:"status"`
	Phone       string        `json:"phone"`
	Email       string        `json:"email"`
	Reserved    bool          `json:"reserved"`
	RoleList    []*RoleIDName `json:"role_list"`
	LastLogin   string        `json:"last_login"`
	Description string        `json:"description"`
	CreatedAt   string        `json:"created_at"`
	UpdatedAt   string        `json:"updated_at"`
}

// ============ Models: File/Dedup types ============

// DedupBy represents the deduplication criteria.
type DedupBy string

const (
	// DedupByName deduplicates files by filename.
	DedupByName DedupBy = "name"
	// DedupByMD5 deduplicates files by MD5 hash.
	DedupByMD5 DedupBy = "md5"
)

// DedupStrategy represents the deduplication strategy.
type DedupStrategy string

const (
	// DedupStrategySkip skips duplicate files (does not upload).
	DedupStrategySkip DedupStrategy = "skip"
	// DedupStrategyReplace replaces duplicate files.
	DedupStrategyReplace DedupStrategy = "replace"
)

type DedupConfig struct {
	By       []string `json:"by,omitempty"`
	Strategy string   `json:"strategy,omitempty"`
}

// NewDedupConfig creates a new DedupConfig with the specified criteria and strategy.
//
// This is a helper function to create DedupConfig in a type-safe way.
// Use DedupBy constants for criteria and DedupStrategy constants for strategy.
//
// Example:
//
//	// Skip files that have the same name or MD5 hash
//	dedup := sdk.NewDedupConfig([]sdk.DedupBy{sdk.DedupByName, sdk.DedupByMD5}, sdk.DedupStrategySkip)
//
//	// Skip files that have the same name
//	dedup := sdk.NewDedupConfig([]sdk.DedupBy{sdk.DedupByName}, sdk.DedupStrategySkip)
func NewDedupConfig(by []DedupBy, strategy DedupStrategy) *DedupConfig {
	if len(by) == 0 {
		return nil
	}
	byStrings := make([]string, len(by))
	for i, b := range by {
		byStrings[i] = string(b)
	}
	return &DedupConfig{
		By:       byStrings,
		Strategy: string(strategy),
	}
}

// NewDedupConfigSkipByNameAndMD5 creates a DedupConfig that skips files with the same name or MD5 hash.
//
// This is a convenience function for the most common deduplication scenario.
//
// Example:
//
//	dedup := sdk.NewDedupConfigSkipByNameAndMD5()
//	resp, err := sdkClient.ImportLocalFileToVolume(ctx, filePath, volumeID, meta, dedup)
func NewDedupConfigSkipByNameAndMD5() *DedupConfig {
	return NewDedupConfig([]DedupBy{DedupByName, DedupByMD5}, DedupStrategySkip)
}

// NewDedupConfigSkipByName creates a DedupConfig that skips files with the same name.
//
// Example:
//
//	dedup := sdk.NewDedupConfigSkipByName()
//	resp, err := sdkClient.ImportLocalFileToVolume(ctx, filePath, volumeID, meta, dedup)
func NewDedupConfigSkipByName() *DedupConfig {
	return NewDedupConfig([]DedupBy{DedupByName}, DedupStrategySkip)
}

// NewDedupConfigSkipByMD5 creates a DedupConfig that skips files with the same MD5 hash.
//
// Example:
//
//	dedup := sdk.NewDedupConfigSkipByMD5()
//	resp, err := sdkClient.ImportLocalFileToVolume(ctx, filePath, volumeID, meta, dedup)
func NewDedupConfigSkipByMD5() *DedupConfig {
	return NewDedupConfig([]DedupBy{DedupByMD5}, DedupStrategySkip)
}

// ============ Infra: MO types ============

type Column struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	IsPk    bool   `json:"is_pk"`
	Default string `json:"default"`
	Comment string `json:"comment"`
}

type ColumnStats struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	MaxValue string `json:"max_value"`
	MinValue string `json:"min_value"`
}

// ============ Handler: Catalog types ============

type CatalogCreateRequest struct {
	CatalogName string `json:"name"`
	Comment     string `json:"description"`
}

type CatalogCreateResponse struct {
	CatalogID CatalogID `json:"id"`
}

type CatalogDeleteRequest struct {
	CatalogID CatalogID `json:"id"`
}

type CatalogDeleteResponse struct {
	CatalogID CatalogID `json:"id"`
}

type CatalogUpdateRequest struct {
	CatalogID   CatalogID `json:"id"`
	CatalogName string    `json:"name"`
	Comment     string    `json:"description"`
}

type CatalogUpdateResponse struct {
	CatalogID CatalogID `json:"id"`
}

type CatalogInfoRequest struct {
	CatalogID CatalogID `json:"id"`
}

type CatalogInfoResponse struct {
	CatalogID   CatalogID `json:"id"`
	CatalogName string    `json:"name"`
	Comment     string    `json:"description"`
}

type CatalogTreeResponse struct {
	Tree []*TreeNode `json:"tree"`
}

type CatalogListResponse struct {
	List []CatalogResponse `json:"list"`
}

type CatalogRefListRequest struct {
	CatalogID CatalogID `json:"id"`
}

type CatalogRefListResponse struct {
	List []*VolumeRefResp `json:"list"`
}

// ============ Handler: Database types ============

type DatabaseCreateRequest struct {
	DatabaseName string    `json:"name"`
	Comment      string    `json:"description"`
	CatalogID    CatalogID `json:"catalog_id"`
}

type DatabaseCreateResponse struct {
	DatabaseID DatabaseID `json:"id"`
}

type DatabaseDeleteRequest struct {
	DatabaseID DatabaseID `json:"id"`
}

type DatabaseDeleteResponse struct {
	DatabaseID DatabaseID `json:"id"`
}

type DatabaseUpdateRequest struct {
	DatabaseID DatabaseID `json:"id"`
	Comment    string     `json:"description"`
}

type DatabaseUpdateResponse struct {
	DatabaseID DatabaseID `json:"id"`
}

type DatabaseInfoRequest struct {
	DatabaseID DatabaseID `json:"id"`
}

type DatabaseInfoResponse struct {
	DatabaseID   DatabaseID `json:"id"`
	DatabaseName string     `json:"name"`
	Comment      string     `json:"description"`
	CreatedAt    string     `json:"created_at"`
	UpdatedAt    string     `json:"updated_at"`
}

type DatabaseListRequest struct {
	CatalogID CatalogID `json:"id"`
}

type DatabaseListResponse struct {
	List []DatabaseResponse `json:"list"`
}

type DatabaseChildrenRequest struct {
	DatabaseID DatabaseID `json:"id"`
}

// ChildrenResponse wraps the list of DatabaseChildrenResponse
type DatabaseChildrenResponseData struct {
	List []DatabaseChildrenResponse `json:"list"`
}

type DatabaseRefListRequest struct {
	DatabaseID DatabaseID `json:"id"`
}

type DatabaseRefListResponse struct {
	List []*VolumeRefResp `json:"list"`
}

// ============ Handler: Table types ============

type TableCreateRequest struct {
	DatabaseID DatabaseID `json:"database_id"`
	Name       string     `json:"name"`
	Columns    []Column   `json:"columns"`
	Comment    string     `json:"comment"`
}

type TableCreateResponse struct {
	TableID TableID `json:"id"`
}

type TableInfoRequest struct {
	TableID TableID `json:"id" binding:"required"`
	//如果是订阅表，那么TableID传-1，这个时候要根据库id和表名来获取信息。
	TableName  string     `json:"table_name"`
	DatabaseID DatabaseID `json:"database_id"`
}

type TableInfoResponse struct {
	Name      string        `json:"name"`
	Lines     int64         `json:"lines"`
	Size      int64         `json:"size"`
	Columns   []Column      `json:"columns"`
	Stats     []ColumnStats `json:"stats"`
	CreateSql string        `json:"create_sql"`
	CreatedAt string        `json:"created_at"`
	CreatedBy string        `json:"created_by"`
	Comment   string        `json:"comment"`
}

type MultiTableInfoRequest struct {
	TableList []TableInfoRequest `json:"table_list" binding:"required"`
}

type MultiTableInfoResponse struct {
	//key = fmt.Sprintf("%d %s", database_id, table_name)
	// 因为如果是订阅表会没有table_id，所以没法拿table_id做key
	// 而发送方一定是可以拿到 database_id和table_name 的
	InfoMap map[string]TableInfoResponse `json:"info_map" binding:"required"`
}

type TableOverview struct {
	DbName    string   `json:"db_name"`
	TableName string   `json:"table_name"`
	ColNames  []string `json:"col_names"`
}

type TableExistRequest struct {
	DatabaseID DatabaseID `json:"database_id"`
	Name       string     `json:"name"`
}

type TablePreviewRequest struct {
	TableID TableID `json:"id"`
	Lines   int     `json:"lines"`
}

type TablePreviewResponse struct {
	Columns []Column        `json:"columns"`
	Data    [][]interface{} `json:"data"`
}

type TableLoadRequest struct {
	TableID     TableID     `json:"id"`
	FileOption  FileOption  `json:"file_option"`
	TableOption TableOption `json:"table_option"`
}

type FileOption struct {
	DataFileUrl string    `json:"data_file_url"`
	Type        string    `json:"type"`
	StartRow    int       `json:"start_row"`
	CsvConfig   CsvConfig `json:"csv_config"`
}

type CsvConfig struct {
	Separator string `json:"separator"`
	Quote     string `json:"quote"`
	IsEscaped bool   `json:"is_escaped"`
}

type TableOption struct {
	ConflictPolicy    int                `json:"conflict_policy"`
	ColumnLoadOptions []ColumnLoadOption `json:"column_load_options"`
}

type ColumnLoadOption struct {
	ColName         string `json:"col_name"`
	DataFrom        int    `json:"data_from"`
	ColNumberInFile int    `json:"col_number_in_file"`
}

type TableLoadResponse struct {
	Lines int64 `json:"lines"`
}

type TableDownloadRequest struct {
	TableID TableID `json:"id"`
}

type TableDownloadResponse struct {
	Url string `json:"url"`
}

type TableDownloadDataRequest struct {
	ID int64 `json:"id"`
}

type TableTruncateRequest struct {
	TableID TableID `json:"id"`
}

type TableTruncateResponse struct{}

type TableDeleteRequest struct {
	TableID TableID `json:"id"`
}

type TableDeleteResponse struct{}

type TableFullPathRequest struct {
	TableIDList []TableID `json:"table_id_list"`
}

type TableFullPathResponse struct {
	TableFullPath []FullPath `json:"table_full_path"`
}

type TableRefListRequest struct {
	TableID TableID `json:"id"`
}

type TableRefListResponse struct {
	List []*TableRefResp `json:"list"`
}

// ============ Handler: Volume types ============

type VolumeCreateRequest struct {
	Name       string     `json:"name"`
	DatabaseID DatabaseID `json:"database_id"`
	Comment    string     `json:"description"`
}

type VolumeCreateResponse struct {
	VolumeID VolumeID `json:"id"`
}

type VolumeDeleteRequest struct {
	VolumeID VolumeID `json:"id"`
}

type VolumeDeleteResponse struct {
	VolumeID VolumeID `json:"id"`
}

type VolumeUpdateRequest struct {
	VolumeID VolumeID `json:"id"`
	Name     string   `json:"name"`
	Comment  string   `json:"description"`
}

type VolumeUpdateResponse struct {
	VolumeID VolumeID `json:"id"`
}

type VolumeInfoRequest struct {
	VolumeID VolumeID `json:"id"`
}

type VolumeInfoResponse struct {
	VolumeID   VolumeID `json:"id"`
	VolumeName string   `json:"name"`
	Comment    string   `json:"description"`
	Ref        bool     `json:"ref"`
	CreatedAt  string   `json:"created_at"`
	UpdatedAt  string   `json:"updated_at"`
}

type VolumeRefListRequest struct {
	VolumeID VolumeID `json:"id"`
}

type VolumeRefListResponse struct {
	List []*VolumeRefResp `json:"list"`
}

type VolumeFullPathRequest struct {
	DatabaseIDList []DatabaseID `json:"database_id_list"`
	VolumeIDList   []VolumeID   `json:"volume_id_list"`
	FolderIDList   []FileID     `json:"folder_id_list"`
}

type VolumeFullPathResponse struct {
	DatabaseFullPath []FullPath `json:"database_full_path"`
	VolumeFullPath   []FullPath `json:"volume_full_path"`
	FolderFullPath   []FullPath `json:"folder_full_path"`
}

type VolumeAddRefWorkflowRequest struct {
	VolumeID VolumeID `json:"id"`
}

type VolumeAddRefWorkflowResponse struct {
	VolumeID VolumeID `json:"id"`
}

type VolumeRemoveRefWorkflowRequest struct {
	VolumeID VolumeID `json:"id"`
}

type VolumeRemoveRefWorkflowResponse struct {
	VolumeID VolumeID `json:"id"`
}

// ============ Handler: File types ============

type FileCreateRequest struct {
	Name          string       `json:"name"`
	VolumeID      VolumeID     `json:"volume_id"`
	ParentID      FileID       `json:"parent_id"`
	Size          int64        `json:"size"`
	ShowType      string       `json:"show_type"`
	OriginFileExt string       `json:"origin_file_ext"`
	RefFileID     string       `json:"ref_file_id"`
	SavePath      string       `json:"save_path"`
	Hash          string       `json:"hash"`
	Dedup         *DedupConfig `json:"dedup,omitempty"`
}

type FileCreateResponse struct {
	FileID FileID `json:"id"`
	Name   string `json:"name"`
}

type FileUpdateRequest struct {
	FileID FileID `json:"id"`
	Name   string `json:"name"`
}

type FileUpdateResponse struct {
	FileID FileID `json:"id"`
}

type FileDeleteRequest struct {
	FileID FileID `json:"id"`
}

type FileDeleteResponse struct {
	FileID FileID `json:"id"`
}

type FileDeleteRefRequest struct {
	RefFileID string `json:"id"`
}

type FileDeleteRefResponse struct {
	FileID FileID `json:"id"`
}

type FileInfoRequest struct {
	FileID FileID `json:"id"`
}

type FileInfoResponse struct {
	ID            FileID `json:"id"`
	Name          string `json:"name"`
	FileType      string `json:"file_type"`
	ShowType      string `json:"show_type"`
	FileExt       string `json:"file_ext"`
	OriginFileExt string `json:"origin_file_ext"`
	RefFileID     string `json:"ref_file_id"`
	Size          int64  `json:"size"`
	ParentID      string `json:"parent_id"`
	VolumeID      string `json:"volume_id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

type FileListRequest struct {
	CommonCondition
	Keyword string `json:"keyword"`
}

type FileListResponse struct {
	Total int                      `json:"total"`
	List  []VolumeChildrenResponse `json:"list"`
}

type FileUploadRequest struct {
	Name     string   `json:"name"`
	VolumeID VolumeID `json:"volume_id"`
	ParentID FileID   `json:"parent_id"`
}

type FileUploadResponse struct {
	FileID FileID `json:"id"`
}

type FileDownloadRequest struct {
	FileID   FileID   `json:"file_id"`
	VolumeID VolumeID `json:"volume_id"`
}

type FileDownloadResponse struct {
	Url string `json:"link"`
}

type FilePreviewLinkRequest struct {
	FileID   FileID   `json:"file_id"`
	VolumeID VolumeID `json:"volume_id"`
}

type FilePreviewLinkResponse struct {
	Url string `json:"link"`
}

type FilePreviewStreamRequest struct {
	FileID FileID `json:"file_id"`
}

// ============ Handler: Folder types ============

type FolderCreateRequest struct {
	Name     string   `json:"name"`
	VolumeID VolumeID `json:"volume_id"`
	ParentID FileID   `json:"parent_id"`
}

type FolderCreateResponse struct {
	FolderID FileID `json:"id"`
	Name     string `json:"name"`
}

type FolderUpdateRequest struct {
	FolderID FileID `json:"id"`
	Name     string `json:"name"`
}

type FolderUpdateResponse struct {
	FolderID FileID `json:"id"`
}

type FolderDeleteRequest struct {
	FolderID FileID `json:"id"`
}

type FolderDeleteResponse struct {
	FolderID FileID `json:"id"`
}

type FolderCleanRequest struct {
	FolderID FileID `json:"id"`
}

type FolderCleanResponse struct {
	FolderID FileID `json:"id"`
}

type FolderRefListRequest struct {
	FolderID FileID `json:"id"`
}

type FolderRefListResponse struct {
	List []*VolumeRefResp `json:"list"`
}

// ============ Handler: Role types ============

type RoleCreateRequest struct {
	RoleName    string            `json:"name"`
	PrivList    []string          `json:"authority_code_list"`
	ObjPrivList []ObjPrivResponse `json:"obj_authority_code_list"`
	Comment     string            `json:"description"`
}

type RoleCreateResponse struct {
	RoleID RoleID `json:"id"`
}

type RoleDeleteRequest struct {
	RoleID RoleID `json:"id"`
}

type RoleDeleteResponse struct {
	RoleID RoleID `json:"id"`
}

type RoleInfoRequest struct {
	RoleID RoleID `json:"id"`
}

type RoleInfoResponse struct {
	RoleID           RoleID             `json:"id"`
	RoleName         string             `json:"name"`
	Status           string             `json:"status"`
	Reserved         bool               `json:"reserved"`
	Comment          string             `json:"description"`
	AuthorityList    []*PrivResponse    `json:"authority_list"`
	ObjAuthorityList []*ObjPrivResponse `json:"obj_authority_list"`
	CreatedAt        string             `json:"created_at"`
	UpdatedAt        string             `json:"updated_at"`
}

type PrivResponse struct {
	PrivCode string `json:"code"`
	PrivName string `json:"name"`
	Comment  string `json:"description"`
	ObjType  string `json:"category"`
}

type RoleListRequest struct {
	CommonCondition
	Keyword string `json:"keyword"`
}

type RoleListResponse struct {
	Total int                `json:"total"`
	List  []RoleInfoResponse `json:"role_list"`
}

type RoleUpdateInfoRequest struct {
	RoleID      RoleID            `json:"id"`
	PrivList    []string          `json:"authority_code_list"`
	ObjPrivList []ObjPrivResponse `json:"obj_authority_code_list"`
	Comment     string            `json:"description"`
}

type RoleUpdateInfoResponse struct {
	RoleID RoleID `json:"id"`
}

type RoleUpdateStatusRequest struct {
	RoleID RoleID `json:"id"`
	Action string `json:"action"`
}

type RoleUpdateStatusResponse struct {
	RoleID RoleID `json:"id"`
}

type RoleUpdateCodeListRequest struct {
	RoleID   RoleID   `json:"role_id"`
	ObjType  string   `json:"category"`
	ObjID    string   `json:"id"`
	CodeList []string `json:"code_list"`
}

type RoleUpdateCodeListResponse struct {
	RoleID RoleID `json:"role_id"`
}

type RoleUpdateRolesByObjectRequest struct {
	ObjID      string   `json:"id"`
	Code       string   `json:"code"`
	RoleIdList []RoleID `json:"role_id_list"`
}

type RoleUpdateRolesByObjectResponse struct{}

type RoleListByCategoryAndObjectRequest struct {
	ObjType string `json:"category"`
	ObjID   string `json:"id"`
}

type RoleListByCategoryAndObjectResponse struct {
	Total int                `json:"total"`
	List  []*PrivAndRoleList `json:"list"`
}

type PrivAndRoleList struct {
	Code     PrivCode              `json:"code"`
	RoleList []*SimpleRoleResponse `json:"role_list"`
}

type SimpleRoleResponse struct {
	RoleID    RoleID `json:"id"`
	RoleName  string `json:"name"`
	Status    string `json:"status"`
	Reserved  bool   `json:"reserved"`
	IsObjPriv bool   `json:"is_obj_priv"`
}

// ============ Handler: User types ============

type UserCreateRequest struct {
	UserName    string   `json:"name"`
	Password    string   `json:"password"`
	RoleIDList  []RoleID `json:"role_id_list"`
	Description string   `json:"description"`
	Phone       string   `json:"phone"`
	Email       string   `json:"email"`
	GetApiKey   bool     `json:"get_api_key"` // Whether to return API key in response
}

type UserCreateResponse struct {
	UserID UserID `json:"id"`
	ApiKey string `json:"api_key,omitempty"` // API key (only present if GetApiKey was true in request)
}

type UserDeleteUserRequest struct {
	UserID UserID `json:"id"`
}

type UserDeleteUserResponse struct {
	UserID UserID `json:"id"`
}

type UserDetailInfoRequest struct {
	UserID UserID `json:"id"`
}

type UserDetailInfoResponse struct {
	UserResponse
}

type UserListRequest struct {
	CommonCondition
	Keyword string `json:"keyword"`
}

type UserListResponse struct {
	Total int            `json:"total"`
	List  []UserResponse `json:"user_list"`
}

type UserUpdatePasswordRequest struct {
	UserID   UserID `json:"id"`
	Password string `json:"password"`
}

type UserUpdatePasswordResponse struct {
	UserID UserID `json:"id"`
}

type UserUpdateInfoRequest struct {
	UserID      UserID `json:"id"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Description string `json:"description"`
}

type UserUpdateInfoResponse struct {
	UserID UserID `json:"id"`
}

type UserUpdateRoleListRequest struct {
	UserID     UserID   `json:"id"`
	RoleIDList []RoleID `json:"role_id_list"`
}

type UserUpdateRoleListResponse struct {
	UserID UserID `json:"id"`
}

type UserUpdateStatusRequest struct {
	UserID UserID `json:"id"`
	Action string `json:"action"`
}

type UserUpdateStatusResponse struct {
	UserID UserID `json:"id"`
}

type UserMeUpdateInfoRequest struct {
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Description string `json:"description"`
}

type UserMeUpdateInfoResponse struct {
	UserID UserID `json:"id"`
}

type UserMeUpdatePasswordRequest struct {
	Password string `json:"password"`
}

type UserMeUpdatePasswordResponse struct {
	UserID UserID `json:"id"`
}

type UserMeInfoRequest struct{}

type UserMeInfoResponse struct {
	UserInfo             *UserResponse      `json:"user_info"`
	AuthorityCodeList    []string           `json:"authority_code_list"`
	ObjAuthorityCodeList []*ObjPrivResponse `json:"obj_authority_code_list"`
}

type UserApiKeyResponse struct {
	Key       string `json:"key"`
	CreatedAt string `json:"created_at"`
}

type UserApiKeyRefreshResonse struct{}

// ============ Handler: Priv types ============

type PrivGetAuthorizedObjectsRequest struct {
	PrivID        PrivID   `json:"priv_id"`
	ObjPrivIDList []PrivID `json:"obj_priv_id_list"`
}

type PrivGetAuthorizedObjectsResponse struct {
	AllAuthorized bool           `json:"all_authorized"`
	ObjectIDList  []PrivObjectID `json:"object_id_list"`
}

type PrivListObjByCategoryRequest struct {
	ObjType string `json:"category"`
}

type PrivListObjByCategoryResponse struct {
	Total int                    `json:"total"`
	List  []*PrivObjectIDAndName `json:"list"`
}

// ============ Handler: GenAI types ============

type GenAIGenerateNodeRequest struct {
	Node       string                            `json:"node"`
	Parameters map[string]map[string]interface{} `json:"parameters"`
}

type GenAIGenerateNodeResponse struct {
	Workflow interface{} `json:"workflow"` // Workflow type would need more context
}

type GenAICreateWorkflowRequest struct {
	FileIDs                []GenAIFileInfo     `json:"file_ids"`
	Steps                  []GenAIWorkflowStep `json:"steps"`
	Name                   string              `json:"name,omitempty"`
	SourceVolumeNames      []string            `json:"source_volume_names,omitempty"`
	SourceVolumeIDs        []string            `json:"source_volume_ids,omitempty"`
	TargetVolumeName       string              `json:"target_volume_name,omitempty"`
	TargetVolumeID         string              `json:"target_volume_id,omitempty"`
	CreateTargetVolumeName string              `json:"create_target_volume_name,omitempty"`
	ProcessMode            interface{}         `json:"process_mode,omitempty"`
	FileTypes              []int               `json:"file_types,omitempty"`
}

type GenAIFileInfo struct {
	ID       int64  `json:"id"`
	FileName string `json:"file_name,omitempty"`
}

type GenAIWorkflowStep struct {
	Node       string                            `json:"node"`
	Parameters map[string]map[string]interface{} `json:"parameters,omitempty"`
}

type GenAICreateWorkflowResponse struct {
	ID string `json:"id"`
}

type GenAICreatePipelineRequest struct {
	FileURLs  []string            `json:"file_urls"`
	FileNames []string            `json:"file_names,omitempty"`
	Steps     []GenAIWorkflowStep `json:"steps"`
}

type GenAICreatePipelineResponse struct {
	JobID string `json:"job_id,omitempty"`
}

type GenAIGetJobDetailRequest struct {
	JobID string `uri:"job_id"`
}

type GenAIWorkflowJobFileResponse struct {
	FileID       string `json:"file_id"`
	FileName     string `json:"file_name"`
	FileType     int    `json:"file_type"`
	FileStatus   string `json:"file_status"`
	ErrorMessage string `json:"error_message"`
	StartTime    string `json:"start_time"`
	EndTime      string `json:"end_time"`
}

type GenAIGetJobDetailResponse struct {
	Status string                         `json:"status"`
	Files  []GenAIWorkflowJobFileResponse `json:"files"`
}

type GenAIDownloadFileResultRequest struct {
	FileID string `uri:"file_id"`
}

// ============ Handler: Workflow types ============

// ProcessMode represents the processing mode for workflows.
// It describes the processing rhythm of a workflow.
type ProcessMode struct {
	Interval int `json:"interval"` // Processing interval in seconds
	Offset   int `json:"offset"`   // Processing offset in seconds
}

// WorkflowJobStatus represents the status of a workflow job.
type WorkflowJobStatus int

const (
	WorkflowJobStatusUnknown   WorkflowJobStatus = 0 // Unknown status
	WorkflowJobStatusRunning   WorkflowJobStatus = 1 // Job is running
	WorkflowJobStatusCompleted WorkflowJobStatus = 2 // Job completed successfully
	WorkflowJobStatusFailed    WorkflowJobStatus = 3 // Job failed
)

// String returns the string representation of the workflow job status.
func (s WorkflowJobStatus) String() string {
	switch s {
	case WorkflowJobStatusRunning:
		return "running"
	case WorkflowJobStatusCompleted:
		return "completed"
	case WorkflowJobStatusFailed:
		return "failed"
	case WorkflowJobStatusUnknown:
		return "unknown"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// WorkflowMetadata represents workflow metadata for creating a workflow.
// This is used by the CreateWorkflow API endpoint.
type WorkflowMetadata struct {
	Name                   string           `json:"name,omitempty"`
	SourceVolumeNames      []string         `json:"source_volume_names"`          // Required: must be present even if empty
	SourceVolumeIDs        []string         `json:"source_volume_ids"`            // Required: must be present even if empty
	TargetVolumeName       string           `json:"target_volume_name,omitempty"` // deprecated at moi 3.2.4
	TargetVolumeID         string           `json:"target_volume_id,omitempty"`
	CreateTargetVolumeName string           `json:"create_target_volume_name,omitempty"`
	ProcessMode            *ProcessMode     `json:"process_mode"` // Required: must be present even if empty
	FileTypes              []int            `json:"file_types,omitempty"`
	Workflow               *CatalogWorkflow `json:"workflow,omitempty"`
}

// CatalogWorkflow represents a workflow definition with nodes and connections.
type CatalogWorkflow struct {
	Nodes       []CatalogWorkflowNode       `json:"node"`
	Connections []CatalogWorkflowConnection `json:"connections"`
}

// CatalogWorkflowNode represents a workflow node definition.
type CatalogWorkflowNode struct {
	ID             string                            `json:"id"`
	Type           string                            `json:"type"`
	InitParameters map[string]map[string]interface{} `json:"init_parameters"` // Required: must be present, use empty map {} if no parameters
}

// CatalogWorkflowConnection represents a workflow connection definition.
type CatalogWorkflowConnection struct {
	Sender       string `json:"sender"`
	Receiver     string `json:"receiver"`                // Note: JSON may use "reciever" instead of "receiver" (typo)
	SenderPort   string `json:"sender_port,omitempty"`   // Optional port for sender component
	ReceiverPort string `json:"receiver_port,omitempty"` // Optional port for receiver component
}

// WorkflowCreateResponse represents the response from creating a workflow.
type WorkflowCreateResponse struct {
	CreatedAt         string `json:"created_at"`
	Creator           string `json:"creator"`
	Content           string `json:"content"`
	UpdatedAt         string `json:"updated_at"`
	Modifier          string `json:"modifier"`
	ID                string `json:"id"`
	FileTypes         string `json:"file_types"`
	Name              string `json:"name"`
	SourceVolumeIDs   string `json:"source_volume_ids"`
	UserID            string `json:"user_id"`
	SourceVolumeNames string `json:"source_volume_names"`
	GroupID           string `json:"group_id"`
	TargetVolumeID    string `json:"target_volume_id"`
	Version           string `json:"version"`
	FlowInterval      int    `json:"flow_interval"`
	TargetVolumeName  string `json:"target_volume_name"`
	Priority          int    `json:"priority"`
	FlowOffset        int    `json:"flow_offset"`
	Files             string `json:"files"`
}

// WorkflowJobListRequest represents a request to list workflow jobs.
type WorkflowJobListRequest struct {
	WorkflowID   string `json:"workflow_id,omitempty"`    // Filter by workflow ID
	SourceFileID string `json:"source_file_id,omitempty"` // Filter by source file ID
	Status       string `json:"status,omitempty"`         // Filter by job status
	Page         int    `json:"page,omitempty"`           // Page number (starts from 1, default 1)
	PageSize     int    `json:"page_size,omitempty"`      // Page size (default 20)
}

// WorkflowJob represents a workflow job in the list.
// This matches the API response structure from /byoa/api/v1/workflow_job
type WorkflowJob struct {
	JobID        string            `json:"id"`                       // Job ID (API returns "id")
	WorkflowID   string            `json:"workflow_id"`              // Workflow ID
	SourceFileID string            `json:"source_file_id,omitempty"` // Source file ID (not in API response, populated from query param)
	Status       WorkflowJobStatus `json:"status"`                   // Job status (API returns number: 1=running, 2=completed, 3=failed)
	StartTime    string            `json:"start_time"`               // Job start time
	EndTime      string            `json:"end_time"`                 // Job end time (null if not finished)
	// Note: API does not return created_at or updated_at fields
}

// workflowJobRaw represents the raw API response structure for a workflow job.
// This is used internally for unmarshaling, then converted to WorkflowJob.
type workflowJobRaw struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      int                    `json:"status"`
	StartTime   string                 `json:"start_time"`
	EndTime     *string                `json:"end_time"`              // Can be null
	Description map[string]interface{} `json:"description,omitempty"` // May contain triggerTaskID
}

// WorkflowJobListResponse represents the response from listing workflow jobs.
// This matches the API response structure: {"code":"ok","msg":"ok","data":{"total":1,"jobs":[...]}}
type WorkflowJobListResponse struct {
	Jobs  []WorkflowJob `json:"jobs"`  // List of workflow jobs (API returns "jobs" not "list")
	Total int           `json:"total"` // Total number of jobs
}

// ============ Handler: NL2SQL types ============

type NL2SQLRunSQLRequest struct {
	Operation  Nl2SqlOperationType `json:"operation"`
	Statement  string              `json:"statement"`
	DbNames    []string            `json:"db_names"`
	TableNames []DbAndTablesInfo   `json:"table_names"`
}

type DbAndTablesInfo struct {
	DbName     string   `json:"db_name"`
	TableNames []string `json:"table_names"`
}

// NL2SQLRunSQLResponse wraps the results returned by the NL2SQL run_sql API.
type NL2SQLRunSQLResponse struct {
	Results []NL2SQLResult `json:"results"`
}

// NL2SQLResult contains the column metadata and row data for a single NL2SQL statement.
type NL2SQLResult struct {
	Columns []string    `json:"columns"`
	Rows    []NL2SQLRow `json:"rows"`
}

// NL2SQLRow represents one row in an NL2SQL result set.
type NL2SQLRow []string

// ============ Models: NL2SQL Knowledge types ============

type Nl2SqlKnowledgeID int64

type Nl2SqlKnowledgeResponse struct {
	ID        Nl2SqlKnowledgeID      `json:"id"`
	Type      string                 `json:"type"`
	Key       string                 `json:"key"`
	Value     []string               `json:"value"`
	Embedding []float64              `json:"embedding,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type Nl2SqlOperationType string

const (
	ShowTable       Nl2SqlOperationType = "show_table"
	DescTable       Nl2SqlOperationType = "desc_table"
	RunSQL          Nl2SqlOperationType = "run_sql"
	Select_3        Nl2SqlOperationType = "select_3"
	ShowCreateTable Nl2SqlOperationType = "show_create_table"
	ShowDatabases   Nl2SqlOperationType = "show_databases"
)

// ============ Handler: NL2SQL Knowledge types ============

type NL2SQLKnowledgeCreateRequest struct {
	Type            string    `json:"knowledge_type"`
	Key             string    `json:"knowledge_key"`
	Value           []string  `json:"knowledge_value"`
	Embedding       []float64 `json:"embedding"`
	AssociateTables []string  `json:"associate_tables"`
	ExplanationType string    `json:"explanation_type"`
}

type NL2SQLKnowledgeCreateResponse struct {
	ID Nl2SqlKnowledgeID `json:"id"`
}

type NL2SQLKnowledgeUpdateRequest struct {
	ID              Nl2SqlKnowledgeID `json:"id"`
	Type            string            `json:"knowledge_type"`
	Key             string            `json:"knowledge_key"`
	Value           []string          `json:"knowledge_value"`
	Embedding       []float64         `json:"embedding"`
	AssociateTables []string          `json:"associate_tables"`
	ExplanationType string            `json:"explanation_type"`
}

type NL2SQLKnowledgeUpdateResponse struct {
	ID Nl2SqlKnowledgeID `json:"id"`
}

type NL2SQLKnowledgeDeleteRequest struct {
	ID Nl2SqlKnowledgeID `json:"id"`
}

type NL2SQLKnowledgeDeleteResponse struct {
	ID Nl2SqlKnowledgeID `json:"id"`
}

type NL2SQLKnowledgeGetRequest struct {
	ID Nl2SqlKnowledgeID `json:"id"`
}

type NL2SQLKnowledgeGetResponse struct {
	*Nl2SqlKnowledgeResponse
}

type NL2SQLKnowledgeListRequest struct {
	Type       string `json:"knowledge_type"`
	PageNumber int    `json:"page_number"`
	PageSize   int    `json:"page_size"`
}

type NL2SQLKnowledgeListResponse struct {
	List  []*Nl2SqlKnowledgeResponse `json:"list"`
	Total int64                      `json:"total"`
}

type NL2SQLKnowledgeSearchRequest struct {
	Type       string `json:"knowledge_type"`
	Key        string `json:"knowledge_key"`
	PageNumber int    `json:"page_number"`
	PageSize   int    `json:"page_size"`
}

type NL2SQLKnowledgeSearchResponse struct {
	List  []*Nl2SqlKnowledgeResponse `json:"list"`
	Total int64                      `json:"total"`
}

// ============ Handler: Log types ============

type LogLogResponse struct {
	LogActionType string `json:"type"`
	UserName      string `json:"user_name"`
	RoleName      string `json:"role_name"`
	CreatedAt     string `json:"created_at"`
	Status        string `json:"status"`
	Description   string `json:"description"`
}

type LogLogListRequest struct {
	CommonCondition
	Keyword string `json:"keyword"`
}

type LogLogListResponse struct {
	Total int              `json:"total"`
	List  []LogLogResponse `json:"role_list"`
}

// ============ Models: LLM Proxy types ============

// LLMTag represents a tag used in LLM Proxy (sessions and messages).
type LLMTag struct {
	Name      string `json:"name"`       // Tag name
	Source    string `json:"source"`     // Tag source (application name)
	CreatedAt int64  `json:"created_at"` // Creation time (Unix timestamp in seconds)
	UpdatedAt int64  `json:"updated_at"` // Update time (Unix timestamp in seconds)
}

// LLMSession represents a session in LLM Proxy.
type LLMSession struct {
	ID        int64    `json:"id"`         // Session ID
	Title     string   `json:"title"`      // Session title
	Source    string   `json:"source"`     // Session source (application name)
	UserID    string   `json:"user_id"`    // User ID
	Config    string   `json:"config"`     // Session configuration (JSON string, application-defined)
	Tags      []LLMTag `json:"tags"`       // Tags bound to the session
	CreatedAt int64    `json:"created_at"` // Creation time (Unix timestamp in seconds)
	UpdatedAt int64    `json:"updated_at"` // Update time (Unix timestamp in seconds)
}

// LLMSessionCreateRequest represents a request to create a session.
type LLMSessionCreateRequest struct {
	Title  string   `json:"title"`            // Required: Session title
	Source string   `json:"source"`           // Required: Session source (application name)
	UserID string   `json:"user_id"`          // Required: User ID
	Config string   `json:"config,omitempty"` // Optional: Session configuration (JSON string)
	Tags   []string `json:"tags,omitempty"`   // Optional: Tag names list
}

// LLMSessionListRequest represents a request to list sessions.
type LLMSessionListRequest struct {
	UserID   string   `json:"user_id,omitempty"`   // Filter by user ID
	Source   string   `json:"source,omitempty"`    // Filter by source
	Keyword  string   `json:"keyword,omitempty"`   // Keyword search (title)
	Tags     []string `json:"tags,omitempty"`      // Tag filter (comma-separated, requires all match)
	Page     int      `json:"page,omitempty"`      // Page number (starts from 1, default 1)
	PageSize int      `json:"page_size,omitempty"` // Page size (default 20, max 100)
}

// LLMSessionListResponse represents a response from listing sessions.
type LLMSessionListResponse struct {
	Sessions []LLMSession `json:"sessions"`
	Total    int64        `json:"total"`     // Total number of records
	Page     int          `json:"page"`      // Current page number
	PageSize int          `json:"page_size"` // Page size
}

// LLMSessionUpdateRequest represents a request to update a session.
type LLMSessionUpdateRequest struct {
	Title  *string   `json:"title,omitempty"`  // Session title
	Source *string   `json:"source,omitempty"` // Session source
	Config *string   `json:"config,omitempty"` // Session configuration
	Tags   *[]string `json:"tags,omitempty"`   // Tag list (complete replacement)
}

// LLMSessionDeleteResponse represents a response from deleting a session.
type LLMSessionDeleteResponse struct {
	Message string `json:"message"`
}

// LLMMessageRole represents the role of a message.
type LLMMessageRole string

const (
	LLMMessageRoleUser      LLMMessageRole = "user"       // User message
	LLMMessageRoleSystem    LLMMessageRole = "system"     // System message
	LLMMessageRoleAssistant LLMMessageRole = "assistant"  // Assistant reply
	LLMMessageRoleAgentTool LLMMessageRole = "agent-tool" // Agent tool call
)

// LLMMessageStatus represents the status of a message.
type LLMMessageStatus string

const (
	LLMMessageStatusSuccess LLMMessageStatus = "success" // Success
	LLMMessageStatusFailed  LLMMessageStatus = "failed"  // Failed
	LLMMessageStatusRetry   LLMMessageStatus = "retry"   // Retry
	LLMMessageStatusAborted LLMMessageStatus = "aborted" // Aborted
)

// LLMChatMessage represents a chat message in LLM Proxy.
type LLMChatMessage struct {
	ID               int64            `json:"id"`                // Message ID
	UserID           string           `json:"user_id"`           // User ID
	SessionID        *int64           `json:"session_id"`        // Session ID (optional)
	Source           string           `json:"source"`            // Application name
	Role             LLMMessageRole   `json:"role"`              // Message role
	OriginalContent  string           `json:"original_content"`  // Original content (user's original input)
	Content          string           `json:"content"`           // Actual content sent to LLM
	Model            string           `json:"model"`             // Model name used
	Status           LLMMessageStatus `json:"status"`            // Status
	Response         string           `json:"response"`          // LLM reply content
	ModifiedResponse string           `json:"modified_response"` // Modified reply content
	Config           string           `json:"config"`            // Message configuration (JSON string, application-defined)
	Tags             []LLMTag         `json:"tags"`              // Tags bound to the message
	CreatedAt        int64            `json:"created_at"`        // Creation time (Unix timestamp in seconds)
	UpdatedAt        int64            `json:"updated_at"`        // Update time (Unix timestamp in seconds)
}

// LLMChatMessageCreateRequest represents a request to create a chat message.
type LLMChatMessageCreateRequest struct {
	UserID          string           `json:"user_id"`                    // Required: User ID
	SessionID       *int64           `json:"session_id,omitempty"`       // Optional: Session ID
	Source          string           `json:"source"`                     // Required: Application name
	Role            LLMMessageRole   `json:"role"`                       // Required: Message role
	OriginalContent string           `json:"original_content,omitempty"` // Optional: User's original input
	Content         string           `json:"content"`                    // Required: Actual content sent to LLM
	Model           string           `json:"model"`                      // Required: Model name
	Status          LLMMessageStatus `json:"status,omitempty"`           // Optional: Message status (default: success)
	Response        string           `json:"response,omitempty"`         // Optional: LLM reply content
	Config          string           `json:"config,omitempty"`           // Optional: Message configuration (JSON string)
	Tags            []string         `json:"tags,omitempty"`             // Optional: Tag names list
}

// LLMChatMessageListRequest represents a request to list chat messages.
type LLMChatMessageListRequest struct {
	UserID    string           `json:"user_id"`              // Required: User ID
	SessionID *int64           `json:"session_id,omitempty"` // Optional: Session ID
	Source    string           `json:"source,omitempty"`     // Optional: Application name
	Role      LLMMessageRole   `json:"role,omitempty"`       // Optional: Message role
	Status    LLMMessageStatus `json:"status,omitempty"`     // Optional: Message status
	Tags      []string         `json:"tags,omitempty"`       // Optional: Tag filter (comma-separated, requires all match)
	Page      int              `json:"page,omitempty"`       // Optional: Page number (starts from 1, default 1)
	PageSize  int              `json:"page_size,omitempty"`  // Optional: Page size (default 20, max 100)
}

// LLMChatMessageListResponse represents a response from listing chat messages.
type LLMChatMessageListResponse struct {
	Messages []LLMChatMessage `json:"messages"`
	Total    int64            `json:"total"`     // Total number of records
	Page     int              `json:"page"`      // Current page number
	PageSize int              `json:"page_size"` // Page size
}

// LLMChatMessageUpdateRequest represents a request to update a chat message.
type LLMChatMessageUpdateRequest struct {
	Status           *LLMMessageStatus `json:"status,omitempty"`            // Message status
	Response         *string           `json:"response,omitempty"`          // LLM reply content (for streaming, use CONCAT to append)
	ModifiedResponse *string           `json:"modified_response,omitempty"` // Modified reply content
	Content          *string           `json:"content,omitempty"`           // Actual content sent to LLM
	Config           *string           `json:"config,omitempty"`            // Message configuration (JSON string)
	Tags             *[]string         `json:"tags,omitempty"`              // Tag list (complete replacement)
}

// LLMChatMessageDeleteResponse represents a response from deleting a chat message.
type LLMChatMessageDeleteResponse struct {
	Message string `json:"message"`
}

// LLMChatMessageTagsUpdateRequest represents a request to update message tags.
type LLMChatMessageTagsUpdateRequest struct {
	Tags []string `json:"tags"` // Required: Tag list (complete replacement)
}

// LLMChatMessageTagDeleteResponse represents a response from deleting a message tag.
type LLMChatMessageTagDeleteResponse struct {
	Message string `json:"message"`
}

// LLMSessionMessagesListRequest represents a request to list session messages.
type LLMSessionMessagesListRequest struct {
	Source string           `json:"source,omitempty"` // Filter by source
	Role   LLMMessageRole   `json:"role,omitempty"`   // Filter by role
	Status LLMMessageStatus `json:"status,omitempty"` // Filter by status
	Model  string           `json:"model,omitempty"`  // Filter by model
	After  *int64           `json:"after,omitempty"`  // Get messages after this message ID (exclusive, > relation)
	Limit  *int             `json:"limit,omitempty"`  // Limit number of messages to return (default 20, max 100)
}

// LLMLatestCompletedMessageResponse represents a response from getting the latest completed message ID.
type LLMLatestCompletedMessageResponse struct {
	SessionID int64 `json:"session_id"`
	MessageID int64 `json:"message_id"`
}

// LLMModifySessionMessageResponseResponse represents a response from modifying a session message's modified response.
type LLMModifySessionMessageResponseResponse struct {
	Message          string `json:"message"`
	SessionID        int64  `json:"session_id"`
	MessageID        int64  `json:"message_id"`
	ModifiedResponse string `json:"modified_response"`
}

// LLMAppendSessionMessageModifiedResponseResponse represents a response from appending to a session message's modified response.
type LLMAppendSessionMessageModifiedResponseResponse struct {
	Message       string `json:"message"`
	SessionID     int64  `json:"session_id"`
	MessageID     int64  `json:"message_id"`
	AppendContent string `json:"append_content"`
}

// ============ Handler: Data Asking types ============

// DataAskingTableConfig represents table configuration for NL2SQL in data asking context.
// This is different from the TableConfig used in table operations.
type DataAskingTableConfig struct {
	Type       string   `json:"type"`                  // "all", "none", "specified"
	DbName     string   `json:"db_name"`               // Required: database name
	DatabaseID *int     `json:"database_id,omitempty"` // Database ID, used when type is "all"
	TableList  []string `json:"table_list,omitempty"`  // Table name list, used when type is "specified"
}

// FileConfig represents file configuration for RAG.
type FileConfig struct {
	Type       string   `json:"type"`                   // "all", "none", "specified"
	DatabaseID *int     `json:"database_id,omitempty"`  // Database ID, used when type is "all"
	FileIDList []string `json:"file_id_list,omitempty"` // File ID list, used when type is "specified"
}

// FilterConditions represents filter conditions.
type FilterConditions struct {
	Type string `json:"type"` // "all", "non_inter_data"
}

// CodeGroup represents a code group.
type CodeGroup struct {
	Code   string   `json:"code"`   // Parent-level code
	Name   string   `json:"name"`   // Code group name
	Values []string `json:"values"` // Code value list
}

// DataScope represents data scope configuration.
type DataScope struct {
	Type      string      `json:"type"`                // "all", "specified"
	CodeType  *int        `json:"code_type,omitempty"` // 0-company, 1-business unit
	CodeGroup []CodeGroup `json:"code_group,omitempty"`
}

// DataSource represents data source configuration.
type DataSource struct {
	Type   string                 `json:"type"` // "all", "specified"
	Tables *DataAskingTableConfig `json:"tables,omitempty"`
	Files  *FileConfig            `json:"files,omitempty"`
}

// DataAnalysisConfig represents data analysis configuration.
type DataAnalysisConfig struct {
	MCPServerURL     *string           `json:"mcp_server_url,omitempty"`   // MCP server URL
	DataObjectType   string            `json:"data_object_type,omitempty"` // "default", "audit_related" (default: "default")
	DataCategory     string            `json:"data_category,omitempty"`    // "admin", "common" (default: "admin")
	FilterConditions *FilterConditions `json:"filter_conditions,omitempty"`
	DataSource       *DataSource       `json:"data_source,omitempty"`
	DataScope        *DataScope        `json:"data_scope,omitempty"`
}

// DataAnalysisRequest represents a request for data analysis.
type DataAnalysisRequest struct {
	Question    string              `json:"question"`
	Source      *string             `json:"source,omitempty"`
	SessionID   *string             `json:"session_id,omitempty"`
	SessionName *string             `json:"session_name,omitempty"`
	Config      *DataAnalysisConfig `json:"config,omitempty"`
}

// QuestionType represents the classification result of a question.
type QuestionType struct {
	Type       string  `json:"type"` // "query", "attribution"
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// DataAnalysisStreamEvent represents a single event in the SSE stream.
// The actual structure depends on the event type.
//
// Common event formats:
//   - init event: step_type="init", data contains request_id and session_title
//   - classification event: type="classification"
//   - events with step_type field: StepType field contains the step type (e.g., "init", "sql_generated")
//   - events with source field: Source field indicates the source (e.g., "rag", "nl2sql")
type DataAnalysisStreamEvent struct {
	Type   string                 `json:"type,omitempty"`
	Source string                 `json:"source,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
	// For events that don't have a "type" field but have other fields directly
	// (e.g., step_type, step_name from NL2SQL, or step_type="init" for initialization)
	StepType string `json:"step_type,omitempty"`
	StepName string `json:"step_name,omitempty"`
	// Raw JSON data for flexible parsing
	RawData json.RawMessage `json:"-"`
}

// InitEventData represents the data field in an init event.
type InitEventData struct {
	RequestID    string `json:"request_id"`
	SessionTitle string `json:"session_title"`
}

// GetInitEventData extracts request_id and session_title from an init event.
// Returns nil if the event is not an init event or if the data cannot be parsed.
//
// Example:
//
//	event, err := stream.ReadEvent()
//	if err != nil {
//		return err
//	}
//	if event.StepType == "init" {
//		initData := event.GetInitEventData()
//		if initData != nil {
//			fmt.Printf("Request ID: %s, Session Title: %s\n", initData.RequestID, initData.SessionTitle)
//		}
//	}
func (e *DataAnalysisStreamEvent) GetInitEventData() *InitEventData {
	if e.StepType != "init" {
		return nil
	}

	// Try to extract from Data field first
	if e.Data != nil {
		requestID, _ := e.Data["request_id"].(string)
		sessionTitle, _ := e.Data["session_title"].(string)
		if requestID != "" {
			return &InitEventData{
				RequestID:    requestID,
				SessionTitle: sessionTitle,
			}
		}
	}

	// Try to parse from RawData
	if len(e.RawData) > 0 {
		var eventData struct {
			Data InitEventData `json:"data"`
		}
		if err := json.Unmarshal(e.RawData, &eventData); err == nil {
			if eventData.Data.RequestID != "" {
				return &eventData.Data
			}
		}
	}

	return nil
}

// CancelAnalyzeRequest represents a request to cancel a data analysis request.
type CancelAnalyzeRequest struct {
	RequestID string `json:"request_id"` // Required: The request ID of the analysis to cancel
}

// CancelAnalyzeResponse represents the response from canceling a data analysis request.
type CancelAnalyzeResponse struct {
	RequestID string `json:"request_id"` // The request ID that was cancelled
	Status    string `json:"status"`     // Status after cancellation (typically "cancelled")
	UserID    string `json:"user_id"`    // User ID who cancelled the request
	UserName  string `json:"user_name"`  // User name who cancelled the request
}

// ============ Handler: Task types ============

type TaskID int64

// TaskInfoRequest represents a request to get task information.
type TaskInfoRequest struct {
	TaskID TaskID `json:"task_id" form:"task_id"`
}

// TaskInfoResponse represents task information response.
type TaskInfoResponse struct {
	ID                  string                 `json:"id"`
	SourceConnectorId   uint64                 `json:"source_connector_id"`
	SourceConnectorType string                 `json:"source_connector_type"`
	VolumeID            string                 `json:"volume_id"`
	VolumeName          string                 `json:"volume_name"`
	VolumePath          *FullPath              `json:"volume_path,omitempty"`
	Name                string                 `json:"name"`
	Creator             string                 `json:"creator"`
	Status              string                 `json:"status"`
	SourceConfig        map[string]interface{} `json:"source_config,omitempty"`
	StartAt             string                 `json:"start_at,omitempty"`
	EndAt               string                 `json:"end_at,omitempty"`
	CreatedAt           string                 `json:"created_at"`
	UpdatedAt           string                 `json:"updated_at"`
	ConnectorName       string                 `json:"connector_name,omitempty"`
	TablePath           *FullPath              `json:"table_path,omitempty"`
	SourceFiles         [][]string             `json:"source_files,omitempty"`
	LoadResults         []*LoadResult          `json:"load_results,omitempty"`
}

// LoadResult represents a single file load result.
type LoadResult struct {
	Lines  int64  `json:"lines"`
	Reason string `json:"reason,omitempty"`
}
