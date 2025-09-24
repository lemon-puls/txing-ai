package tool

import (
	"context"
	"testing"
)

func Test_readPdfText(t *testing.T) {
	type args struct {
		ctx    context.Context
		params *pdfReadParams
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test pdf read",
			args: args{
				ctx:    context.Background(),
				params: &pdfReadParams{FilePath: "runtime/lzw_resume.pdf"},
			},
			want:    "This is a test PDF file.",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readPdfText(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("readPdfText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("readPdfText() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validatePdf(t *testing.T) {
	type args struct {
		ctx    context.Context
		params *pdfValidateParams
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
			got, err := validatePdf(tt.args.ctx, tt.args.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePdf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validatePdf() got = %v, want %v", got, tt.want)
			}
		})
	}
}
