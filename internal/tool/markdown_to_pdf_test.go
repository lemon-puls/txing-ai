package tool

import (
	"context"
	"reflect"
	"testing"
)

func Test_cleanupTempImages(t *testing.T) {
	type args struct {
		imagePaths []string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTempImages(tt.args.imagePaths)
		})
	}
}

func Test_convertHTMLToPDF(t *testing.T) {
	type args struct {
		htmlPath string
		pdfPath  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := convertHTMLToPDF(tt.args.htmlPath, tt.args.pdfPath); (err != nil) != tt.wantErr {
				t.Errorf("convertHTMLToPDF() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_downloadImageFromURL(t *testing.T) {
	type args struct {
		url string
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := downloadImageFromURL(tt.args.url, tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("downloadImageFromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("downloadImageFromURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_markdownToHTML(t *testing.T) {
	type args struct {
		content string
		title   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := markdownToHTML(tt.args.content, tt.args.title); got != tt.want {
				t.Errorf("markdownToHTML() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processMarkdownImages(t *testing.T) {
	type args struct {
		content string
		tempDir string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := processMarkdownImages(tt.args.content, tt.args.tempDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("processMarkdownImages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("processMarkdownImages() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("processMarkdownImages() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_saveMarkdownToPDF(t *testing.T) {
	type args struct {
		ctx    context.Context
		params *markdownToPDFParams
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test_saveMarkdownToPDF",
			args: args{
				ctx: context.Background(),
				params: &markdownToPDFParams{
					Content:  "# 深圳旅游攻略\n\n## 一、城市概览\n\n深圳，位于中国广东省南部，是一座充满活力的现代化大都市。作为中国改革开放的前沿阵地，深圳以其飞速的发展、创新的精神和多元的文化吸引着来自世界各地的游客。这里不仅有摩天大楼林立的城市天际线，也有山海相依的自然风光，更有融合了各地风味的美食江湖。\n\n![深圳地标建筑](shenzhen_landmark.jpg)\n\n## 二、必游景点推荐\n\n### 1. 深圳湾公园\n\n深圳湾公园是市民和游客休闲放松的好去处。这里有长达13公里的海滨长廊，可以骑行、散步，远眺对岸的香港。每年冬季，还能在这里观赏到成群的候鸟。\n\n### 2. 世界之窗\n\n世界之窗是一个集世界著名景观微缩版于一体的大型主题公园。在这里，你可以一天之内“环游世界”，从埃菲尔铁塔到自由女神像，从金字塔到比萨斜塔，应有尽有。\n\n### 3. 欢乐谷\n\n欢乐谷是年轻人和家庭游客的乐园，拥有众多刺激的游乐设施和精彩的表演项目。无论是过山车爱好者还是寻求亲子乐趣的家庭，都能在这里找到属于自己的快乐。\n\n### 4. 大鹏所城\n\n大鹏所城是深圳保存最完好的明清海防军事古城，有着600多年的历史。漫步在古老的街巷中，感受历史的厚重与沧桑。\n\n### 5. 东部华侨城\n\n东部华侨城是一个集生态旅游、娱乐休闲、度假养生于一体的综合性旅游区。这里有茶溪谷的田园风光，也有大侠谷的惊险刺激。\n\n## 三、地道美食推荐\n\n深圳是一座移民城市，汇聚了全国各地乃至世界各地的美食。以下是一些不容错过的地道美味：\n\n- **潮汕牛肉火锅**：选用新鲜现切的黄牛肉，肉质鲜嫩，搭配沙茶酱或普宁豆酱，回味无穷。\n- **光明乳鸽**：皮脆肉嫩，香气扑鼻，是深圳本地的传统名菜。\n- **沙井蚝**：产自深圳沙井，个大肥美，可炭烤、可生吃，鲜味十足。\n- **南澳鲍鱼**：出自深圳南澳海域，肉质滑爽脆嫩，营养丰富。\n- **公明烧鹅**：皮脆肉嫩，肥而不腻，搭配酸梅酱食用，酸甜可口，解腻开胃。\n\n\n![深圳美食](shenzhen_food.jpg)\n\n## 四、交通出行指南\n\n深圳的公共交通系统非常发达，出行十分便利。\n\n### 1. 地铁\n\n深圳地铁网络覆盖全市，线路密集，是市内出行的最佳选择。支持扫码乘车（微信/支付宝小程序“深圳地铁”）或使用“深圳通”卡。\n\n### 2. 公交车\n\n公交车线路遍布各个角落，票价便宜，适合短途出行。同样支持扫码或刷卡支付。\n\n### 3. 出租车/网约车\n\n出租车和网约车（如滴滴出行）在深圳非常普及，方便快捷。建议使用网约车，价格透明，服务规范。\n\n### 4. 高铁/火车\n\n深圳拥有多个火车站，其中深圳北站是华南地区重要的交通枢纽，连接全国各大城市。\n\n### 5. 机场\n\n深圳宝安国际机场是深圳的主要航空门户，开通了众多国内外航线。\n\n## 五、住宿推荐\n\n根据不同的需求和预算，深圳提供了多样化的住宿选择：\n\n- **福田/罗湖中心区**：交通便利，靠近商业区和景点，适合商务出行和初次游客。推荐酒店：星辰酒店(深圳中心公园店)、深圳1982精品酒店。\n- **南山科技园/后海片区**：现代化程度高，靠近深圳湾，环境优美，适合追求品质的游客。\n- **大鹏新区/盐田区**：靠近海边和自然景区，适合度假休闲。可以选择海边民宿或度假酒店。\n\n## 六、旅行小贴士\n\n- 深圳气候温暖湿润，夏季较长，请注意防晒和防蚊。\n- 市民普遍使用手机支付，建议提前开通微信支付或支付宝。\n- 深圳是一座文明城市，请遵守公共秩序，爱护环境。\n- 部分景点需要提前预约，请出行前查询官方信息。\n\n祝您在深圳度过一段愉快的旅程！",
					Filename: "test_pdf",
				},
			},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := saveMarkdownToPDF(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("saveMarkdownToPDF() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("saveMarkdownToPDF() got = %v, want %v", got, tt.want)
			}
		})
	}
}
