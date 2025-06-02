package main

import (
	"encoding/json"
	"fmt"
	"time"

	// 关键改动：导入 "code" 子包
	"github.com/xuperchain/contract-sdk-go/code"
	"github.com/xuperchain/contract-sdk-go/driver" // driver.Serve 需要这个包
)

// DigitalCopyright 定义了数字版权的结构体
type DigitalCopyright struct {
	CopyrightID        string   `json:"copyright_id"`        // 版权唯一ID (主键)
	WorkName           string   `json:"work_name"`           // 作品名称
	Author             string   `json:"author"`              // 作者名称
	OwnerAddress       string   `json:"owner_address"`       // 当前版权持有人地址
	RegistrationDate   string   `json:"registration_date"`   // 登记日期 (例如 "YYYY-MM-DD HH:MM:SS")
	WorkType           string   `json:"work_type"`           // 作品类型 (例如 "文字", "图片", "音视频")
	Description        string   `json:"description"`         // 作品描述
	TransactionHistory []string `json:"transaction_history"` // 交易/变更历史记录
	IsDeleted          bool     `json:"is_deleted"`          // 标记是否已删除
}

// CopyrightManager 是合约结构体
// 按照您截图的示例，合约结构体可以是空的，方法直接定义在其指针类型上
type CopyrightManager struct {
}

// Initialize 方法在合约部署时调用
// args: {"creator": "address_of_creator"}
// 关键改动：使用 code.Context 和 code.Response
func (cm *CopyrightManager) Initialize(ctx code.Context) code.Response {
	args := struct {
		Creator string `json:"creator"`
	}{}
	// 尝试从 ctx.Args() 获取部署参数
	deployArgs := ctx.Args()
	creatorBytes, ok := deployArgs["creator"]
	if !ok {
		return code.Errors("missing creator argument during initialization")
	}
	args.Creator = string(creatorBytes)

	if args.Creator == "" {
		return code.Errors("creator address cannot be empty")
	}

	// 将合约创建者信息存储到链上，键为 "contract_creator"
	err := ctx.PutObject([]byte("contract_creator"), []byte(args.Creator))
	if err != nil {
		return code.Error(fmt.Errorf("failed to save creator: %v", err))
	}

	return code.OK([]byte("Initialized successfully by " + args.Creator))
}

// RegisterCopyright 注册新的数字版权信息
// args: {"copyright_id": "...", "work_name": "...", "author": "...", "owner_address": "...", "work_type": "...", "description": "..."}
func (cm *CopyrightManager) RegisterCopyright(ctx code.Context) code.Response {
	callArgs := ctx.Args()
	var newCopyright DigitalCopyright

	// 从参数中提取字段
	// 注意：ctx.Args() 返回 map[string][]byte，需要进行类型转换
	cidBytes, ok := callArgs["copyright_id"]
	if !ok {
		return code.Errors("missing copyright_id")
	}
	newCopyright.CopyrightID = string(cidBytes)

	wnBytes, ok := callArgs["work_name"]
	if !ok {
		return code.Errors("missing work_name")
	}
	newCopyright.WorkName = string(wnBytes)

	authorBytes, ok := callArgs["author"]
	if !ok {
		return code.Errors("missing author")
	}
	newCopyright.Author = string(authorBytes)

	ownerBytes, ok := callArgs["owner_address"]
	if !ok {
		return code.Errors("missing owner_address")
	}
	newCopyright.OwnerAddress = string(ownerBytes)

	wtBytes, ok := callArgs["work_type"]
	if !ok {
		return code.Errors("missing work_type")
	}
	newCopyright.WorkType = string(wtBytes)

	descBytes, ok := callArgs["description"]
	if !ok {
		return code.Errors("missing description")
	}
	newCopyright.Description = string(descBytes)

	// 参数校验
	if newCopyright.CopyrightID == "" || newCopyright.WorkName == "" || newCopyright.Author == "" || newCopyright.OwnerAddress == "" || newCopyright.WorkType == "" {
		return code.Errors("missing required fields for copyright registration")
	}

	// 检查版权ID是否已存在
	key := []byte(newCopyright.CopyrightID)
	existingData, err := ctx.GetObject(key)
	if err == nil && existingData != nil { // GetObject 成功且数据非空，说明已存在
		return code.Errors(fmt.Sprintf("copyright ID %s already exists", newCopyright.CopyrightID))
	}
	// 如果 err != nil, 可能是 key not found, 这是我们期望的

	// 设置登记日期和初始交易历史
	newCopyright.RegistrationDate = time.Now().Format("2006-01-02 15:04:05")
	newCopyright.TransactionHistory = []string{fmt.Sprintf("Registered on %s by %s", newCopyright.RegistrationDate, newCopyright.OwnerAddress)}
	newCopyright.IsDeleted = false

	copyrightJSON, err := json.Marshal(newCopyright)
	if err != nil {
		return code.Error(fmt.Errorf("failed to marshal copyright data: %v", err))
	}

	err = ctx.PutObject(key, copyrightJSON)
	if err != nil {
		return code.Error(fmt.Errorf("failed to save copyright data: %v", err))
	}

	// 触发事件 (可选)
	// 注意：EmitJSONEvent 的第二个参数在某些SDK版本中可能是 map[string][]byte 或 map[string]interface{}
	// 这里我们用 map[string]interface{} 尝试，如果编译不通过，可能需要调整为 map[string][]byte
	eventData := map[string]interface{}{
		"copyrightId": newCopyright.CopyrightID,
		"workName":    newCopyright.WorkName,
		"owner":       newCopyright.OwnerAddress,
	}
	ctx.EmitJSONEvent("CopyrightRegistered", eventData)

	return code.OK([]byte(fmt.Sprintf("Copyright %s registered successfully", newCopyright.CopyrightID)))
}

// QueryCopyright 根据版权ID查询版权详情
// args: {"copyright_id": "..."}
func (cm *CopyrightManager) QueryCopyright(ctx code.Context) code.Response {
	argsMap := ctx.Args()
	copyrightIDBytes, ok := argsMap["copyright_id"]
	if !ok {
		return code.Errors("missing copyright_id argument")
	}
	copyrightID := string(copyrightIDBytes)
	if copyrightID == "" {
		return code.Errors("copyright_id cannot be empty")
	}

	key := []byte(copyrightID)
	copyrightJSON, err := ctx.GetObject(key)
	if err != nil { // GetObject 在 key 不存在时也会返回 error
		return code.Error(fmt.Errorf("failed to get copyright %s: %v (it may not exist)", copyrightID, err))
	}
	if copyrightJSON == nil { // 理论上 err != nil 时 copyrightJSON 应该也是 nil
		return code.Errors(fmt.Sprintf("copyright %s not found", copyrightID))
	}

	var copyrightData DigitalCopyright
	if err := json.Unmarshal(copyrightJSON, &copyrightData); err != nil {
		return code.Error(fmt.Errorf("failed to unmarshal copyright data for check: %v", err))
	}
	if copyrightData.IsDeleted {
		return code.Error(fmt.Errorf("copyright %s has been deleted", copyrightID))
	}

	return code.OK(copyrightJSON)
}

// TransferCopyright 转移版权给新的持有人
// args: {"copyright_id": "...", "new_owner_address": "..."}
func (cm *CopyrightManager) TransferCopyright(ctx code.Context) code.Response {
	argsMap := ctx.Args()
	copyrightIDBytes, okCid := argsMap["copyright_id"]
	newOwnerAddressBytes, okOwner := argsMap["new_owner_address"]

	if !okCid || !okOwner {
		return code.Errors("missing copyright_id or new_owner_address")
	}
	copyrightID := string(copyrightIDBytes)
	newOwnerAddress := string(newOwnerAddressBytes)

	if copyrightID == "" || newOwnerAddress == "" {
		return code.Errors("copyright_id or new_owner_address cannot be empty")
	}

	key := []byte(copyrightID)
	copyrightJSON, err := ctx.GetObject(key)
	if err != nil {
		return code.Error(fmt.Errorf("failed to get copyright %s: %v", copyrightID, err))
	}
	if copyrightJSON == nil {
		return code.Errors(fmt.Sprintf("copyright %s not found", copyrightID))
	}

	var copyrightData DigitalCopyright
	if err := json.Unmarshal(copyrightJSON, &copyrightData); err != nil {
		return code.Error(fmt.Errorf("failed to unmarshal copyright data: %v", err))
	}

	if copyrightData.IsDeleted {
		return code.Error(fmt.Errorf("cannot transfer a deleted copyright: %s", copyrightID))
	}

	// 权限验证：例如，只有当前拥有者才能转移
	// initiator := string(ctx.Initiator()) // 获取调用者地址
	// if initiator != copyrightData.OwnerAddress {
	// return code.Error(fmt.Errorf("only current owner %s can transfer, caller is %s", copyrightData.OwnerAddress, initiator))
	// }
	// 注意：ctx.Initiator() 返回 []byte，需要转为 string。权限控制需仔细设计。

	oldOwner := copyrightData.OwnerAddress
	copyrightData.OwnerAddress = newOwnerAddress
	transferRecord := fmt.Sprintf("Transferred from %s to %s on %s", oldOwner, newOwnerAddress, time.Now().Format("2006-01-02 15:04:05"))
	copyrightData.TransactionHistory = append(copyrightData.TransactionHistory, transferRecord)

	updatedCopyrightJSON, err := json.Marshal(copyrightData)
	if err != nil {
		return code.Error(fmt.Errorf("failed to marshal updated copyright data: %v", err))
	}

	err = ctx.PutObject(key, updatedCopyrightJSON)
	if err != nil {
		return code.Error(fmt.Errorf("failed to save updated copyright data: %v", err))
	}

	eventData := map[string]interface{}{
		"copyrightId": copyrightID,
		"oldOwner":    oldOwner,
		"newOwner":    newOwnerAddress,
	}
	ctx.EmitJSONEvent("CopyrightTransferred", eventData)
	return code.OK([]byte(fmt.Sprintf("Copyright %s transferred to %s", copyrightID, newOwnerAddress)))
}

// UpdateCopyrightDescription 更新版权作品的描述
// args: {"copyright_id": "...", "new_description": "..."}
func (cm *CopyrightManager) UpdateCopyrightDescription(ctx code.Context) code.Response {
	argsMap := ctx.Args()
	copyrightIDBytes, okCid := argsMap["copyright_id"]
	newDescriptionBytes, okDesc := argsMap["new_description"]

	if !okCid {
		return code.Errors("missing copyright_id")
	}
	// 允许 new_description 为空字符串，所以不检查 okDesc 是否为 false

	copyrightID := string(copyrightIDBytes)
	newDescription := ""
	if okDesc { // 只有当 new_description 参数存在时才更新
		newDescription = string(newDescriptionBytes)
	}

	if copyrightID == "" {
		return code.Errors("copyright_id cannot be empty")
	}

	key := []byte(copyrightID)
	copyrightJSON, err := ctx.GetObject(key)
	if err != nil {
		return code.Error(fmt.Errorf("failed to get copyright %s: %v", copyrightID, err))
	}
	if copyrightJSON == nil {
		return code.Errors(fmt.Sprintf("copyright %s not found", copyrightID))
	}

	var copyrightData DigitalCopyright
	if err := json.Unmarshal(copyrightJSON, &copyrightData); err != nil {
		return code.Error(fmt.Errorf("failed to unmarshal copyright data: %v", err))
	}

	if copyrightData.IsDeleted {
		return code.Error(fmt.Errorf("cannot update description of a deleted copyright: %s", copyrightID))
	}

	oldDescription := copyrightData.Description
	copyrightData.Description = newDescription
	updateRecord := fmt.Sprintf("Description updated on %s. Old: '%s', New: '%s'", time.Now().Format("2006-01-02 15:04:05"), oldDescription, newDescription)
	copyrightData.TransactionHistory = append(copyrightData.TransactionHistory, updateRecord)

	updatedCopyrightJSON, err := json.Marshal(copyrightData)
	if err != nil {
		return code.Error(fmt.Errorf("failed to marshal updated copyright data: %v", err))
	}

	err = ctx.PutObject(key, updatedCopyrightJSON)
	if err != nil {
		return code.Error(fmt.Errorf("failed to save updated copyright data: %v", err))
	}

	eventData := map[string]interface{}{
		"copyrightId":    copyrightID,
		"newDescription": newDescription,
	}
	ctx.EmitJSONEvent("CopyrightDescriptionUpdated", eventData)
	return code.OK([]byte(fmt.Sprintf("Description for copyright %s updated", copyrightID)))
}

// DeleteCopyright 标记删除版权信息 (逻辑删除)
// args: {"copyright_id": "..."}
func (cm *CopyrightManager) DeleteCopyright(ctx code.Context) code.Response {
	argsMap := ctx.Args()
	copyrightIDBytes, ok := argsMap["copyright_id"]
	if !ok {
		return code.Errors("missing copyright_id argument")
	}
	copyrightID := string(copyrightIDBytes)
	if copyrightID == "" {
		return code.Errors("copyright_id cannot be empty")
	}

	key := []byte(copyrightID)
	copyrightJSON, err := ctx.GetObject(key)
	if err != nil {
		return code.Error(fmt.Errorf("failed to get copyright %s: %v", copyrightID, err))
	}
	if copyrightJSON == nil {
		return code.Errors(fmt.Sprintf("copyright %s not found", copyrightID))
	}

	var copyrightData DigitalCopyright
	if err := json.Unmarshal(copyrightJSON, &copyrightData); err != nil {
		return code.Error(fmt.Errorf("failed to unmarshal copyright data: %v", err))
	}

	if copyrightData.IsDeleted {
		return code.Errors(fmt.Sprintf("copyright %s is already deleted", copyrightID))
	}

	// 权限校验 (简化)
	// initiator := string(ctx.Initiator())
	// contractCreatorBytes, _ := ctx.GetObject([]byte("contract_creator"))
	// contractCreator := string(contractCreatorBytes)
	// if initiator != copyrightData.OwnerAddress && initiator != contractCreator {
	//  return code.Error(fmt.Errorf("permission denied to delete copyright %s", copyrightID))
	// }

	copyrightData.IsDeleted = true
	deleteRecord := fmt.Sprintf("Copyright marked as deleted on %s by caller %s", time.Now().Format("2006-01-02 15:04:05"), string(ctx.Initiator()))
	copyrightData.TransactionHistory = append(copyrightData.TransactionHistory, deleteRecord)

	updatedCopyrightJSON, err := json.Marshal(copyrightData)
	if err != nil {
		return code.Error(fmt.Errorf("failed to marshal updated copyright data for deletion: %v", err))
	}

	err = ctx.PutObject(key, updatedCopyrightJSON)
	if err != nil {
		return code.Error(fmt.Errorf("failed to save updated copyright data for deletion: %v", err))
	}

	eventData := map[string]interface{}{
		"copyrightId": copyrightID,
		"deletedBy":   string(ctx.Initiator()),
	}
	ctx.EmitJSONEvent("CopyrightDeleted", eventData)
	return code.OK([]byte(fmt.Sprintf("Copyright %s marked as deleted", copyrightID)))
}

func main() {
	// 关键改动：使用 driver.Serve
	driver.Serve(new(CopyrightManager))
}
