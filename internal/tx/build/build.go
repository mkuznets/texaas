package build

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"mkuznets.com/go/texaas/internal/repo"
	"mkuznets.com/go/texaas/internal/tx"
)

type Command struct {
	BaseURL string `long:"base-url" env:"BASE_URL" default:"http://127.0.0.1:7777"`
	tx.Command
}

type buildData struct {
	Inputs []*repo.Input
}

type prepareResult struct {
	BuildID  uint64 `json:"build_id"`
	Uncashed []string
}

func (cmd *Command) Execute([]string) error {
	ctx := context.Background()

	r, er := repo.New("./")
	if er != nil {
		panic(er)
	}

	mf, er := r.Makefile("./tx.make.yml")
	if er != nil {
		panic(er)
	}

	data := &buildData{
		Inputs: mf.Inputs,
	}

	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(data); err != nil {
		return err
	}

	resp, err := http.Post(cmd.BaseURL+"/builds", "application/json", body)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf(buf.String())
	}

	result := &prepareResult{}
	if err := json.NewDecoder(buf).Decode(&result); err != nil {
		return err
	}

	hashes := make(map[string]struct{}, len(result.Uncashed))
	for _, h := range result.Uncashed {
		hashes[h] = struct{}{}
	}

	err = cmd.upload(ctx, func(writer *multipart.Writer) error {
		for _, input := range mf.Inputs {
			if _, ok := hashes[input.Hash]; !ok {
				continue
			}
			delete(hashes, input.Hash)

			err := func() error {
				f, err := os.Open(input.Path)
				if err != nil {
					return err
				}
				defer f.Close()

				part, err := writer.CreateFormFile("files", input.Hash)
				if err != nil {
					return err
				}

				gz := gzip.NewWriter(part)
				defer gz.Close()

				_, err = io.Copy(gz, f)
				if err != nil {
					return err
				}

				return nil
			}()
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/builds/%d/start", cmd.BaseURL, result.BuildID)

	resp, err = http.Post(url, "application/json", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		buf := bytes.NewBuffer(nil)
		buf.ReadFrom(resp.Body)
		return fmt.Errorf("the server responded with a status %d: %v", resp.StatusCode, buf.String())
	}

	return nil
}

func (cmd *Command) upload(ctx context.Context, fn func(*multipart.Writer) error) error {
	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)

	errChan := make(chan error, 1)
	go func() {
		defer pipeWriter.Close()

		if err := fn(writer); err != nil {
			errChan <- err
			return
		}
		if err := writer.Close(); err != nil {
			errChan <- err
			return
		}

		close(errChan)
	}()

	uploadURL := cmd.BaseURL + "/upload"
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, pipeReader)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Body = ioutil.NopCloser(pipeReader)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed doing request: %v", err)
	}
	defer resp.Body.Close()

	// Handling the error the routine may caused
	if err := <-errChan; err != nil {
		return err
	}

	if resp.StatusCode != 201 {
		buf := bytes.NewBuffer(nil)
		buf.ReadFrom(resp.Body)
		return fmt.Errorf("the server responded with a status %d: %v", resp.StatusCode, buf.String())
	}

	return nil
}
