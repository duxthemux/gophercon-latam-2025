package tool

import (
	"context"
	"os/exec"
	"strings"
)

type hostnameTool struct{}

func (h hostnameTool) Query(ctx context.Context, params map[string]string) (ret string, err error) {
	bs, err := exec.CommandContext(ctx, "hostname").CombinedOutput()
	if err != nil {
		return "", err
	}

	return "Seu hostname é: " + strings.TrimSpace(string(bs)), nil
}

type ifconfigTool struct{}

func (h ifconfigTool) Query(ctx context.Context, params map[string]string) (ret string, err error) {
	bs, err := exec.CommandContext(ctx, "ifconfig").CombinedOutput()
	if err != nil {
		return "", err
	}

	return "As configurações de rede e ip são: " + strings.TrimSpace(string(bs)), nil
}

type dateTool struct{}

func (h dateTool) Query(ctx context.Context, params map[string]string) (ret string, err error) {
	bs, err := exec.CommandContext(ctx, "date").CombinedOutput()
	if err != nil {
		return "", err
	}

	return "A data e hora atual é: " + strings.TrimSpace(string(bs)), nil
}

type diskFreeTool struct{}

func (h diskFreeTool) Query(ctx context.Context, params map[string]string) (ret string, err error) {
	bs, err := exec.CommandContext(ctx, "df", "-h").CombinedOutput()
	if err != nil {
		return "", err
	}

	return "As informações sobre disco livre são: " + strings.TrimSpace(string(bs)), nil
}
