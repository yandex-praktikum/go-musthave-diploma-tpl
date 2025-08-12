package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/kdv2001/onlyMetrics/internal/domain"
)

func TestClient_send(t *testing.T) {
	t.Parallel()
	type fields struct {
		client    httpClient
		serverURL url.URL
	}
	type args struct {
		ctx   context.Context
		value domain.MetricValue
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: fields{
				client: &httpClientMock{
					response: http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					},
				},
			},
			args: args{
				ctx: context.Background(),
				value: domain.MetricValue{
					Type:       domain.GaugeMetricType,
					Name:       "some metric",
					GaugeValue: 100,
				},
			},
			wantErr: false,
		},
		{
			name: "success",
			fields: fields{
				client: &httpClientMock{
					response: http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					},
				},
			},
			args: args{
				ctx: context.Background(),
				value: domain.MetricValue{
					Type:       domain.CounterMetricType,
					Name:       "some metric",
					GaugeValue: 100,
				},
			},
			wantErr: false,
		},
		{
			name: "unknown metric type",
			fields: fields{
				client: &httpClientMock{
					response: http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					},
				},
			},
			args: args{
				ctx: context.Background(),
				value: domain.MetricValue{
					Type:       "unknown metric type",
					Name:       "some metric",
					GaugeValue: 100,
				},
			},
			wantErr: true,
		},
		{
			name: "internal server error",
			fields: fields{
				client: &httpClientMock{
					response: http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
					},
				},
			},
			args: args{
				ctx: context.Background(),
				value: domain.MetricValue{
					Type:       domain.GaugeMetricType,
					Name:       "some metric",
					GaugeValue: 100,
				},
			},
			wantErr: true,
		},
		{
			name: "do err",
			fields: fields{
				client: &httpClientMock{
					err:      errors.New("some error"),
					response: http.Response{},
				},
			},
			args: args{
				ctx: context.Background(),
				value: domain.MetricValue{
					Type:       domain.GaugeMetricType,
					Name:       "some metric",
					GaugeValue: 100,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{
				client:    tt.fields.client,
				serverURL: tt.fields.serverURL,
			}
			if err := c.send(tt.args.ctx, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
