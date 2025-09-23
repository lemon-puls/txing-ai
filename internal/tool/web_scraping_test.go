package tool

import (
	"context"
	"reflect"
	"testing"
)

func Test_scrapeWebPage(t *testing.T) {
	type args struct {
		ctx context.Context
		req *webScrapingRequest
	}
	tests := []struct {
		name    string
		args    args
		want    webScrapingResponse
		wantErr bool
	}{
		{
			name: "成功抓取网页",
			args: args{
				ctx: context.Background(),
				req: &webScrapingRequest{URL: "https://baijiahao.baidu.com/s?id=1843853053263796289&wfr=spider&for=pc"}},
			want: webScrapingResponse{
				Title:   "百度百科 - 百度百科，你就在这里",
				Content: "百度百科，你就在这里",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := scrapeWebPage(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("scrapeWebPage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scrapeWebPage() got = %v, want %v", got, tt.want)
			}
		})
	}
}
