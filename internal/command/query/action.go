package query

import (
	"context"
	"fmt"
	"time"

	"github.com/lwmacct/251203-mc-metrics/internal/command"
	"github.com/lwmacct/251203-mc-metrics/internal/config"
	"github.com/lwmacct/251203-mc-metrics/internal/output"
	"github.com/urfave/cli/v3"
)

// actionQuery 执行查询
func actionQuery(ctx context.Context, cmd *cli.Command) error {
	query := cmd.Args().First()
	if query == "" {
		return cli.ShowAppHelp(cmd)
	}

	cfg := command.GetConfig(cmd)
	client, err := command.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// 解析时间
	ts, err := command.ParseTime(cmd.String("time"))
	if err != nil {
		return fmt.Errorf("invalid time format: %w", err)
	}

	// 判断是 Instant 还是 Range 查询
	rangeDuration := cmd.Duration("range")
	if rangeDuration > 0 {
		// Range Query
		end := ts
		start := end.Add(-rangeDuration)
		step := cmd.Duration("step")

		result, err := client.QueryRange(ctx, query, start, end, step)
		if err != nil {
			return err
		}
		w, err := newWriter(cfg)
		if err != nil {
			return err
		}
		return w.WriteQueryResult(result)
	}

	// Instant Query
	result, err := client.Query(ctx, query, ts)
	if err != nil {
		return err
	}

	w, err := newWriter(cfg)
	if err != nil {
		return err
	}
	return w.WriteQueryResult(result)
}

// actionMetrics 列出所有指标名称
func actionMetrics(ctx context.Context, cmd *cli.Command) error {
	cfg := command.GetConfig(cmd)
	client, err := command.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// 获取 __name__ 标签的所有值
	result, err := client.LabelValues(ctx, "__name__", time.Time{}, time.Time{})
	if err != nil {
		return err
	}

	w, err := newWriter(cfg)
	if err != nil {
		return err
	}
	return w.WriteStrings(result.Values)
}

// actionLabels 列出所有标签名称
func actionLabels(ctx context.Context, cmd *cli.Command) error {
	cfg := command.GetConfig(cmd)
	client, err := command.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	result, err := client.Labels(ctx, time.Time{}, time.Time{})
	if err != nil {
		return err
	}

	w, err := newWriter(cfg)
	if err != nil {
		return err
	}
	return w.WriteStrings(result.Labels)
}

// actionLabelValues 获取指定标签的所有值
func actionLabelValues(ctx context.Context, cmd *cli.Command) error {
	label := cmd.Args().First()
	if label == "" {
		return fmt.Errorf("label name is required")
	}

	cfg := command.GetConfig(cmd)
	client, err := command.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	result, err := client.LabelValues(ctx, label, time.Time{}, time.Time{})
	if err != nil {
		return err
	}

	w, err := newWriter(cfg)
	if err != nil {
		return err
	}
	return w.WriteStrings(result.Values)
}

// actionSeries 列出匹配的时间序列
func actionSeries(ctx context.Context, cmd *cli.Command) error {
	match := cmd.Args().First()
	if match == "" {
		return fmt.Errorf("match selector is required")
	}

	cfg := command.GetConfig(cmd)
	client, err := command.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	result, err := client.Series(ctx, []string{match}, time.Time{}, time.Time{})
	if err != nil {
		return err
	}

	w, err := newWriter(cfg)
	if err != nil {
		return err
	}
	return w.WriteSeries(result.Series)
}

// newWriter 创建输出 Writer
func newWriter(cfg *config.Config) (output.Writer, error) {
	return output.New(cfg.Output.Format, output.Options{
		NoHeaders: cfg.Output.NoHeaders,
	})
}
