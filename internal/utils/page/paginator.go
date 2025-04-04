package page

import "gorm.io/gorm"

type PageRequest struct {
	Page    int    `json:"page" form:"page"`         // 当前页码，默认从 1 开始
	Limit   int    `json:"limit" form:"limit"`       // 每页显示的条数
	OrderBy string `json:"order_by" form:"order_by"` // 排序字段
	// 排序方式，asc 升序，desc 降序，默认 desc
	Order string `json:"order" form:"order"`
}

type PageVo[T any] struct {
	Total   int64 `json:"total"`   // 总记录数
	Page    int   `json:"page"`    // 当前页码
	Limit   int   `json:"limit"`   // 每页记录数
	Records []T   `json:"records"` // 当前页记录，使用泛型 T 类型的切片
}

func Paginate[T any](db *gorm.DB, param PageRequest, result *[]T) (*PageVo[T], error) {
	// 计算偏移量
	offset := (param.Page - 1) * param.Limit

	// 统计总记录数
	var total int64
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, err
	}

	orderStr := ""
	if param.OrderBy != "" {
		orderStr = param.OrderBy + " "
		if param.Order == "asc" {
			orderStr += "asc"
		} else {
			orderStr += "desc"
		}
	}

	if orderStr != "" {
		db = db.Order(orderStr)
	}

	// 查询当前页数据
	if err := db.Limit(param.Limit).Offset(offset).Find(result).Error; err != nil {
		return nil, err
	}

	// 组装分页结果
	pageResult := &PageVo[T]{
		Total:   total,
		Page:    param.Page,
		Limit:   param.Limit,
		Records: *result,
	}

	return pageResult, nil
}

// Convert 将一个类型的 PageVo 转换为另一个类型的 PageVo
// From: 原始类型
// To: 目标类型
// records: 新的记录数据
func Convert[From any, To any](original *PageVo[From], records []To) *PageVo[To] {
	return &PageVo[To]{
		Total:   original.Total,
		Page:    original.Page,
		Limit:   original.Limit,
		Records: records,
	}
}
