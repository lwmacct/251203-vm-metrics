// Package completion 提供 shell 补全脚本生成
package completion

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

// Command completion 命令
var Command = &cli.Command{
	Name:  "completion",
	Usage: "生成 shell 补全脚本",
	Commands: []*cli.Command{
		bashCommand,
		zshCommand,
	},
}

var bashCommand = &cli.Command{
	Name:  "bash",
	Usage: "生成 bash 补全脚本",
	Description: `生成 bash 补全脚本。

启用补全:

  # 临时生效 (当前会话)
  source <(mc-metrics completion bash)

  # 永久生效 (添加到 ~/.bashrc)
  mc-metrics completion bash >> ~/.bashrc

  # 系统级安装 (需要 root)
  mc-metrics completion bash | sudo tee /etc/bash_completion.d/mc-metrics
`,
	Action: func(ctx context.Context, cmd *cli.Command) error {
		return generateBashCompletion(os.Stdout)
	},
}

var zshCommand = &cli.Command{
	Name:  "zsh",
	Usage: "生成 zsh 补全脚本",
	Description: `生成 zsh 补全脚本。

启用补全:

  # 确保 completions 目录在 fpath 中
  echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
  echo 'autoload -Uz compinit && compinit' >> ~/.zshrc

  # 生成补全脚本
  mkdir -p ~/.zsh/completions
  mc-metrics completion zsh > ~/.zsh/completions/_mc-metrics

  # 重新加载 zsh
  exec zsh
`,
	Action: func(ctx context.Context, cmd *cli.Command) error {
		return generateZshCompletion(os.Stdout)
	},
}

// generateBashCompletion 和 generateZshCompletion 在各自的文件中实现

func init() {
	// 添加示例提示
	if Command.Description == "" {
		Command.Description = fmt.Sprintf(`生成 shell 补全脚本，支持 bash 和 zsh。

示例:
  %s completion bash    # 输出 bash 补全脚本
  %s completion zsh     # 输出 zsh 补全脚本
`, "mc-metrics", "mc-metrics")
	}
}
