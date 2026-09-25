package wiki

import (
	"context"
	"sort"
	"strings"

	"txing-ai/internal/domain"
)

// 知识图谱只读视图（管理后台可视化用）
// 数据源：已发布页为节点，wiki_links 出链为边（[[slug|文字]] 互链在 confirm/update 时落库）

// GraphNode 图谱节点（已发布页；dangling 为悬空目标——被链接但尚无对应已发布页的 slug）
type GraphNode struct {
	Id       int64  `json:"id"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	PageType string `json:"pageType"`
	Summary  string `json:"summary"`
	Degree   int    `json:"degree"`
	Dangling bool   `json:"dangling"`
}

// GraphEdge 图谱边（同一对页面的多条链接去重为一条，anchorText 用 / 拼接）
type GraphEdge struct {
	Source     string `json:"source"`
	Target     string `json:"target"`
	AnchorText string `json:"anchorText"`
}

// GraphData 图谱整体数据（节点 + 边；概览统计由前端按节点/边派生）
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// ListGraph 构建已发布页的知识图谱：节点按度排序，边按 (source,target) 去重
func (s *Service) ListGraph(ctx context.Context) (*GraphData, error) {
	// 已发布页 → 节点（slug 在已发布集合内唯一，confirm 按 slug 合并保证）
	var pages []domain.WikiPage
	if err := s.db.WithContext(ctx).
		Select("id", "slug", "title", "page_type", "summary").
		Where("status = ?", domain.WikiPageStatusPublished).
		Find(&pages).Error; err != nil {
		return nil, err
	}

	nodeBySlug := make(map[string]*GraphNode, len(pages))
	for i := range pages {
		p := &pages[i]
		nodeBySlug[p.Slug] = &GraphNode{
			Id:       p.Id,
			Slug:     p.Slug,
			Title:    p.Title,
			PageType: p.PageType,
			Summary:  p.Summary,
		}
	}

	// 出链 → 边（只统计已发布页发出的链接）
	// 注意不可给 wiki_links 起别名：GORM 会按模型表名注入软删除条件 wiki_links.delete_time IS NULL；
	// 源页 slug 随 JOIN 直接带出（JOIN 条件已限定源页为已发布，无需 page_id→slug 映射）
	var links []struct {
		FromSlug   string
		ToSlug     string
		AnchorText string
	}
	if err := s.db.WithContext(ctx).
		Model(&domain.WikiLink{}).
		Select("p.slug AS from_slug", "wiki_links.to_slug", "wiki_links.anchor_text").
		Joins("JOIN wiki_pages p ON wiki_links.from_page_id = p.id AND p.status = ?", domain.WikiPageStatusPublished).
		Scan(&links).Error; err != nil {
		return nil, err
	}

	edgeByPair := make(map[[2]string]*GraphEdge, len(links))
	addEdge := func(srcSlug, dstSlug, anchor string) {
		key := [2]string{srcSlug, dstSlug}
		e, ok := edgeByPair[key]
		if !ok {
			e = &GraphEdge{Source: srcSlug, Target: dstSlug}
			edgeByPair[key] = e
		}
		if anchor != "" && !strings.Contains(e.AnchorText, anchor) {
			e.AnchorText = strings.TrimSpace(e.AnchorText + " / " + anchor)
		}
	}

	for _, l := range links {
		if l.ToSlug == l.FromSlug {
			continue // 自链不成边
		}
		// 悬空目标补一个节点（提示待补充或待修正的链接，前端渲染为灰色空心）；
		// 只以 Dangling 布尔标记，不改写 PageType——该字段语义仍属于真实页面
		if _, ok := nodeBySlug[l.ToSlug]; !ok {
			nodeBySlug[l.ToSlug] = &GraphNode{Slug: l.ToSlug, Title: l.ToSlug, Dangling: true}
		}
		addEdge(l.FromSlug, l.ToSlug, l.AnchorText)
	}

	// 汇总输出：度数 + 排序（度大的在前）
	edges := make([]GraphEdge, 0, len(edgeByPair))
	for _, e := range edgeByPair {
		edges = append(edges, *e)
		nodeBySlug[e.Source].Degree++
		nodeBySlug[e.Target].Degree++
	}
	nodes := make([]GraphNode, 0, len(nodeBySlug))
	for _, n := range nodeBySlug {
		nodes = append(nodes, *n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Degree != nodes[j].Degree {
			return nodes[i].Degree > nodes[j].Degree
		}
		return nodes[i].Slug < nodes[j].Slug
	})

	return &GraphData{Nodes: nodes, Edges: edges}, nil
}
