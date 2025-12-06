package export

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/lwmacct/251203-vm-metrics/internal/command"
	"github.com/lwmacct/251203-vm-metrics/internal/vmapi"
	"github.com/urfave/cli/v3"
)

// getWriter 获取输出 Writer
func getWriter(cmd *cli.Command) (io.WriteCloser, error) {
	outputPath := cmd.String("output")
	useGzip := cmd.Bool("gzip")

	var w io.WriteCloser
	if outputPath == "" || outputPath == "-" {
		w = os.Stdout
	} else {
		f, err := os.Create(outputPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create output file: %w", err)
		}
		w = f
	}

	if useGzip {
		return gzip.NewWriter(w), nil
	}
	return w, nil
}

// buildExportOptions 构建导出选项
func buildExportOptions(cmd *cli.Command) (*vmapi.ExportOptions, error) {
	match := cmd.Args().Slice()
	if len(match) == 0 {
		return nil, fmt.Errorf("at least one match selector is required")
	}

	start, err := command.ParseTime(cmd.String("start"))
	if err != nil {
		return nil, err
	}
	end, err := command.ParseTime(cmd.String("end"))
	if err != nil {
		return nil, err
	}

	return &vmapi.ExportOptions{
		Match:          match,
		Start:          start,
		End:            end,
		MaxRowsPerLine: cmd.Int("max-rows-per-line"),
		CSVFormat:      cmd.String("csv-format"),
		ReduceMemUsage: cmd.Bool("reduce-mem-usage"),
	}, nil
}

// actionExportJSON 导出 JSON Line 格式
func actionExportJSON(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() == 0 {
		return cli.ShowAppHelp(cmd)
	}

	client, err := command.NewClient(command.GetConfig(cmd))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	opts, err := buildExportOptions(cmd)
	if err != nil {
		return err
	}

	w, err := getWriter(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()

	exporter, ok := client.(vmapi.Exporter)
	if !ok {
		return fmt.Errorf("client does not support export")
	}

	return exporter.ExportJSON(ctx, w, opts)
}

// actionExportCSV 导出 CSV 格式
func actionExportCSV(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() == 0 {
		return cli.ShowAppHelp(cmd)
	}

	client, err := command.NewClient(command.GetConfig(cmd))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	opts, err := buildExportOptions(cmd)
	if err != nil {
		return err
	}

	w, err := getWriter(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()

	exporter, ok := client.(vmapi.Exporter)
	if !ok {
		return fmt.Errorf("client does not support export")
	}

	return exporter.ExportCSV(ctx, w, opts)
}

// actionExportNative 导出 Native 二进制格式
func actionExportNative(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() == 0 {
		return cli.ShowAppHelp(cmd)
	}

	client, err := command.NewClient(command.GetConfig(cmd))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	opts, err := buildExportOptions(cmd)
	if err != nil {
		return err
	}

	w, err := getWriter(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()

	exporter, ok := client.(vmapi.Exporter)
	if !ok {
		return fmt.Errorf("client does not support export")
	}

	return exporter.ExportNative(ctx, w, opts)
}
