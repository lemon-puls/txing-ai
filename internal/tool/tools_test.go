package tool

import "testing"

func TestBuildResponseShowMsg(t *testing.T) {
	type args struct {
		toolName string
		response string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "TestBuildResponseShowMsg",
			args: args{
				toolName: webSearchToolName,
				response: "{\"date\":\"Sep 28, 2025\",\"displayed_link\":\"https://shuyeidc.com › ...\",\"domain\":\"shuyeidc.com\",\"link\":\"https://shuyeidc.com/wp/370569.html\",\"position\":1,\"snippet\":\"A1：腾讯对学历的要求通常为本科及以上，部分核心技术岗位可能偏好硕士或博士学历，但更注重实际能力，工作经验方面，初级岗位（如开发工程师）要求1-3年经验，中 ...\",\"snippet_highlighted_words\":[\"腾讯\",\"要求\",\"岗位\",\"岗位\",\"开发工程师\",\"要求\"],\"source\":\"树叶云\",\"title\":\"腾讯Java招聘有何具体要求？\"},{\"displayed_link\":\"https://www.zhipin.com › zhaopin\",\"domain\":\"www.zhipin.com\",\"favicon\":\"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAAkFBMVEVd1cjfum3/sUL/s0r/y41L1s7/vWj/7tz/4sP/rjb/1qf/27L/tlL/xoD/yYlb1cf/4L//z5dT08ZO0sRv2c1p18vZuGn/2a6P4Nb////l9/Wy6eK97Obx+/rb9PE7z8DQ9faC3dOp59+K497R8e3Y8/D5/f3s+fj/7t2j5d0rzb1r3NaX4tnG7unM8Oyl5d4p5eg3AAABnElEQVR4AXTR55abMBQEYLl349UFh1HzSsEiSpD9/m+XC8eFlJ0fgw7nAzXxbybT2Ww6nXP3+Q9YsFiu1g8h/is2291qtZx+BVjsd6vl4Tj9AhQF/2OzPX3wLA8gCyJuWUopBFWlXJwZfNvxLMMXNQBVkAZgqLKAu3weZvPTx+owfQDpUdcIMqDWkFR6fD/vj6vT6fgAoVF0tRdxiU2Na0kVWhbrH+vNA7jWVkqRIOUuCdDkW3yep9vNcw3FTxgDSRKG6HIFV8CvyWb2BKZBJxBzhDSNjq1XWmE4U9HH55xrWVTZGVH4hpuHt7LoxeMcpCz6B8lHP4Ys+n4iMQp/MA5xOt23kFxSUqhpTGzbtgCX9Zo7mwx7SyMRUupc7FIK1R0GCTkq5eh9fVWMEeCqqUOGh0WM1zcQJUzHSegYKAYhXhv3B7BDBtDB5+Z2a4Icg0YbzgCAG2w/ZzkGJnHCADxSDWfMXYwWiUc60mgBH22EeU9R6ucicxmcRk3+nqBfoKhja4e0MRBR9DLEqMriPQW90r/kko/h7w0RAADo5jXfbAKo1wAAAABJRU5ErkJggg==\",\"link\":\"https://www.zhipin.com/zhaopin/dd9afdff749e54901XJ43dy_/\",\"position\":2,\"snippet\":\"本科及以上学历； 精通java或c++程序设计和开发； 有较强的分析调试能力； 熟悉Unix/Linux网络编程、数据结构和算法熟悉操作系统基本实现原理； 至 ...\",\"snippet_highlighted_words\":[\"java\",\"开发\"],\"source\":\"BOSS直聘\",\"title\":\"腾讯Java招聘信息\"},{\"date\":\"Feb 4, 2024\",\"displayed_link\":\"https://blog.csdn.net › details\",\"domain\":\"blog.csdn.net\",\"favicon\":\"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABwAAAAcCAMAAABF0y+mAAAAb1BMVEX8VTH8USz8Tyj8Tif8Uy78TST8eF38kn79mIb8fGP8Wjb+zsb+493/8/D+6+j+0sv9nYr8b1L/9/X////9xbj9va7+3NP9qpn8Zkj+6OL8Rhb8Sx/9x7v8gmr8YD39pJT8iHD/+vn+tqn8j3j9sKPVly74AAAA5klEQVR4Aa2SRZLDQBAENSQcsKbErJX//8Ulk3DRec2uiCbnSRDKKOF7hgvH9XwvCNlOyvEjqZSW5rQJ0ziyuGCNWDsJwCZpBgD5wpKiBJRX1XVVREBSz2XdAGUrLh2nOibzYAzY07VLTkO6CEaA6e49LprloYLthbMLHUZol+xL0Y+QzgH1CyCrv0iRA5IcSDqtGpoX8s5iPNcPFbOZrRNATzW/nDU05UDnUQDqHFbvOCcNZPP2WHu5pJQKHzT1cklyxA0VdXzZcZEbqZXSZdZ49fpPiAjdyfe9tmB07/0IZZfX/DdvphAP/wvKn48AAAAASUVORK5CYII=\",\"link\":\"https://blog.csdn.net/AI_data_cloud/article/details/136021408\",\"position\":3,\"snippet\":\"1.本科及以上学历，5年以上相关工作经验; · 2.精通J2EE，精通SSH及常用开源框架和技术，深刻理解各框架的优缺点，熟悉分布式、多线程、缓存、消息等高性能架构 ...\",\"source\":\"CSDN博客\",\"title\":\"高级Java开发工程师岗位的基本职责（合集） 原创\"},{\"date\":\"Dec 10, 2021\",\"displayed_link\":\"https://cloud.tencent.com › news\",\"domain\":\"cloud.tencent.com\",\"favicon\":\"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAACNklEQVR4AezTA+wcQRTH8fkbQW0jrq2gCGvbtts4qWK7NqLatm3b5vnu9ZtkLpnsaXdrTPI57pvfWzwlIr/U/wZ+eQPO1jjJw2Csw1OEEMZzbMBg5KvvuMzwPngGSeElBn7P4CwsgsThgwcSx0Jkfo8GFkMMdzAKZY1jymI07kIMi741vCfEMB+5SY7PjdNwH7fh2XgK0ebZrEuzNPEcOW4fOtEeIs/htNzDdexHUzcNrIZoU1zUpylX6/jtGpilVr04gctq0euzauujOfxW2vYey6U0pmE9DmAjZqFGsuBCWANJIICZyEwSnIWZCEASWINC1vBSuAGxYS+KxAnPx3aIDTdQKhqehkMQLYIVaINa6IGDEMMtVDHCi+EkxHAQ3VALbbAcEYh2CGmKjTpAtBDaK2MZTU5GGKKdM8IvQQzTkRbnKrVHCKJ1UGy0EaLNTvGAtocHj1GVDcriFkQLY0CKB3Q2RNuo2OgNRCtvY0rq6vBKuA3RAuhqY0LKQ7Q3ygj3OxizqnhsCW/noN4P6kDwJ4hW0EZxdTyBaB60chBeEKJ9UpYJGGEj/JW5AVooJ2u5jFDmJBA6BqK9RtUEhQ3xHqJ9QEOH4VXxGqKNUQTm4j5Ee4XRKGEUNsMHiPYKDRwEl8Boy9W7j9zok90YHoiFRx25d0mtDHssc+7He5s8EAsPGlvHqx7uQaIIv6lWhbxG4fdwD3UTzXguhmMX4efMM/9Gr7ATI/F1yI2Tge8XjDpg1AEAsMEBDrhlyj8AAAAASUVORK5CYII=\",\"link\":\"https://cloud.tencent.com/developer/news/882920\",\"position\":4,\"snippet\":\"成为优秀程序员需70%技术+30%软技能。技术方面涵盖Java基础、主流工具（如Maven、Git）、框架（Spring MVC）、应用服务器及云开发；软技能包括沟通、解决 ...\",\"snippet_highlighted_words\":[\"成为优秀程序员需70%技术+30%软技能\"],\"source\":\"腾讯云\",\"title\":\"如何成为合格的Java开发人员\"},{\"date\":\"Oct 13, 2020\",\"displayed_link\":\"https://zhuanlan.zhihu.com › ...\",\"domain\":\"zhuanlan.zhihu.com\",\"favicon\":\"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABwAAAAcCAMAAABF0y+mAAAAMFBMVEUAZv8Chv8Cev8Bk/8BcP8AoP/7/P9glf+rw//P3f9Ifv/o7v+Dpv8BWP8nbf8qkv8dvAK6AAABD0lEQVQokXWSia6EIAxFy2I3Cvz/376y+Ywzc6PR5HBKpUKMYSWFNHKl6w7EeHBKbzpguOUXneZ2U3hVhvygZ9+UDsx5Vo4Pem3qME45vORj5l+lB2y86ek6hN6ppws8LErUyKORmUVKQURKCSA3fy/F/EbJ5pTRjMjbctGKojbClllyFIuEMRROwwRqDRUONCTCLKUvCGZexyaEDKyEwD2GASvPPRnb6AzvlDDMWhVrpQlBldTQu6c+obuC3lD155yD77m+dkBF4kLoS2Qe2BNqkVqxGbKRQxVBWEftRY2br/CrwhRNaB/lbAjuzDFA3lOCdx5Tih9s0TXiD3PJ6wf4gv7db+a98W/m9Lu5+R9rdwulN9KtVAAAAABJRU5ErkJggg==\",\"link\":\"https://zhuanlan.zhihu.com/p/265358398\",\"position\":5,\"snippet\":\"相比阿里和百度，腾讯在招的Java研发工程师岗位就较少了，只有财付通的金融科技部门和外汇交易部门在招人。 尽管三家企业对不同部门Java研发工程师的技能 ...\",\"snippet_highlighted_words\":[\"腾讯\",\"Java\",\"工程师岗位\",\"Java\",\"工程师\",\"技能\"],\"source\":\"知乎专栏\",\"title\":\"【面试类】大厂Java研发工程师的岗位要求你够格吗？\"}",
			},
			want:    "test response",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildResponseShowMsg(tt.args.toolName, tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildResponseShowMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BuildResponseShowMsg() got = %v, want %v", got, tt.want)
			}
		})
	}
}
