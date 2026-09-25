package domain

// 关于我页面 - Hero 区配置
// About Me page - Hero section configuration
type AboutMeHero struct {
	BaseModel
	AvatarText string `gorm:"type:varchar(10);not null;comment:头像占位文字" json:"avatarText"`
	StatusText string `gorm:"type:varchar(100);comment:状态徽标文字" json:"statusText"`
	Name       string `gorm:"type:varchar(50);not null;comment:打字机显示的名字" json:"name"`
	Subtitle   string `gorm:"type:varchar(255);comment:打字机显示的副标题" json:"subtitle"`
}

// 浮动小图标
// Floating tech icons in hero
type AboutMeFloatingIcon struct {
	BaseModel
	Name   string `gorm:"type:varchar(50);not null;comment:图标名" json:"name"`
	Symbol string `gorm:"type:varchar(20);not null;comment:显示符号(emoji/字符)" json:"symbol"`
	Sort   int    `gorm:"type:int;default:0;comment:排序" json:"sort"`
}

// 为什么选择我卡片
// Why choose me card
type AboutMeReason struct {
	BaseModel
	Emoji string   `gorm:"type:varchar(10);not null;comment:emoji 符号" json:"emoji"`
	Title string   `gorm:"type:varchar(100);not null;comment:标题" json:"title"`
	Desc  string   `gorm:"type:text;not null;comment:描述" json:"desc"`
	Tags  []string `gorm:"type:json;serializer:json;comment:标签列表" json:"tags"`
	Stats []struct {
		Value string `json:"value"`
		Label string `json:"label"`
	} `gorm:"type:json;serializer:json;comment:统计数据" json:"stats"`
	Sort int `gorm:"type:int;default:0;comment:排序" json:"sort"`
}

// 核心能力分类
// Skill set category
type AboutMeSkill struct {
	BaseModel
	Category string   `gorm:"type:varchar(50);not null;comment:分类名" json:"category"`
	IconKey  string   `gorm:"type:varchar(50);not null;comment:Element Plus 图标 key" json:"iconKey"`
	Tags     []string `gorm:"type:json;serializer:json;comment:技能标签" json:"tags"`
	Level    int      `gorm:"type:int;default:0;comment:熟练度 0-100" json:"level"`
	Sort     int      `gorm:"type:int;default:0;comment:排序" json:"sort"`
}

// AboutMeMediaItem 项目媒体项（图片/视频，落库只存 COS key，读取时换签名 URL）
// Project media item; only the COS key is persisted, signed URLs are generated on read
type AboutMeMediaItem struct {
	Type    string `json:"type"`
	Key     string `json:"key"`
	Caption string `json:"caption"`
	// 兼容旧数据：历史版本直接存完整签名 URL（会过期），读取时会提取 key 并重新签名
	// Legacy: old rows store full signed URLs; keys are extracted and re-signed on read
	URL string `json:"url,omitempty"`
}

// AboutMeCoverMediaItem 封面轮播项（图片/视频混合）
// Cover carousel item (mixed image/video)
type AboutMeCoverMediaItem struct {
	Type string `json:"type"`
	Key  string `json:"key"`
	URL  string `json:"url,omitempty"` // 读取时填充签名 URL / filled with signed URL on read
}

// 精选作品
// Featured project
type AboutMeProject struct {
	BaseModel
	Name       string             `gorm:"type:varchar(100);not null;comment:项目名" json:"name"`
	Desc       string             `gorm:"type:text;not null;comment:项目描述" json:"desc"`
	IconKey    string             `gorm:"type:varchar(50);not null;comment:Element Plus 图标 key" json:"iconKey"`
	Gradient   string             `gorm:"type:varchar(20);default:'1';comment:封面渐变编号 1-6" json:"gradient"`
	Tags       []string           `gorm:"type:json;serializer:json;comment:项目标签" json:"tags"`
	Link       string             `gorm:"type:varchar(500);comment:跳转链接" json:"link"`
	Badge      string             `gorm:"type:varchar(50);comment:角标文字" json:"badge"`
	Category   string             `gorm:"type:varchar(20);default:'company';comment:项目类别 company/personal" json:"category"`
	Highlights []string           `gorm:"type:json;serializer:json;comment:亮点列表" json:"highlights"`
	Media      []AboutMeMediaItem `gorm:"type:json;serializer:json;comment:媒体列表(图/视频,存key)" json:"media"`
	// 封面轮播：卡片左侧大图区，支持图片/视频混合多项轮播
	// Cover carousel: card's left visual area, mixed image/video multi-item carousel
	CoverMedia []AboutMeCoverMediaItem `gorm:"type:json;serializer:json;comment:封面轮播媒体(图/视频,存key)" json:"coverMedia"`
	TechStack  []struct {
		Name string `json:"name"`
		Icon string `json:"icon"`
	} `gorm:"type:json;serializer:json;comment:技术栈" json:"techStack"`
	// 系统架构总览（Markdown，支持 mermaid 图表），前端在项目详情中独立展示
	// System architecture overview (Markdown with mermaid support), shown standalone in project detail
	Architecture string `gorm:"type:text;comment:系统架构(Markdown,支持mermaid)" json:"architecture"`
	Features     []struct {
		Icon   string `json:"icon"`
		Title  string `json:"title"`
		Desc   string `json:"desc"`
		Detail string `json:"detail"` // 详细设计（前端点击展开展示 / expanded on click in frontend）
	} `gorm:"type:json;serializer:json;comment:工作亮点" json:"features"`
	Sort int `gorm:"type:int;default:0;comment:排序" json:"sort"`
}

// 成长轨迹
// Timeline / growth track
type AboutMeTimeline struct {
	BaseModel
	Time  string   `gorm:"type:varchar(100);not null;comment:时间范围" json:"time"`
	Title string   `gorm:"type:varchar(100);not null;comment:标题" json:"title"`
	Desc  string   `gorm:"type:text;not null;comment:描述" json:"desc"`
	Tags  []string `gorm:"type:json;serializer:json;comment:标签" json:"tags"`
	Sort  int      `gorm:"type:int;default:0;comment:排序" json:"sort"`
}

// 联系区
// Contact section
type AboutMeContact struct {
	BaseModel
	Title string `gorm:"type:varchar(100);not null;comment:标题" json:"title"`
	Desc  string `gorm:"type:text;comment:描述" json:"desc"`
	Links []struct {
		IconKey string `json:"iconKey"`
		Label   string `json:"label"`
		URL     string `json:"url"`
	} `gorm:"type:json;serializer:json;comment:联系链接" json:"links"`
}

// 关于我页面整体 VO（聚合只读视图）
// About me aggregated read-only VO
type AboutMeSnapshot struct {
	Hero          AboutMeHero           `json:"hero"`
	FloatingIcons []AboutMeFloatingIcon `json:"floatingIcons"`
	Reasons       []AboutMeReason       `json:"reasons"`
	Skills        []AboutMeSkill        `json:"skills"`
	Projects      []AboutMeProject      `json:"projects"`
	Timeline      []AboutMeTimeline     `json:"timeline"`
	Contact       AboutMeContact        `json:"contact"`
}
