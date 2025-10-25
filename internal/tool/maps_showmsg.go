package tool

import (
	"encoding/json"
	"strconv"
	"strings"
	"txing-ai/internal/global/logging/log"

	"go.uber.org/zap"
	"html"
)

// 兼容字符串或空数组的字符串类型（如 rating 可能为 "4.8" 或 []）
type FlexibleString struct {
	value string
}

func (fs *FlexibleString) UnmarshalJSON(b []byte) error {
	// 优先按字符串解析
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		fs.value = s
		return nil
	}
	// 如果是数组（如 []），如果不为空，则取第一个元素，否则视为为空字符串
	var arr []any
	if err := json.Unmarshal(b, &arr); err == nil {
		fs.value = ""
		if len(arr) > 0 {
			var al []string
			for _, a := range arr {
				as := strings.TrimSpace(a.(string))
				if as != "" {
					al = append(al, as)
				}
			}
			if len(al) > 0 {
				fs.value = strings.Join(al, "、")
			}
		}
		return nil
	}
	// 如果是 null
	var v interface{}
	if err := json.Unmarshal(b, &v); err == nil {
		if v == nil {
			fs.value = ""
			return nil
		}
	}
	// 其他类型（数字/对象等）统一视为空
	fs.value = ""
	return nil
}

func (fs FlexibleString) String() string { return fs.value }

// 地图地理工具展示构造
// 请求示例: {"address":"广州塔","city":"广州"}
// 响应为 MCP 包裹或直接 JSON：{"return":[{...}]}
type mapsGeoShowBuilder struct{}

func (mapsGeoShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params struct {
		Address string `json:"address"`
		City    string `json:"city"`
	}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建地图地理请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	addr := strings.TrimSpace(params.Address)
	city := strings.TrimSpace(params.City)
	if city != "" {
		return "地理位置查询：" + addr + "（" + city + "）", nil
	}
	return "地理位置查询：" + addr, nil
}

func (mapsGeoShowBuilder) BuildResponse(response string) (string, error) {
	r := strings.TrimSpace(response)
	if r == "" {
		return "未解析到地理位置", nil
	}
	// 尝试解析 MCP 包裹的响应 {"content":[{"type":"text","text":"{...}"}]}
	var mcp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	inner := ""
	if err := json.Unmarshal([]byte(r), &mcp); err == nil && len(mcp.Content) > 0 {
		for _, c := range mcp.Content {
			if strings.TrimSpace(c.Text) != "" {
				inner = c.Text
				break
			}
		}
	}
	if inner == "" {
		inner = r
	}
	inner = strings.TrimSpace(inner)
	inner = strings.TrimSuffix(inner, ",")

	type mapsGeoItem struct {
		Country  string         `json:"country"`
		Province string         `json:"province"`
		City     FlexibleString `json:"city"`
		Citycode string         `json:"citycode"`
		District FlexibleString `json:"district"`
		Adcode   string         `json:"adcode"`
		Location string         `json:"location"`
		Level    string         `json:"level"`
	}
	type mapsGeoResp struct {
		Return []mapsGeoItem `json:"return"`
	}
	var mg mapsGeoResp
	if err := json.Unmarshal([]byte(inner), &mg); err != nil {
		log.Error("构建地图地理响应显示信息失败, 无法解析JSON", zap.Error(err))
		return "", ErrInvalidJSON
	}
	count := len(mg.Return)
	if count == 0 {
		return "未解析到地理位置", nil
	}

	var msgBuilder strings.Builder
	msgBuilder.WriteString("共解析出 " + strconv.Itoa(count) + " 个地理位置：\n")

	for i, item := range mg.Return {
		addrParts := []string{}
		if strings.TrimSpace(item.Province) != "" {
			addrParts = append(addrParts, strings.TrimSpace(item.Province))
		}
		if strings.TrimSpace(item.City.String()) != "" {
			addrParts = append(addrParts, strings.TrimSpace(item.City.String()))
		}
		if strings.TrimSpace(item.District.String()) != "" {
			addrParts = append(addrParts, strings.TrimSpace(item.District.String()))
		}
		addr := strings.Join(addrParts, " ")
		if addr == "" {
			addr = strings.TrimSpace(item.Country)
		}
		level := strings.TrimSpace(item.Level)
		loc := strings.TrimSpace(item.Location)

		msgBuilder.WriteString(strconv.Itoa(i+1) + ". " + addr)
		if loc != "" {
			msgBuilder.WriteString("，坐标 " + loc)
		}
		if level != "" {
			msgBuilder.WriteString("（" + level + "）")
		}
		if i < count-1 {
			msgBuilder.WriteString("\n")
		}
	}

	return msgBuilder.String(), nil
}

// 文本地点搜索工具展示构造
// 请求示例: {"keywords":"广州塔","city":"广州"}
// 响应为 MCP 包裹或直接 JSON：{"suggestion":{...},"pois":[{...}]}
type mapsTextSearchShowBuilder struct{}

func (mapsTextSearchShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params struct {
		Keywords string `json:"keywords"`
		City     string `json:"city"`
	}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建地图文本搜索请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	kw := strings.TrimSpace(params.Keywords)
	city := strings.TrimSpace(params.City)
	if kw == "" {
		kw = "(未提供关键词)"
	}
	if city != "" {
		return "地点文本搜索：" + kw + "（" + city + "）", nil
	}
	return "地点文本搜索：" + kw, nil
}

func (mapsTextSearchShowBuilder) BuildResponse(response string) (string, error) {
	r := strings.TrimSpace(response)
	if r == "" {
		return "未找到相关地点", nil
	}
	// 解析 MCP 包裹的响应 {"content":[{"type":"text","text":"{...}"}]}
	var mcp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	inner := ""
	if err := json.Unmarshal([]byte(r), &mcp); err == nil && len(mcp.Content) > 0 {
		for _, c := range mcp.Content {
			if strings.TrimSpace(c.Text) != "" {
				inner = c.Text
				break
			}
		}
	}
	if inner == "" {
		inner = r
	}
	inner = strings.TrimSpace(inner)
	inner = strings.TrimSuffix(inner, ",")

	type photoInfo struct {
		Url string `json:"url"`
	}
	type poiItem struct {
		Id       string         `json:"id"`
		Name     string         `json:"name"`
		Address  FlexibleString `json:"address"`
		Typecode string         `json:"typecode"`
		Location string         `json:"location"`
		Photos   *photoInfo     `json:"photos"`
	}
	type textSearchResp struct {
		Suggestion struct {
			Keywords []string `json:"keywords"`
			Ciytes   []string `json:"ciytes"`
			Cities   []string `json:"cities"`
		} `json:"suggestion"`
		Pois []poiItem `json:"pois"`
	}
	var ts textSearchResp
	if err := json.Unmarshal([]byte(inner), &ts); err != nil {
		log.Error("构建地图文本搜索响应显示信息失败, 无法解析JSON", zap.Error(err))
		return "", ErrInvalidJSON
	}

	count := len(ts.Pois)
	if count == 0 {
		return "未找到相关地点", nil
	}

	var msgBuilder strings.Builder
	msgBuilder.WriteString("共找到 " + strconv.Itoa(count) + " 个地点：<br/>")

	for i, p := range ts.Pois {
		name := strings.TrimSpace(p.Name)
		addr := strings.TrimSpace(p.Address.String())
		typecode := strings.TrimSpace(p.Typecode)
		loc := strings.TrimSpace(p.Location)
		url := ""
		if p.Photos != nil {
			u := strings.TrimSpace(p.Photos.Url)
			if u != "" {
				// 去除可能包裹的反引号和引号
				u = strings.Trim(u, " `\"")
				url = u
			}
		}

		msgBuilder.WriteString(strconv.Itoa(i+1) + ". ")
		if name != "" {
			msgBuilder.WriteString(name)
		}
		if addr != "" {
			msgBuilder.WriteString("，地址 " + addr)
		}
		if loc != "" {
			msgBuilder.WriteString("，坐标 " + loc)
		}
		if typecode != "" {
			msgBuilder.WriteString("，类型 " + typecode)
		}
		if url != "" {
			msgBuilder.WriteString("<br/><img src=\"" + html.EscapeString(url) + "\" alt=\"" + html.EscapeString(name) + "\" style=\"max-width:120px;max-height:100px;object-fit:cover;border-radius:6px;margin-left:6px;\"/>")
		}
		if i < count-1 {
			msgBuilder.WriteString("<br/>")
		}
	}

	return msgBuilder.String(), nil
}

// 公交换乘一体化展示构造
// 请求示例: {"city":"广州","cityd":"广州","origin":"113.264499,23.130061","destination":"113.264499,23.130061"}
// 响应为 MCP 包裹或直接 JSON：{"route":{...}}
type mapsTransitIntegratedShowBuilder struct{}

func (mapsTransitIntegratedShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params struct {
		City        string `json:"city"`
		Cityd       string `json:"cityd"`
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
	}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建公交换乘请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	city := strings.TrimSpace(params.City)
	cityd := strings.TrimSpace(params.Cityd)
	origin := strings.TrimSpace(params.Origin)
	destination := strings.TrimSpace(params.Destination)

	var b strings.Builder
	b.WriteString("公交换乘规划：")
	if origin != "" && destination != "" {
		b.WriteString(origin + " → " + destination)
	} else if origin != "" {
		b.WriteString("从 " + origin)
	} else if destination != "" {
		b.WriteString("到 " + destination)
	} else {
		b.WriteString("(未提供坐标)")
	}
	if city != "" || cityd != "" {
		b.WriteString("（")
		if city != "" && cityd != "" {
			b.WriteString(city + " → " + cityd)
		} else if city != "" {
			b.WriteString(city)
		} else {
			b.WriteString(cityd)
		}
		b.WriteString("）")
	}
	return b.String(), nil
}

func (mapsTransitIntegratedShowBuilder) BuildResponse(response string) (string, error) {
	r := strings.TrimSpace(response)
	if r == "" {
		return "未解析到公交换乘方案", nil
	}
	var mcp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	inner := ""
	if err := json.Unmarshal([]byte(r), &mcp); err == nil && len(mcp.Content) > 0 {
		for _, c := range mcp.Content {
			if strings.TrimSpace(c.Text) != "" {
				inner = c.Text
				break
			}
		}
	}
	if inner == "" {
		inner = r
	}
	inner = strings.TrimSpace(inner)
	inner = strings.TrimSuffix(inner, ",")

	// 数据结构定义（关注必要字段）
	type WalkingStep struct {
		Instruction     string      `json:"instruction"`
		Road            interface{} `json:"road"`
		Distance        string      `json:"distance"`
		Action          interface{} `json:"action"`
		AssistantAction interface{} `json:"assistant_action"`
	}
	type Walking struct {
		Origin      string        `json:"origin"`
		Destination string        `json:"destination"`
		Distance    string        `json:"distance"`
		Duration    string        `json:"duration"`
		Steps       []WalkingStep `json:"steps"`
	}
	type Stop struct {
		Name string `json:"name"`
	}
	type Busline struct {
		Name          string `json:"name"`
		DepartureStop Stop   `json:"departure_stop"`
		ArrivalStop   Stop   `json:"arrival_stop"`
		Distance      string `json:"distance"`
		Duration      string `json:"duration"`
		ViaStops      []Stop `json:"via_stops"`
	}
	type Bus struct {
		Buslines []Busline `json:"buslines"`
	}
	type Named struct {
		Name string `json:"name"`
	}
	type Segment struct {
		Walking  *Walking    `json:"walking"`
		Bus      *Bus        `json:"bus"`
		Entrance *Named      `json:"entrance"`
		Exit     *Named      `json:"exit"`
		Railway  interface{} `json:"railway"`
	}
	type Transit struct {
		Duration        string    `json:"duration"`
		WalkingDistance string    `json:"walking_distance"`
		Segments        []Segment `json:"segments"`
	}
	type Route struct {
		Origin      string    `json:"origin"`
		Destination string    `json:"destination"`
		Distance    string    `json:"distance"`
		Transits    []Transit `json:"transits"`
	}
	type TransitResp struct {
		Route Route `json:"route"`
	}

	var tr TransitResp
	if err := json.Unmarshal([]byte(inner), &tr); err != nil {
		log.Error("构建公交换乘响应显示信息失败, 无法解析JSON", zap.Error(err))
		return "", ErrInvalidJSON
	}

	if len(tr.Route.Transits) == 0 {
		return "未解析到公交换乘方案", nil
	}

	var msgBuilder strings.Builder
	origin := strings.TrimSpace(tr.Route.Origin)
	destination := strings.TrimSpace(tr.Route.Destination)
	distance := strings.TrimSpace(tr.Route.Distance)
	if origin != "" || destination != "" || distance != "" {
		msgBuilder.WriteString("换乘方案（起点 ")
		if origin != "" {
			msgBuilder.WriteString(origin)
		} else {
			msgBuilder.WriteString("未知")
		}
		msgBuilder.WriteString(" → 终点 ")
		if destination != "" {
			msgBuilder.WriteString(destination)
		} else {
			msgBuilder.WriteString("未知")
		}
		if distance != "" {
			msgBuilder.WriteString("，总距离 " + distance + " 米")
		}
		msgBuilder.WriteString("）：\n")
	}

	for i, t := range tr.Route.Transits {
		msgBuilder.WriteString("方案 " + strconv.Itoa(i+1) + "：总时长 " + strings.TrimSpace(t.Duration))
		if strings.TrimSpace(t.WalkingDistance) != "" {
			msgBuilder.WriteString(" 秒，步行 " + strings.TrimSpace(t.WalkingDistance) + " 米")
		}
		msgBuilder.WriteString("\n")
		for _, seg := range t.Segments {
			if seg.Entrance != nil && strings.TrimSpace(seg.Entrance.Name) != "" {
				msgBuilder.WriteString("  进站：" + strings.TrimSpace(seg.Entrance.Name) + "\n")
			}
			if seg.Walking != nil {
				w := seg.Walking
				// 如果没有步行步骤但有距离/时长也展示
				if strings.TrimSpace(w.Distance) != "" || strings.TrimSpace(w.Duration) != "" {
					msgBuilder.WriteString("  步行：")
					if strings.TrimSpace(w.Distance) != "" {
						msgBuilder.WriteString(strings.TrimSpace(w.Distance) + " 米")
					}
					if strings.TrimSpace(w.Duration) != "" {
						if strings.TrimSpace(w.Distance) != "" {
							msgBuilder.WriteString("，")
						}
						msgBuilder.WriteString(strings.TrimSpace(w.Duration) + " 秒")
					}
					msgBuilder.WriteString("\n")
				}
				for _, step := range w.Steps {
					ins := strings.TrimSpace(step.Instruction)
					dist := strings.TrimSpace(step.Distance)
					if ins != "" || dist != "" {
						msgBuilder.WriteString("    - " + ins)
						if dist != "" {
							msgBuilder.WriteString("（" + dist + " 米）")
						}
						msgBuilder.WriteString("\n")
					}
				}
			}
			if seg.Bus != nil {
				for _, bl := range seg.Bus.Buslines {
					name := strings.TrimSpace(bl.Name)
					dep := strings.TrimSpace(bl.DepartureStop.Name)
					arr := strings.TrimSpace(bl.ArrivalStop.Name)
					bdist := strings.TrimSpace(bl.Distance)
					bdur := strings.TrimSpace(bl.Duration)
					var via []string
					for _, vs := range bl.ViaStops {
						s := strings.TrimSpace(vs.Name)
						if s != "" {
							via = append(via, s)
						}
					}
					msgBuilder.WriteString("  乘坐：")
					if name != "" {
						msgBuilder.WriteString(name)
					} else {
						msgBuilder.WriteString("公交/地铁")
					}
					if dep != "" || arr != "" {
						msgBuilder.WriteString("（")
						if dep != "" {
							msgBuilder.WriteString("从 " + dep)
						}
						if arr != "" {
							if dep != "" {
								msgBuilder.WriteString(" 到 ")
							} else {
								msgBuilder.WriteString("到 ")
							}
							msgBuilder.WriteString(arr)
						}
						msgBuilder.WriteString("）")
					}
					if len(via) > 0 {
						msgBuilder.WriteString("，途经 " + strings.Join(via, "、"))
					}
					if bdist != "" {
						msgBuilder.WriteString("，距离 " + bdist + " 米")
					}
					if bdur != "" {
						msgBuilder.WriteString("，时长 " + bdur + " 秒")
					}
					msgBuilder.WriteString("\n")
				}
			}
			if seg.Exit != nil && strings.TrimSpace(seg.Exit.Name) != "" {
				msgBuilder.WriteString("  出站：" + strings.TrimSpace(seg.Exit.Name) + "\n")
			}
		}
		if i < len(tr.Route.Transits)-1 {
			msgBuilder.WriteString("\n")
		}
	}

	return msgBuilder.String(), nil
}

// 新增：距离测算展示构造器
type mapsDistanceShowBuilder struct{}

func (mapsDistanceShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params struct {
		Origins     string `json:"origins"`
		Destination string `json:"destination"`
		Type        string `json:"type"`
	}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建距离测算请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	origins := strings.TrimSpace(params.Origins)
	destination := strings.TrimSpace(params.Destination)
	t := strings.TrimSpace(params.Type)
	var typeName string
	switch t {
	case "1":
		typeName = "驾车"
	case "2":
		typeName = "步行"
	case "3":
		typeName = "骑行"
	default:
		if t != "" {
			typeName = "类型 " + t
		}
	}
	var b strings.Builder
	b.WriteString("距离测算：")
	if origins != "" && destination != "" {
		parts := strings.Split(origins, "|")
		first := strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			b.WriteString(first + " 等 " + strconv.Itoa(len(parts)) + " 个起点 → " + destination)
		} else {
			b.WriteString(first + " → " + destination)
		}
	} else if origins != "" {
		b.WriteString("从 " + origins)
	} else if destination != "" {
		b.WriteString("到 " + destination)
	} else {
		b.WriteString("(未提供坐标)")
	}
	if typeName != "" {
		b.WriteString("（方式 " + typeName + "）")
	}
	return b.String(), nil
}

func (mapsDistanceShowBuilder) BuildResponse(response string) (string, error) {
	r := strings.TrimSpace(response)
	if r == "" {
		return "未解析到距离信息", nil
	}
	var mcp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	inner := ""
	if err := json.Unmarshal([]byte(r), &mcp); err == nil && len(mcp.Content) > 0 {
		for _, c := range mcp.Content {
			if strings.TrimSpace(c.Text) != "" {
				inner = c.Text
				break
			}
		}
	}
	if inner == "" {
		inner = r
	}
	inner = strings.TrimSpace(inner)
	inner = strings.TrimSuffix(inner, ",")

	type distanceItem struct {
		OriginID string `json:"origin_id"`
		DestID   string `json:"dest_id"`
		Distance string `json:"distance"`
		Duration string `json:"duration"`
	}
	var dr struct {
		Results []distanceItem `json:"results"`
	}
	if err := json.Unmarshal([]byte(inner), &dr); err != nil {
		log.Error("构建距离测算响应显示信息失败, 无法解析JSON", zap.Error(err))
		return "", ErrInvalidJSON
	}
	if len(dr.Results) == 0 {
		return "未解析到距离信息", nil
	}
	var b strings.Builder
	b.WriteString("共计算 " + strconv.Itoa(len(dr.Results)) + " 条结果：\n")
	for i, it := range dr.Results {
		origin := strings.TrimSpace(it.OriginID)
		dest := strings.TrimSpace(it.DestID)
		dist := strings.TrimSpace(it.Distance)
		dur := strings.TrimSpace(it.Duration)
		if origin == "" {
			origin = "?"
		}
		if dest == "" {
			dest = "?"
		}
		b.WriteString(strconv.Itoa(i+1) + ". 起点 " + origin + " → 终点 " + dest)
		if dist != "" {
			b.WriteString("：距离 " + dist + " 米")
		}
		if dur != "" {
			b.WriteString("，时长 " + dur + " 秒")
		}
		if i < len(dr.Results)-1 {
			b.WriteString("\n")
		}
	}
	return b.String(), nil
}

// 地点详情展示构造器
// 请求示例: {"id":"B0FFHOSV0F"}
// 响应为 MCP 包裹或直接 JSON：{"id":"...","name":"...",...}
type mapsSearchDetailShowBuilder struct{}

func (mapsSearchDetailShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建地点详情请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	id := strings.TrimSpace(params.ID)
	if id == "" {
		return "地点详情查询：(未提供ID)", nil
	}
	return "地点详情查询：" + id, nil
}

func (mapsSearchDetailShowBuilder) BuildResponse(response string) (string, error) {
	r := strings.TrimSpace(response)
	if r == "" {
		return "未解析到地点详情", nil
	}
	// 尝试解析 MCP 包裹的响应 {"content":[{"type":"text","text":"{...}"}]}
	var mcp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	inner := ""
	if err := json.Unmarshal([]byte(r), &mcp); err == nil && len(mcp.Content) > 0 {
		for _, c := range mcp.Content {
			if strings.TrimSpace(c.Text) != "" {
				inner = c.Text
				break
			}
		}
	}
	if inner == "" {
		inner = r
	}
	inner = strings.TrimSpace(inner)
	inner = strings.TrimSuffix(inner, ",")

	// 解析地点详情主体
	type photo struct {
		Url string `json:"url"`
	}
	var detail struct {
		ID           string         `json:"id"`
		Name         string         `json:"name"`
		Location     string         `json:"location"`
		Address      string         `json:"address"`
		BusinessArea FlexibleString `json:"business_area"`
		City         string         `json:"city"`
		Type         string         `json:"type"`
		Alias        FlexibleString `json:"alias"`
		Photos       *photo         `json:"photos"`
		Rating       FlexibleString `json:"rating"`
		OpenTime     FlexibleString `json:"open_time"`
		Opentime2    FlexibleString `json:"opentime2"`
		TicketOrder  string         `json:"ticket_ordering"`
	}
	if err := json.Unmarshal([]byte(inner), &detail); err != nil {
		log.Error("构建地点详情响应显示信息失败, 无法解析JSON", zap.Error(err))
		return "", ErrInvalidJSON
	}

	name := strings.TrimSpace(detail.Name)
	if name == "" && strings.TrimSpace(detail.ID) == "" {
		return "未解析到地点详情", nil
	}

	address := strings.TrimSpace(detail.Address)
	city := strings.TrimSpace(detail.City)
	ba := strings.TrimSpace(detail.BusinessArea.String())
	loc := strings.TrimSpace(detail.Location)
	typeStr := strings.TrimSpace(detail.Type)
	rating := strings.TrimSpace(detail.Rating.String())
	open := strings.TrimSpace(detail.OpenTime.String())
	if open == "" {
		open = strings.TrimSpace(detail.Opentime2.String())
	}
	order := strings.TrimSpace(detail.TicketOrder)
	var orderStr string
	switch order {
	case "1":
		orderStr = "可订票"
	case "0":
		orderStr = "不可订票"
	default:
		orderStr = ""
	}

	// 处理图片 URL
	url := ""
	if detail.Photos != nil {
		u := strings.TrimSpace(detail.Photos.Url)
		if u != "" {
			u = strings.Trim(u, " `\"")
			url = u
		}
	}

	// 组装输出
	lines := []string{}
	if name != "" {
		lines = append(lines, "名称："+name)
	}
	if address != "" || city != "" {
		if city != "" && address != "" {
			lines = append(lines, "地址："+city+" - "+address)
		} else if address != "" {
			lines = append(lines, "地址："+address)
		} else {
			lines = append(lines, "城市："+city)
		}
	}
	if ba != "" {
		lines = append(lines, "商圈："+ba)
	}
	if loc != "" {
		lines = append(lines, "坐标："+loc)
	}
	if typeStr != "" {
		lines = append(lines, "类型："+typeStr)
	}
	if rating != "" {
		lines = append(lines, "评分："+rating)
	}
	if open != "" {
		lines = append(lines, "营业时间："+open)
	}
	if orderStr != "" {
		lines = append(lines, "订票："+orderStr)
	}
	if detail.Alias.value != "" {
		lines = append(lines, "别名："+detail.Alias.String())
	}

	if url != "" {
		// 使用 HTML 渲染以便图片展示与换行
		var b strings.Builder
		b.WriteString(strings.Join(lines, "<br/>"))
		b.WriteString("<br/>")
		b.WriteString("<img src=\"" + html.EscapeString(url) + "\" alt=\"" + html.EscapeString(name) + "\" style=\"max-width:260px;max-height:180px;object-fit:cover;border-radius:6px;margin-top:6px;\"/>")
		return b.String(), nil
	}
	// 无图片则使用纯文本，便于前端以 <pre> 展示
	return strings.Join(lines, "\n"), nil
}

// 天气预报工具消息展示构造器
type mapsWeatherShowBuilder struct{}

func (mapsWeatherShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params struct {
		City string `json:"city"`
	}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建天气请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	city := strings.TrimSpace(params.City)
	if city == "" {
		city = "(未提供城市)"
	}
	return "天气预报查询：" + city, nil
}

func weekToCN(week string) string {
	switch strings.TrimSpace(week) {
	case "1", "一":
		return "周一"
	case "2", "二":
		return "周二"
	case "3", "三":
		return "周三"
	case "4", "四":
		return "周四"
	case "5", "五":
		return "周五"
	case "6", "六":
		return "周六"
	case "7", "日", "天":
		return "周日"
	default:
		return ""
	}
}

func (mapsWeatherShowBuilder) BuildResponse(response string) (string, error) {
	r := strings.TrimSpace(response)
	if r == "" {
		return "未解析到天气信息", nil
	}
	// 解析 MCP 包裹的响应 {"content":[{"type":"text","text":"{...}"}]}
	var mcp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	inner := ""
	if err := json.Unmarshal([]byte(r), &mcp); err == nil && len(mcp.Content) > 0 {
		for _, c := range mcp.Content {
			if strings.TrimSpace(c.Text) != "" {
				inner = c.Text
				break
			}
		}
	}
	if inner == "" {
		inner = r
	}
	inner = strings.TrimSpace(inner)
	inner = strings.TrimSuffix(inner, ",")

	type Forecast struct {
		Date           string         `json:"date"`
		Week           string         `json:"week"`
		Dayweather     FlexibleString `json:"dayweather"`
		Nightweather   FlexibleString `json:"nightweather"`
		Daytemp        FlexibleString `json:"daytemp"`
		Nighttemp      FlexibleString `json:"nighttemp"`
		Daywind        FlexibleString `json:"daywind"`
		Nightwind      FlexibleString `json:"nightwind"`
		Daypower       FlexibleString `json:"daypower"`
		Nightpower     FlexibleString `json:"nightpower"`
		DaytempFloat   FlexibleString `json:"daytemp_float"`
		NighttempFloat FlexibleString `json:"nighttemp_float"`
	}
	var wr struct {
		City      FlexibleString `json:"city"`
		Forecasts []Forecast     `json:"forecasts"`
	}
	if err := json.Unmarshal([]byte(inner), &wr); err != nil {
		log.Error("构建天气响应显示信息失败, 无法解析JSON", zap.Error(err))
		return "", ErrInvalidJSON
	}
	if len(wr.Forecasts) == 0 {
		return "未解析到天气信息", nil
	}

	city := strings.TrimSpace(wr.City.String())
	var b strings.Builder
	if city != "" {
		b.WriteString("城市：" + html.EscapeString(city) + "<br/>")
	}
	b.WriteString("未来 " + strconv.Itoa(len(wr.Forecasts)) + " 天预报：<br/>")

	for i, f := range wr.Forecasts {
		date := strings.TrimSpace(f.Date)
		weekStr := weekToCN(strings.TrimSpace(f.Week))
		dayw := strings.TrimSpace(f.Dayweather.String())
		nightw := strings.TrimSpace(f.Nightweather.String())
		daytemp := strings.TrimSpace(f.Daytemp.String())
		nighttemp := strings.TrimSpace(f.Nighttemp.String())
		daywind := strings.TrimSpace(f.Daywind.String())
		nightwind := strings.TrimSpace(f.Nightwind.String())
		daypower := strings.TrimSpace(f.Daypower.String())
		nightpower := strings.TrimSpace(f.Nightpower.String())

		b.WriteString(strconv.Itoa(i+1) + ". ")
		if date != "" {
			b.WriteString(html.EscapeString(date))
			if weekStr != "" {
				b.WriteString("（" + weekStr + "）")
			}
			b.WriteString("：")
		}
		if dayw != "" || daytemp != "" {
			b.WriteString("白天 ")
			if dayw != "" {
				b.WriteString(html.EscapeString(dayw))
			}
			if daytemp != "" {
				if dayw != "" {
					b.WriteString(" ")
				}
				b.WriteString(html.EscapeString(daytemp) + "℃")
			}
			b.WriteString("；")
		}
		if nightw != "" || nighttemp != "" {
			b.WriteString("夜间 ")
			if nightw != "" {
				b.WriteString(html.EscapeString(nightw))
			}
			if nighttemp != "" {
				if nightw != "" {
					b.WriteString(" ")
				}
				b.WriteString(html.EscapeString(nighttemp) + "℃")
			}
			b.WriteString("；")
		}
		// 风向/风力
		if daywind != "" || nightwind != "" || daypower != "" || nightpower != "" {
			var parts []string
			if daywind != "" || nightwind != "" {
				w := daywind
				if w == "" {
					w = nightwind
				}
				if daywind != "" && nightwind != "" {
					w = daywind + "/" + nightwind
				}
				parts = append(parts, "风向 "+html.EscapeString(w))
			}
			if daypower != "" || nightpower != "" {
				p := daypower
				if p == "" {
					p = nightpower
				}
				if daypower != "" && nightpower != "" {
					p = daypower + "/" + nightpower
				}
				parts = append(parts, "风力 "+html.EscapeString(p))
			}
			if len(parts) > 0 {
				b.WriteString(strings.Join(parts, "，"))
			}
		}
		if i < len(wr.Forecasts)-1 {
			b.WriteString("<br/>")
		}
	}

	return b.String(), nil
}

func init() {
	RegisterShowMsgBuilder(mapsGeoToolName, mapsGeoShowBuilder{})
	RegisterShowMsgBuilder(mapsTextSearchToolName, mapsTextSearchShowBuilder{})
	RegisterShowMsgBuilder(mapsDirectionToolName, mapsTransitIntegratedShowBuilder{})
	RegisterShowMsgBuilder(mapsDistanceToolName, mapsDistanceShowBuilder{})
	RegisterShowMsgBuilder(mapsSearchDetailToolName, mapsSearchDetailShowBuilder{})
	RegisterShowMsgBuilder(mapsWeatherToolName, mapsWeatherShowBuilder{})
	RegisterShowMsgBuilder(mapsDirectionDrivingToolName, mapsDirectionDrivingShowBuilder{})
	RegisterShowMsgBuilder(mapsDirectionWalkingToolName, mapsDirectionWalkingShowBuilder{})
	RegisterShowMsgBuilder(mapsAroundSearchToolName, mapsAroundSearchShowBuilder{})
}

// 驾车路线展示构造器
type mapsDirectionDrivingShowBuilder struct{}

func (mapsDirectionDrivingShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params struct {
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
	}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建驾车路线请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	origin := strings.TrimSpace(params.Origin)
	destination := strings.TrimSpace(params.Destination)

	var b strings.Builder
	b.WriteString("驾车路线规划：")
	if origin != "" && destination != "" {
		b.WriteString(origin + " → " + destination)
	} else if origin != "" {
		b.WriteString("从 " + origin)
	} else if destination != "" {
		b.WriteString("到 " + destination)
	} else {
		b.WriteString("(未提供坐标)")
	}
	return b.String(), nil
}

func (mapsDirectionDrivingShowBuilder) BuildResponse(response string) (string, error) {
	r := strings.TrimSpace(response)
	if r == "" {
		return "未解析到驾车路线", nil
	}
	// 解析 MCP 包裹的响应 {"content":[{"type":"text","text":"{...}"}]}
	var mcp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	inner := ""
	if err := json.Unmarshal([]byte(r), &mcp); err == nil && len(mcp.Content) > 0 {
		for _, c := range mcp.Content {
			if strings.TrimSpace(c.Text) != "" {
				inner = c.Text
				break
			}
		}
	}
	if inner == "" {
		inner = r
	}
	inner = strings.TrimSpace(inner)
	inner = strings.TrimSuffix(inner, ",")

	// 数据结构（驾车路径）
	type DrivingStep struct {
		Instruction string         `json:"instruction"`
		Road        FlexibleString `json:"road"`
		Distance    string         `json:"distance"`
		Orientation FlexibleString `json:"orientation"`
		Duration    string         `json:"duration"`
	}
	type DrivingPath struct {
		Distance string        `json:"distance"`
		Duration string        `json:"duration"`
		Steps    []DrivingStep `json:"steps"`
	}
	type DrivingRoute struct {
		Origin      string        `json:"origin"`
		Destination string        `json:"destination"`
		Paths       []DrivingPath `json:"paths"`
	}
	var dr struct {
		Route DrivingRoute `json:"route"`
	}
	if err := json.Unmarshal([]byte(inner), &dr); err != nil {
		log.Error("构建驾车路线响应显示信息失败, 无法解析JSON", zap.Error(err))
		return "", ErrInvalidJSON
	}
	if len(dr.Route.Paths) == 0 {
		return "未解析到驾车路线", nil
	}

	var b strings.Builder
	origin := strings.TrimSpace(dr.Route.Origin)
	destination := strings.TrimSpace(dr.Route.Destination)
	if origin != "" || destination != "" {
		b.WriteString("驾车路线（起点 ")
		if origin != "" {
			b.WriteString(origin)
		} else {
			b.WriteString("未知")
		}
		b.WriteString(" → 终点 ")
		if destination != "" {
			b.WriteString(destination)
		} else {
			b.WriteString("未知")
		}
		b.WriteString("）：\n")
	}

	for i, p := range dr.Route.Paths {
		dist := strings.TrimSpace(p.Distance)
		dur := strings.TrimSpace(p.Duration)
		b.WriteString("方案 " + strconv.Itoa(i+1))
		if dist != "" || dur != "" {
			b.WriteString("：")
			if dist != "" {
				b.WriteString("总距离 " + dist + " 米")
			}
			if dur != "" {
				if dist != "" {
					b.WriteString("，")
				}
				b.WriteString("总时长 " + dur + " 秒")
			}
		}
		b.WriteString("\n")
		for _, s := range p.Steps {
			ins := strings.TrimSpace(s.Instruction)
			road := strings.TrimSpace(s.Road.String())
			sdist := strings.TrimSpace(s.Distance)
			sdur := strings.TrimSpace(s.Duration)
			ori := strings.TrimSpace(s.Orientation.String())
			if ins != "" || road != "" || sdist != "" || sdur != "" || ori != "" {
				b.WriteString("  - ")
				if ins != "" {
					b.WriteString(ins)
				}
				var details []string
				if road != "" {
					details = append(details, "道路 "+road)
				}
				if ori != "" {
					details = append(details, "方向 "+ori)
				}
				if len(details) > 0 {
					if ins != "" {
						b.WriteString("（" + strings.Join(details, "，") + "）")
					} else {
						b.WriteString(strings.Join(details, "，"))
					}
				}
				if sdist != "" {
					b.WriteString("，距离 " + sdist + " 米")
				}
				if sdur != "" {
					b.WriteString("，时长 " + sdur + " 秒")
				}
				b.WriteString("\n")
			}
		}
		if i < len(dr.Route.Paths)-1 {
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}

// 步行路线展示构造器
type mapsDirectionWalkingShowBuilder struct{}

func (mapsDirectionWalkingShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params struct {
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
	}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建步行路线请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	origin := strings.TrimSpace(params.Origin)
	destination := strings.TrimSpace(params.Destination)

	var b strings.Builder
	b.WriteString("步行路线规划：")
	if origin != "" && destination != "" {
		b.WriteString(origin + " → " + destination)
	} else if origin != "" {
		b.WriteString("从 " + origin)
	} else if destination != "" {
		b.WriteString("到 " + destination)
	} else {
		b.WriteString("(未提供坐标)")
	}
	return b.String(), nil
}

func (mapsDirectionWalkingShowBuilder) BuildResponse(response string) (string, error) {
	r := strings.TrimSpace(response)
	if r == "" {
		return "未解析到步行路线", nil
	}
	// 解析 MCP 包裹的响应 {"content":[{"type":"text","text":"{...}"}]}
	var mcp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	inner := ""
	if err := json.Unmarshal([]byte(r), &mcp); err == nil && len(mcp.Content) > 0 {
		for _, c := range mcp.Content {
			if strings.TrimSpace(c.Text) != "" {
				inner = c.Text
				break
			}
		}
	}
	if inner == "" {
		inner = r
	}
	inner = strings.TrimSpace(inner)
	inner = strings.TrimSuffix(inner, ",")

	// 数据结构（步行路径）
	type WalkingStep struct {
		Instruction string         `json:"instruction"`
		Road        FlexibleString `json:"road"`
		Distance    string         `json:"distance"`
		Orientation FlexibleString `json:"orientation"`
		Duration    string         `json:"duration"`
	}
	type WalkingPath struct {
		Distance string        `json:"distance"`
		Duration string        `json:"duration"`
		Steps    []WalkingStep `json:"steps"`
	}
	type WalkingRoute struct {
		Origin      string        `json:"origin"`
		Destination string        `json:"destination"`
		Paths       []WalkingPath `json:"paths"`
	}
	var wr struct {
		Route WalkingRoute `json:"route"`
	}
	if err := json.Unmarshal([]byte(inner), &wr); err != nil {
		log.Error("构建步行路线响应显示信息失败, 无法解析JSON", zap.Error(err))
		return "", ErrInvalidJSON
	}
	if len(wr.Route.Paths) == 0 {
		return "未解析到步行路线", nil
	}

	var b strings.Builder
	origin := strings.TrimSpace(wr.Route.Origin)
	destination := strings.TrimSpace(wr.Route.Destination)
	if origin != "" || destination != "" {
		b.WriteString("步行路线（起点 ")
		if origin != "" {
			b.WriteString(origin)
		} else {
			b.WriteString("未知")
		}
		b.WriteString(" → 终点 ")
		if destination != "" {
			b.WriteString(destination)
		} else {
			b.WriteString("未知")
		}
		b.WriteString("）：\n")
	}

	for i, p := range wr.Route.Paths {
		dist := strings.TrimSpace(p.Distance)
		dur := strings.TrimSpace(p.Duration)
		b.WriteString("方案 " + strconv.Itoa(i+1))
		if dist != "" || dur != "" {
			b.WriteString("：")
			if dist != "" {
				b.WriteString("总距离 " + dist + " 米")
			}
			if dur != "" {
				if dist != "" {
					b.WriteString("，")
				}
				b.WriteString("总时长 " + dur + " 秒")
			}
		}
		b.WriteString("\n")
		for _, s := range p.Steps {
			ins := strings.TrimSpace(s.Instruction)
			road := strings.TrimSpace(s.Road.String())
			sdist := strings.TrimSpace(s.Distance)
			sdur := strings.TrimSpace(s.Duration)
			ori := strings.TrimSpace(s.Orientation.String())
			if ins != "" || road != "" || sdist != "" || sdur != "" || ori != "" {
				b.WriteString("  - ")
				if ins != "" {
					b.WriteString(ins)
				}
				var details []string
				if road != "" {
					details = append(details, "道路 "+road)
				}
				if ori != "" {
					details = append(details, "方向 "+ori)
				}
				if len(details) > 0 {
					if ins != "" {
						b.WriteString("（" + strings.Join(details, "，") + "）")
					} else {
						b.WriteString(strings.Join(details, "，"))
					}
				}
				if sdist != "" {
					b.WriteString("，距离 " + sdist + " 米")
				}
				if sdur != "" {
					b.WriteString("，时长 " + sdur + " 秒")
				}
				b.WriteString("\n")
			}
		}
		if i < len(wr.Route.Paths)-1 {
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}

// 周边搜索展示构造器
// 请求示例: {"keywords":"景点","location":"121.489956,31.182650","radius":"5000"}
// 响应为 MCP 包裹或直接 JSON：{"pois":[{...}]}
type mapsAroundSearchShowBuilder struct{}

func (mapsAroundSearchShowBuilder) BuildRequest(paramsStr string) (string, error) {
	var params struct {
		Keywords string `json:"keywords"`
		Location string `json:"location"`
		Radius   string `json:"radius"`
	}
	if err := json.Unmarshal([]byte(paramsStr), &params); err != nil {
		log.Error("构建周边搜索请求显示信息失败", zap.Error(err))
		return "", ErrInvalidJSON
	}
	kw := strings.TrimSpace(params.Keywords)
	loc := strings.TrimSpace(params.Location)
	radius := strings.TrimSpace(params.Radius)

	var b strings.Builder
	b.WriteString("周边搜索：")
	if kw == "" {
		kw = "(未提供关键词)"
	}
	b.WriteString(kw)

	if loc != "" {
		b.WriteString("，位置 " + loc)
	}

	if radius != "" {
		b.WriteString("，半径 " + radius + " 米")
	}

	return b.String(), nil
}

func (mapsAroundSearchShowBuilder) BuildResponse(response string) (string, error) {
	r := strings.TrimSpace(response)
	if r == "" {
		return "未找到周边地点", nil
	}

	// 解析 MCP 包裹的响应 {"content":[{"type":"text","text":"{...}"}]}
	var mcp struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	}
	inner := ""
	if err := json.Unmarshal([]byte(r), &mcp); err == nil && len(mcp.Content) > 0 {
		for _, c := range mcp.Content {
			if strings.TrimSpace(c.Text) != "" {
				inner = c.Text
				break
			}
		}
	}
	if inner == "" {
		inner = r
	}
	inner = strings.TrimSpace(inner)
	inner = strings.TrimSuffix(inner, ",")

	// 解析周边搜索结果
	type photoInfo struct {
		Provider []string `json:"provider"`
		Title    []string `json:"title"`
		Url      string   `json:"url"`
	}
	type poiItem struct {
		ID       string     `json:"id"`
		Name     string     `json:"name"`
		Address  string     `json:"address"`
		Typecode string     `json:"typecode"`
		Photos   *photoInfo `json:"photos"`
	}
	type aroundSearchResp struct {
		Pois []poiItem `json:"pois"`
	}

	var as aroundSearchResp
	if err := json.Unmarshal([]byte(inner), &as); err != nil {
		log.Error("构建周边搜索响应显示信息失败, 无法解析JSON", zap.Error(err))
		return "", ErrInvalidJSON
	}

	count := len(as.Pois)
	if count == 0 {
		return "未找到周边地点", nil
	}

	var msgBuilder strings.Builder
	msgBuilder.WriteString("共找到 " + strconv.Itoa(count) + " 个周边地点：<br/>")

	for i, p := range as.Pois {
		name := strings.TrimSpace(p.Name)
		addr := strings.TrimSpace(p.Address)
		typecode := strings.TrimSpace(p.Typecode)
		url := ""
		if p.Photos != nil {
			u := strings.TrimSpace(p.Photos.Url)
			if u != "" {
				// 去除可能包裹的反引号和引号
				u = strings.Trim(u, " `\"")
				url = u
			}
		}

		msgBuilder.WriteString(strconv.Itoa(i+1) + ". ")
		if name != "" {
			msgBuilder.WriteString(name)
		}
		if addr != "" {
			msgBuilder.WriteString("，地址 " + addr)
		}
		if typecode != "" {
			msgBuilder.WriteString("，类型 " + typecode)
		}
		if url != "" {
			msgBuilder.WriteString("<br/><img src=\"" + html.EscapeString(url) + "\" alt=\"" + html.EscapeString(name) + "\" style=\"max-width:120px;max-height:100px;object-fit:cover;border-radius:6px;margin-left:6px;\"/>")
		}
		if i < count-1 {
			msgBuilder.WriteString("<br/>")
		}
	}

	return msgBuilder.String(), nil
}
