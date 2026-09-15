package hosts

import (
	"fmt"
	"os"

	"github.com/TacoContent/ironstate/sdk/handler"
)

type EntryHandler struct{}

func (EntryHandler) Emoji() string { return "📇" }

func (EntryHandler) RequiredTools() []string { return []string{} }

func (EntryHandler) Test(item map[string]any, _ string, ctx handler.Context) (bool, error) {
	if expression, ok := item["callback_condition"].(string); ok && expression != "" {
		if ctx.Callbacks == nil {
			return false, fmt.Errorf("host callbacks are required for callback_condition")
		}
		matches, err := ctx.Callbacks.EvaluateCondition(expression, ctx.Flat)
		if err != nil {
			return false, err
		}
		if !matches {
			return false, nil
		}
	}
	spec, err := parseSpec(item)
	if err != nil {
		return false, err
	}
	lines, err := readLines(spec.Path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return hasMapping(lines, spec.IP, spec.Hostname), nil
}

func (EntryHandler) Describe(item map[string]any, action handler.Action, ctx handler.Context) (string, error) {
	if template, ok := item["callback_template"].(string); ok && template != "" {
		if ctx.Callbacks == nil {
			return "", fmt.Errorf("host callbacks are required for callback_template")
		}
		return ctx.Callbacks.RenderTemplate(template, ctx.Flat)
	}
	spec, err := parseSpec(item)
	if err != nil {
		return "", err
	}
	verb := "add"
	if action == handler.ActionUninstall {
		verb = "remove"
	}
	return fmt.Sprintf("would %s hosts entry %s", verb, spec.Entry), nil
}

func (EntryHandler) Install(item map[string]any, _ string, ctx handler.Context) (handler.ExecResult, error) {
	spec, err := parseSpec(item)
	if err != nil {
		return failedResult(err), err
	}
	lines, err := readLines(spec.Path)
	if err != nil && !os.IsNotExist(err) {
		return failedResult(err), err
	}
	if hasMapping(lines, spec.IP, spec.Hostname) {
		return handler.ExecResult{RC: 0}, nil
	}
	lines = removeHostname(lines, spec.Hostname)
	lines = append(lines, spec.Entry)
	if spec.Comment != "" {
		lines = append(lines, "# "+spec.Comment)
	}
	if err := writeLines(spec.Path, lines, ctx.Become); err != nil {
		return failedResult(err), err
	}
	return resultFor(spec), nil
}

func (EntryHandler) Uninstall(item map[string]any, _ string, ctx handler.Context) (handler.ExecResult, error) {
	spec, err := parseSpec(item)
	if err != nil {
		return failedResult(err), err
	}
	lines, err := readLines(spec.Path)
	if os.IsNotExist(err) {
		return handler.ExecResult{RC: 0}, nil
	}
	if err != nil {
		return failedResult(err), err
	}
	filtered := removeMapping(lines, spec.IP, spec.Hostname, spec.Comment)
	if len(filtered) == len(lines) {
		return handler.ExecResult{RC: 0}, nil
	}
	if err := writeLines(spec.Path, filtered, ctx.Become); err != nil {
		return failedResult(err), err
	}
	return resultFor(spec), nil
}

func resultFor(spec hostSpec) handler.ExecResult {
	return handler.ExecResult{RC: 0, Extra: map[string]any{"value": map[string]any{"path": spec.Path, "ip": spec.IP, "hostname": spec.Hostname}}}
}

func failedResult(err error) handler.ExecResult {
	message := err.Error()
	return handler.ExecResult{RC: 1, Stderr: message, StderrLines: []string{message}}
}

var _ handler.Handler = EntryHandler{}
var _ handler.EmojiProvider = EntryHandler{}
