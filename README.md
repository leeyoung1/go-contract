# 数字版权管理智能合约 (Digital Copyright Management Smart Contract)

这是一个基于 XuperChain 开发的数字版权管理智能合约，用于在区块链上注册、管理和转移数字版权。

## 功能特性

- **版权注册**: 在区块链上注册新的数字版权
- **版权查询**: 根据版权ID查询版权详细信息
- **版权转移**: 将版权转移给新的持有人
- **描述更新**: 更新版权作品的描述信息
- **逻辑删除**: 标记删除版权信息（软删除）
- **历史记录**: 完整的交易和变更历史追踪

## 技术栈

- **区块链平台**: XuperChain
- **开发语言**: Go
- **合约SDK**: github.com/xuperchain/contract-sdk-go

## 合约结构

### 数据结构

```go
type DigitalCopyright struct {
    CopyrightID        string   `json:"copyright_id"`        // 版权唯一ID
    WorkName           string   `json:"work_name"`           // 作品名称
    Author             string   `json:"author"`              // 作者名称
    OwnerAddress       string   `json:"owner_address"`       // 当前版权持有人地址
    RegistrationDate   string   `json:"registration_date"`   // 登记日期
    WorkType           string   `json:"work_type"`           // 作品类型
    Description        string   `json:"description"`         // 作品描述
    TransactionHistory []string `json:"transaction_history"` // 交易历史记录
    IsDeleted          bool     `json:"is_deleted"`          // 删除标记
}
```

### 主要方法

1. **Initialize**: 合约初始化
2. **RegisterCopyright**: 注册新版权
3. **QueryCopyright**: 查询版权信息
4. **TransferCopyright**: 转移版权所有权
5. **UpdateCopyrightDescription**: 更新版权描述
6. **DeleteCopyright**: 删除版权（逻辑删除）

## 编译和部署

### 编译合约

```bash
go build -o digital_copyright_service main.go
```

### 部署到 XuperChain

1. 确保 XuperChain 节点正在运行
2. 使用客户端工具部署合约
3. 调用 Initialize 方法初始化合约

## 使用示例

### 注册版权

```json
{
    "copyright_id": "CR001",
    "work_name": "我的数字作品",
    "author": "张三",
    "owner_address": "XC1234567890...",
    "work_type": "图片",
    "description": "这是一幅数字艺术作品"
}
```

### 查询版权

```json
{
    "copyright_id": "CR001"
}
```

### 转移版权

```json
{
    "copyright_id": "CR001",
    "new_owner_address": "XC0987654321..."
}
```

## 事件

合约会触发以下事件：
- `CopyrightRegistered`: 版权注册成功
- `CopyrightTransferred`: 版权转移完成
- `CopyrightDescriptionUpdated`: 描述更新完成
- `CopyrightDeleted`: 版权删除完成

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。
