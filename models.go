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

type FullPath struct {
	IDList   []string `json:"id_list"`
	NameList []string `json:"name_list"`
}

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

	//4.1之后如果有新的类别，那么PrivID从300开始，每个类别100个值
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
)

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
	Code     string             `json:"code"`
	RuleList []*TableRowColRule `json:"rule_list"`
}

// TableRowColRule represents a table row/column rule with expressions.
type TableRowColRule struct {
	Column         string                   `json:"column"`
	Relation       string                   `json:"relation"` // and or
	ExpressionList []*TableRowColExpression `json:"expression_list"`
}

// TableRowColExpression represents a single expression in a table row/column rule.
type TableRowColExpression struct {
	Operator   string `json:"operator"` // = != like > >= < <=
	Expression string `json:"expression"`
}

type ObjPrivResponse struct {
	ObjID             string                  `json:"id"`
	ObjType           string                  `json:"category"`
	ObjName           string                  `json:"name"`
	AuthorityCodeList []*AuthorityCodeAndRule `json:"authority_code_list"`
}

type PrivObjectIDAndName struct {
	ObjectID   string `json:"id"`
	ObjectName string `json:"name"`
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

type DedupConfig struct {
	By       []string `json:"by,omitempty"`
	Strategy string   `json:"strategy,omitempty"`
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
	TableID TableID `json:"id"`
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
}

type UserCreateResponse struct {
	UserID UserID `json:"id"`
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
	ID              int64            `json:"id"`               // Message ID
	UserID          string           `json:"user_id"`          // User ID
	SessionID       *int64           `json:"session_id"`       // Session ID (optional)
	Source          string           `json:"source"`           // Application name
	Role            LLMMessageRole   `json:"role"`             // Message role
	OriginalContent string           `json:"original_content"` // Original content (user's original input)
	Content         string           `json:"content"`          // Actual content sent to LLM
	Model           string           `json:"model"`            // Model name used
	Status          LLMMessageStatus `json:"status"`           // Status
	Response        string           `json:"response"`         // LLM reply content
	Tags            []LLMTag         `json:"tags"`             // Tags bound to the message
	CreatedAt       int64            `json:"created_at"`       // Creation time (Unix timestamp in seconds)
	UpdatedAt       int64            `json:"updated_at"`       // Update time (Unix timestamp in seconds)
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
	Status   *LLMMessageStatus `json:"status,omitempty"`   // Message status
	Response *string           `json:"response,omitempty"` // LLM reply content (for streaming, use CONCAT to append)
	Content  *string           `json:"content,omitempty"`  // Actual content sent to LLM
	Tags     *[]string         `json:"tags,omitempty"`     // Tag list (complete replacement)
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
}

// LLMLatestCompletedMessageResponse represents a response from getting the latest completed message ID.
type LLMLatestCompletedMessageResponse struct {
	SessionID int64 `json:"session_id"`
	MessageID int64 `json:"message_id"`
}

// ============ Handler: Data Asking types ============

// DataAskingTableConfig represents table configuration for NL2SQL in data asking context.
// This is different from the TableConfig used in table operations.
type DataAskingTableConfig struct {
	Type      string   `json:"type"` // "all", "none", "specified"
	DbName    *string  `json:"db_name,omitempty"`
	TableList []string `json:"table_list,omitempty"`
}

// FileConfig represents file configuration for RAG.
type FileConfig struct {
	Type             string   `json:"type"` // "all", "none", "specified"
	TargetVolumeName *string  `json:"target_volume_name,omitempty"`
	TargetVolumeID   *string  `json:"target_volume_id,omitempty"`
	FileIDList       []string `json:"file_id_list,omitempty"`
}

// FilterConditions represents filter conditions.
type FilterConditions struct {
	Type string `json:"type"` // "all", "non_inter_data"
}

// CodeGroup represents a code group.
type CodeGroup struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
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
	DataCategory     string            `json:"data_category"` // "admin", "common"
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
type DataAnalysisStreamEvent struct {
	Type   string                 `json:"type,omitempty"`
	Source string                 `json:"source,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
	// For events that don't have a "type" field but have other fields directly
	// (e.g., step_type, step_name from NL2SQL)
	StepType string `json:"step_type,omitempty"`
	StepName string `json:"step_name,omitempty"`
	// Raw JSON data for flexible parsing
	RawData json.RawMessage `json:"-"`
}
