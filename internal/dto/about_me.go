package dto

import "txing-ai/internal/utils/page"

// 通用请求体包装（直接复用 domain 的 stats/media 等嵌套结构）

type UpdateAboutMeHeroReq struct {
	AvatarText string `json:"avatarText" binding:"required"`
	StatusText string `json:"statusText"`
	Name       string `json:"name" binding:"required"`
	Subtitle   string `json:"subtitle"`
}

// === Floating Icon ===

type CreateAboutMeFloatingIconReq struct {
	Name   string `json:"name" binding:"required"`
	Symbol string `json:"symbol" binding:"required"`
	Sort   int    `json:"sort"`
}

type UpdateAboutMeFloatingIconReq struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Sort   int    `json:"sort"`
}

type ListAboutMeFloatingIconReq struct {
	page.PageRequest
}

// === Reason ===

type CreateAboutMeReasonReq struct {
	Emoji string   `json:"emoji" binding:"required"`
	Title string   `json:"title" binding:"required"`
	Desc  string   `json:"desc" binding:"required"`
	Tags  []string `json:"tags"`
	Stats []struct {
		Value string `json:"value"`
		Label string `json:"label"`
	} `json:"stats"`
	Sort int `json:"sort"`
}

type UpdateAboutMeReasonReq struct {
	Emoji string   `json:"emoji"`
	Title string   `json:"title"`
	Desc  string   `json:"desc"`
	Tags  []string `json:"tags"`
	Stats []struct {
		Value string `json:"value"`
		Label string `json:"label"`
	} `json:"stats"`
	Sort int `json:"sort"`
}

type ListAboutMeReasonReq struct {
	page.PageRequest
}

// === Skill ===

type CreateAboutMeSkillReq struct {
	Category string   `json:"category" binding:"required"`
	IconKey  string   `json:"iconKey" binding:"required"`
	Tags     []string `json:"tags"`
	Level    int      `json:"level"`
	Sort     int      `json:"sort"`
}

type UpdateAboutMeSkillReq struct {
	Category string   `json:"category"`
	IconKey  string   `json:"iconKey"`
	Tags     []string `json:"tags"`
	Level    int      `json:"level"`
	Sort     int      `json:"sort"`
}

type ListAboutMeSkillReq struct {
	page.PageRequest
}

// === Project ===

type CreateAboutMeProjectReq struct {
	Name       string   `json:"name" binding:"required"`
	Desc       string   `json:"desc" binding:"required"`
	IconKey    string   `json:"iconKey" binding:"required"`
	Gradient   string   `json:"gradient"`
	Tags       []string `json:"tags"`
	Link       string   `json:"link"`
	Badge      string   `json:"badge"`
	Highlights []string `json:"highlights"`
	Media      []struct {
		Type    string `json:"type"`
		URL     string `json:"url"`
		Caption string `json:"caption"`
	} `json:"media"`
	TechStack []struct {
		Name string `json:"name"`
		Icon string `json:"icon"`
	} `json:"techStack"`
	Features []struct {
		Icon  string `json:"icon"`
		Title string `json:"title"`
		Desc  string `json:"desc"`
	} `json:"features"`
	Sort int `json:"sort"`
}

type UpdateAboutMeProjectReq struct {
	Name       string   `json:"name"`
	Desc       string   `json:"desc"`
	IconKey    string   `json:"iconKey"`
	Gradient   string   `json:"gradient"`
	Tags       []string `json:"tags"`
	Link       string   `json:"link"`
	Badge      string   `json:"badge"`
	Highlights []string `json:"highlights"`
	Media      []struct {
		Type    string `json:"type"`
		URL     string `json:"url"`
		Caption string `json:"caption"`
	} `json:"media"`
	TechStack []struct {
		Name string `json:"name"`
		Icon string `json:"icon"`
	} `json:"techStack"`
	Features []struct {
		Icon  string `json:"icon"`
		Title string `json:"title"`
		Desc  string `json:"desc"`
	} `json:"features"`
	Sort int `json:"sort"`
}

type ListAboutMeProjectReq struct {
	page.PageRequest
}

// === Timeline ===

type CreateAboutMeTimelineReq struct {
	Time  string   `json:"time" binding:"required"`
	Title string   `json:"title" binding:"required"`
	Desc  string   `json:"desc" binding:"required"`
	Tags  []string `json:"tags"`
	Sort  int      `json:"sort"`
}

type UpdateAboutMeTimelineReq struct {
	Time  string   `json:"time"`
	Title string   `json:"title"`
	Desc  string   `json:"desc"`
	Tags  []string `json:"tags"`
	Sort  int      `json:"sort"`
}

type ListAboutMeTimelineReq struct {
	page.PageRequest
}

// === Contact ===

type UpdateAboutMeContactReq struct {
	Title string `json:"title" binding:"required"`
	Desc  string `json:"desc"`
	Links []struct {
		IconKey string `json:"iconKey"`
		Label   string `json:"label"`
		URL     string `json:"url"`
	} `json:"links"`
}