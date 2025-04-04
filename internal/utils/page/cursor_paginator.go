package page

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type CursorPageBaseRequest struct {
	Cursor   string `json:"cursor"`
	PageSize int    `json:"pageSize"`
}

type CursorPageBaseVO[T any] struct {
	Cursor string `json:"cursor"`
	IsLast bool   `json:"isLast"`
	Data   []T    `json:"data"`
}

func GetCursorPageByMySQL[T any](db *gorm.DB, request CursorPageBaseRequest,
	initWrapper func(*gorm.DB), cursorColumn func(*T) interface{}) (*CursorPageBaseVO[T], error) {

	t := new(T)
	column := cursorColumn(t)

	// 获取结构体的类型和值
	structValue := reflect.ValueOf(t).Elem()
	structType := structValue.Type()

	// 获取字段的类型和名称
	cursorFieldValue := reflect.ValueOf(column).Elem()
	cursorFieldType := cursorFieldValue.Type()

	// 定义一个递归函数来查找字段
	var fieldName string
	var found bool

	// 递归查找字段
	var findField func(reflect.Value, reflect.Type)
	findField = func(value reflect.Value, typ reflect.Type) {
		for i := 0; i < value.NumField(); i++ {
			fieldValue := value.Field(i)
			fieldType := typ.Field(i).Type

			// 如果字段是结构体，则递归进入
			if fieldValue.Kind() == reflect.Struct && typ.Field(i).Anonymous {
				findField(fieldValue, fieldType)
				if found {
					return
				}
			}

			// 比较字段的地址和类型
			if fieldValue.Addr().Pointer() == cursorFieldValue.Addr().Pointer() && fieldType == cursorFieldType {
				gormColumn := typ.Field(i).Tag.Get("gorm")
				// 从 GORM 标签中提取 column 部分
				// 找到 "column:" 的起始位置
				columnPrefix := "column:"
				start := strings.Index(gormColumn, columnPrefix)
				if start == -1 {
					fieldName = ""
					found = true
					return
				}
				// 从 "column:" 之后开始查找 ";"
				start += len(columnPrefix)
				end := strings.Index(gormColumn[start:], ";")
				if end == -1 {
					// 如果没有找到 ";"，直接返回 "column:" 之后的所有内容
					fieldName = gormColumn[start:]
					found = true
					return
				}
				// 返回 "column:" 至 ";" 之间的内容
				fieldName = gormColumn[start : start+end]
				found = true
				return
			}
		}
	}

	// 开始查找
	findField(structValue, structType)

	if !found {
		panic("Field not found in struct")
	}

	// 获取游标字段类型
	cursorType := cursorFieldType.Kind()

	// 初始化查询条件
	query := db.Model(t)
	if initWrapper != nil {
		initWrapper(query)
	}

	// 游标条件
	if request.Cursor != "" {
		cursorValue := parseCursor(request.Cursor, cursorType)
		query = query.Where(fmt.Sprintf("%v < ?", fieldName), cursorValue)
	}

	// 游标方向
	query = query.Order(fmt.Sprintf("%v DESC", fieldName))

	// 分页查询
	var results []T
	if err := query.Limit(request.PageSize).Find(&results).Error; err != nil {
		return nil, err
	}

	// 取出游标
	var cursor string
	if len(results) > 0 {
		lastRecord := results[len(results)-1]
		cursorValue := cursorColumn(&lastRecord)
		cursor = toCursor(cursorValue)
	}

	// 判断是否最后一页
	isLast := len(results) < request.PageSize

	return &CursorPageBaseVO[T]{
		Cursor: cursor,
		IsLast: isLast,
		Data:   results,
	}, nil
}

// 解析游标
func parseCursor(cursor string, cursorType reflect.Kind) interface{} {
	switch cursorType {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			return val
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val, err := strconv.ParseUint(cursor, 10, 64); err == nil {
			return val
		}
	case reflect.Float32, reflect.Float64:
		if val, err := strconv.ParseFloat(cursor, 64); err == nil {
			return val
		}
	case reflect.String:
		return cursor
	case reflect.Struct:
		if cursorType == reflect.TypeOf(time.Time{}).Kind() {
			location, _ := time.LoadLocation("Asia/Shanghai")
			if val, err := time.ParseInLocation("2006-01-02 15:04:05.000", cursor, location); err == nil {
				return val
			}
		}
	default:
		return nil
	}
	return nil
}

// 生成游标
func toCursor(value interface{}) string {
	cursorValue := reflect.ValueOf(value).Elem()
	cursorType := cursorValue.Type().Kind()
	switch cursorType {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(cursorValue.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(cursorValue.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(cursorValue.Float(), 'f', -1, 64)
	case reflect.String:
		return cursorValue.String()
	case reflect.Struct:
		// 处理 time.Time 类型
		if cursorValue.Type() == reflect.TypeOf(time.Time{}) {
			t := cursorValue.Interface().(time.Time)
			return t.Format("2006-01-02 15:04:05.000")
		}
		return ""
	default:
		return ""
	}
}

// ConvertCursorPageVO 将当前类型的分页数据转换为目标类型
func ConvertCursorPageVO[T any, D any](p *CursorPageBaseVO[T]) (*CursorPageBaseVO[D], error) {
	result := &CursorPageBaseVO[D]{
		Cursor: p.Cursor,
		IsLast: p.IsLast,
		Data:   make([]D, len(p.Data)),
	}

	err := copier.Copy(&result.Data, &p.Data)
	if err != nil {
		return nil, err
	}

	return result, nil
}
