package vo

import (
	"time"

	"txing-ai/internal/domain"
)

// AboutMeHeroVO 关于我 - Hero 视图对象
// About Me Hero view object
type AboutMeHeroVO struct {
	Id         int64     `json:"id"`
	AvatarText string    `json:"avatarText"`
	StatusText string    `json:"statusText"`
	Name       string    `json:"name"`
	Subtitle   string    `json:"subtitle"`
	CreateAt   time.Time `json:"createAt"`
	UpdateAt   time.Time `json:"updateAt"`
}

func ToAboutMeHeroVO(h domain.AboutMeHero) AboutMeHeroVO {
	return AboutMeHeroVO{
		Id:         h.Id,
		AvatarText: h.AvatarText,
		StatusText: h.StatusText,
		Name:       h.Name,
		Subtitle:   h.Subtitle,
		CreateAt:   h.CreateTime,
		UpdateAt:   h.UpdateTime,
	}
}

// AboutMeFloatingIconVO 浮动图标视图对象
// Floating icon view object
type AboutMeFloatingIconVO struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Sort   int    `json:"sort"`
}

func ToAboutMeFloatingIconVO(f domain.AboutMeFloatingIcon) AboutMeFloatingIconVO {
	return AboutMeFloatingIconVO{
		Id:     f.Id,
		Name:   f.Name,
		Symbol: f.Symbol,
		Sort:   f.Sort,
	}
}

func ToAboutMeFloatingIconVOs(items []domain.AboutMeFloatingIcon) []AboutMeFloatingIconVO {
	out := make([]AboutMeFloatingIconVO, 0, len(items))
	for _, it := range items {
		out = append(out, ToAboutMeFloatingIconVO(it))
	}
	return out
}

// AboutMeReasonVO 为什么选择我视图对象
type AboutMeReasonVO struct {
	Id    int64              `json:"id"`
	Emoji string             `json:"emoji"`
	Title string             `json:"title"`
	Desc  string             `json:"desc"`
	Tags  []string           `json:"tags"`
	Stats []AboutMeStatItem  `json:"stats"`
	Sort  int                `json:"sort"`
}

type AboutMeStatItem struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

func ToAboutMeReasonVO(r domain.AboutMeReason) AboutMeReasonVO {
	stats := make([]AboutMeStatItem, 0, len(r.Stats))
	for _, s := range r.Stats {
		stats = append(stats, AboutMeStatItem{Value: s.Value, Label: s.Label})
	}
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}
	return AboutMeReasonVO{
		Id:    r.Id,
		Emoji: r.Emoji,
		Title: r.Title,
		Desc:  r.Desc,
		Tags:  tags,
		Stats: stats,
		Sort:  r.Sort,
	}
}

func ToAboutMeReasonVOs(items []domain.AboutMeReason) []AboutMeReasonVO {
	out := make([]AboutMeReasonVO, 0, len(items))
	for _, it := range items {
		out = append(out, ToAboutMeReasonVO(it))
	}
	return out
}

// AboutMeSkillVO 技能视图对象
type AboutMeSkillVO struct {
	Id       int64    `json:"id"`
	Category string   `json:"category"`
	IconKey  string   `json:"iconKey"`
	Tags     []string `json:"tags"`
	Level    int      `json:"level"`
	Sort     int      `json:"sort"`
}

func ToAboutMeSkillVO(s domain.AboutMeSkill) AboutMeSkillVO {
	tags := s.Tags
	if tags == nil {
		tags = []string{}
	}
	return AboutMeSkillVO{
		Id:       s.Id,
		Category: s.Category,
		IconKey:  s.IconKey,
		Tags:     tags,
		Level:    s.Level,
		Sort:     s.Sort,
	}
}

func ToAboutMeSkillVOs(items []domain.AboutMeSkill) []AboutMeSkillVO {
	out := make([]AboutMeSkillVO, 0, len(items))
	for _, it := range items {
		out = append(out, ToAboutMeSkillVO(it))
	}
	return out
}

// AboutMeProjectVO 作品视图对象
type AboutMeProjectVO struct {
	Id         int64                   `json:"id"`
	Name       string                  `json:"name"`
	Desc       string                  `json:"desc"`
	IconKey    string                  `json:"iconKey"`
	Gradient   string                  `json:"gradient"`
	Tags       []string                `json:"tags"`
	Link       string                  `json:"link"`
	Badge      string                  `json:"badge"`
	Highlights []string                `json:"highlights"`
	Media      []AboutMeMediaItem      `json:"media"`
	TechStack  []AboutMeTechItem       `json:"techStack"`
	Features   []AboutMeFeatureItem    `json:"features"`
	Sort       int                     `json:"sort"`
}

type AboutMeMediaItem struct {
	Type    string `json:"type"`
	URL     string `json:"url"`
	Caption string `json:"caption"`
}

type AboutMeTechItem struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type AboutMeFeatureItem struct {
	Icon  string `json:"icon"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

func ToAboutMeProjectVO(p domain.AboutMeProject) AboutMeProjectVO {
	media := make([]AboutMeMediaItem, 0, len(p.Media))
	for _, m := range p.Media {
		media = append(media, AboutMeMediaItem{Type: m.Type, URL: m.URL, Caption: m.Caption})
	}
	tech := make([]AboutMeTechItem, 0, len(p.TechStack))
	for _, t := range p.TechStack {
		tech = append(tech, AboutMeTechItem{Name: t.Name, Icon: t.Icon})
	}
	feats := make([]AboutMeFeatureItem, 0, len(p.Features))
	for _, f := range p.Features {
		feats = append(feats, AboutMeFeatureItem{Icon: f.Icon, Title: f.Title, Desc: f.Desc})
	}
	tags := p.Tags
	if tags == nil {
		tags = []string{}
	}
	highlights := p.Highlights
	if highlights == nil {
		highlights = []string{}
	}
	return AboutMeProjectVO{
		Id:         p.Id,
		Name:       p.Name,
		Desc:       p.Desc,
		IconKey:    p.IconKey,
		Gradient:   p.Gradient,
		Tags:       tags,
		Link:       p.Link,
		Badge:      p.Badge,
		Highlights: highlights,
		Media:      media,
		TechStack:  tech,
		Features:   feats,
		Sort:       p.Sort,
	}
}

func ToAboutMeProjectVOs(items []domain.AboutMeProject) []AboutMeProjectVO {
	out := make([]AboutMeProjectVO, 0, len(items))
	for _, it := range items {
		out = append(out, ToAboutMeProjectVO(it))
	}
	return out
}

// AboutMeTimelineVO 时间线条目视图对象
type AboutMeTimelineVO struct {
	Id    int64    `json:"id"`
	Time  string   `json:"time"`
	Title string   `json:"title"`
	Desc  string   `json:"desc"`
	Tags  []string `json:"tags"`
	Sort  int      `json:"sort"`
}

func ToAboutMeTimelineVO(t domain.AboutMeTimeline) AboutMeTimelineVO {
	tags := t.Tags
	if tags == nil {
		tags = []string{}
	}
	return AboutMeTimelineVO{
		Id:    t.Id,
		Time:  t.Time,
		Title: t.Title,
		Desc:  t.Desc,
		Tags:  tags,
		Sort:  t.Sort,
	}
}

func ToAboutMeTimelineVOs(items []domain.AboutMeTimeline) []AboutMeTimelineVO {
	out := make([]AboutMeTimelineVO, 0, len(items))
	for _, it := range items {
		out = append(out, ToAboutMeTimelineVO(it))
	}
	return out
}

// AboutMeContactVO 联系区视图对象
type AboutMeContactVO struct {
	Id    int64                 `json:"id"`
	Title string                `json:"title"`
	Desc  string                `json:"desc"`
	Links []AboutMeContactLink  `json:"links"`
}

type AboutMeContactLink struct {
	IconKey string `json:"iconKey"`
	Label   string `json:"label"`
	URL     string `json:"url"`
}

func ToAboutMeContactVO(c domain.AboutMeContact) AboutMeContactVO {
	links := make([]AboutMeContactLink, 0, len(c.Links))
	for _, l := range c.Links {
		links = append(links, AboutMeContactLink{IconKey: l.IconKey, Label: l.Label, URL: l.URL})
	}
	return AboutMeContactVO{
		Id:    c.Id,
		Title: c.Title,
		Desc:  c.Desc,
		Links: links,
	}
}

// AboutMeSnapshotVO 关于我页面整体快照（公开端点返回）
// Aggregated about me snapshot returned by GET /api/about
type AboutMeSnapshotVO struct {
	Hero         AboutMeHeroVO          `json:"hero"`
	FloatingIcon []AboutMeFloatingIconVO `json:"floatingIcons"`
	Reasons      []AboutMeReasonVO      `json:"reasons"`
	Skills       []AboutMeSkillVO       `json:"skills"`
	Projects     []AboutMeProjectVO     `json:"projects"`
	Timeline     []AboutMeTimelineVO    `json:"timeline"`
	Contact      AboutMeContactVO       `json:"contact"`
}