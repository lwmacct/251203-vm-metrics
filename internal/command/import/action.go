package importcmd

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/lwmacct/251203-mc-metrics/internal/command"
	"github.com/lwmacct/251203-mc-metrics/internal/vmapi"
	"github.com/urfave/cli/v3"
)

// getReader 获取输入 Reader
func getReader(cmd *cli.Command) (io.ReadCloser, error) {
	// 优先使用 --input 参数
	inputPath := cmd.String("input")
	if inputPath == "" {
		// 其次使用位置参数
		if cmd.Args().Len() > 0 {
			inputPath = cmd.Args().First()
		}
	}

	useGzip := cmd.Bool("gzip")

	var r io.ReadCloser
	if inputPath == "" || inputPath == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(inputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open input file: %w", err)
		}
		r = f
	}

	if useGzip {
		gr, err := gzip.NewReader(r)
		if err != nil {
			_ = r.Close()
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		return gr, nil
	}
	return r, nil
}

// actionImportJSON 导入 JSON Line 格式
func actionImportJSON(ctx context.Context, cmd *cli.Command) error {
	client, err := command.NewClient(command.GetConfig(cmd))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	r, err := getReader(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	importer, ok := client.(vmapi.Importer)
	if !ok {
		return fmt.Errorf("client does not support import")
	}

	return importer.ImportJSON(ctx, r)
}

// actionImportCSV 导入 CSV 格式
func actionImportCSV(ctx context.Context, cmd *cli.Command) error {
	client, err := command.NewClient(command.GetConfig(cmd))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	r, err := getReader(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	importer, ok := client.(vmapi.Importer)
	if !ok {
		return fmt.Errorf("client does not support import")
	}

	return importer.ImportCSV(ctx, r)
}

// actionImportNative 导入 Native 二进制格式
func actionImportNative(ctx context.Context, cmd *cli.Command) error {
	client, err := command.NewClient(command.GetConfig(cmd))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	r, err := getReader(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	importer, ok := client.(vmapi.Importer)
	if !ok {
		return fmt.Errorf("client does not support import")
	}

	return importer.ImportNative(ctx, r)
}

// actionImportPrometheus 导入 Prometheus exposition 格式
func actionImportPrometheus(ctx context.Context, cmd *cli.Command) error {
	client, err := command.NewClient(command.GetConfig(cmd))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	r, err := getReader(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	importer, ok := client.(vmapi.Importer)
	if !ok {
		return fmt.Errorf("client does not support import")
	}

	opts := &vmapi.ImportOptions{
		Job:      cmd.String("job"),
		Instance: cmd.String("instance"),
	}

	return importer.ImportPrometheus(ctx, r, opts)
}
