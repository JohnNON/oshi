package oshi_test

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/JohnNON/oshi"
	"github.com/stretchr/testify/assert"
)

func prepareData(t *testing.T) []byte {
	t.Helper()

	file, err := os.ReadFile("gopher.png")
	assert.NoError(t, err)

	return file
}

func Test_Upload(t *testing.T) {
	ctx := context.Background()

	client := oshi.NewClient(&http.Client{})

	img := oshi.NewImage(prepareData(t), "name_test", 5, true, false, false)

	wg := sync.WaitGroup{}
	wg.Add(5)

	resp, err := client.Upload(ctx, img)
	assert.NoError(t, err)

	assert.NotEmpty(t, resp.Admin)
	assert.NotEmpty(t, resp.Download)
	assert.NotEmpty(t, resp.TorDownload)
}

func Test_GetHashsum(t *testing.T) {
	ctx := context.Background()

	client := oshi.NewClient(&http.Client{})

	img := oshi.NewImage(prepareData(t), "name_test", 5, true, false, false)

	wg := sync.WaitGroup{}
	wg.Add(5)

	resp, err := client.Upload(ctx, img)
	assert.NoError(t, err)

	info, err := client.GetHashsum(ctx, strings.Split(resp.Download, "/")[3])
	assert.NoError(t, err)
	assert.NotEmpty(t, info.Algorithm)
	assert.NotEmpty(t, info.Hashsum)
}

func Test_Delete(t *testing.T) {
	ctx := context.Background()

	client := oshi.NewClient(&http.Client{})

	img := oshi.NewImage(prepareData(t), "name_test", 5, true, false, false)

	wg := sync.WaitGroup{}
	wg.Add(5)

	resp, err := client.Upload(ctx, img)
	assert.NoError(t, err)

	err = client.Delete(ctx, resp.Admin)
	assert.NoError(t, err)
}

func Test_GetTorEndpoint(t *testing.T) {
	ctx := context.Background()

	client := oshi.NewClient(&http.Client{})

	resp, err := client.GetTorEndpoint(ctx)
	assert.NoError(t, err)

	assert.NotEmpty(t, resp)
}
