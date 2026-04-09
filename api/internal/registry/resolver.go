package registry

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/gobwas/glob"
)

type compiledEntry struct {
	glob  glob.Glob
	entry ImageEntry
}

type Resolver struct {
	entries []compiledEntry
}

func NewResolver(reg *ImageRegistry) (*Resolver, error) {
	entries := make([]compiledEntry, 0, len(reg.Images))
	for _, e := range reg.Images {
		g, err := glob.Compile(e.Match)
		if err != nil {
			return nil, fmt.Errorf("compile glob %q: %w", e.Match, err)
		}
		entries = append(entries, compiledEntry{glob: g, entry: e})
	}
	return &Resolver{entries: entries}, nil
}

func (r *Resolver) Resolve(imageName string, ctx TemplateContext) (*ResolvedImage, error) {
	nameOnly := stripTag(imageName)

	for _, ce := range r.entries {
		if !ce.glob.Match(imageName) && !ce.glob.Match(nameOnly) {
			continue
		}
		return renderEntry(ce.entry, ctx)
	}

	return &ResolvedImage{}, nil
}

func (r *Resolver) ResolveEntry(imageName string) *ImageEntry {
	nameOnly := stripTag(imageName)
	for _, ce := range r.entries {
		if ce.glob.Match(imageName) || ce.glob.Match(nameOnly) {
			return &ce.entry
		}
	}
	return nil
}

func stripTag(image string) string {
	if strings.Contains(image, "@") {
		return image
	}
	for i := len(image) - 1; i >= 0; i-- {
		if image[i] == ':' {
			return image[:i]
		}
		if image[i] == '/' {
			break
		}
	}
	return image
}

func renderEntry(entry ImageEntry, ctx TemplateContext) (*ResolvedImage, error) {
	if ctx.Meta == nil {
		ctx.Meta = make(map[string]string)
	}
	for _, m := range entry.Metadata {
		if _, ok := ctx.Meta[m.Key]; !ok && m.Value != "" {
			ctx.Meta[m.Key] = m.Value
		}
	}

	if entry.SSH != nil {
		ctx.SSHPort = strconv.Itoa(entry.SSH.Port)
		ctx.SSHUsername = entry.SSH.Username
		ctx.SSHPassword = entry.SSH.Password
	}

	resolved := &ResolvedImage{
		HealthCheck: entry.HealthCheck,
	}

	if entry.InternalPort != nil {
		resolved.InternalPort = *entry.InternalPort
	}

	keys := make([]string, 0, len(entry.Env))
	for k := range entry.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		rendered, err := renderTemplate(k+"="+entry.Env[k], ctx)
		if err != nil {
			return nil, fmt.Errorf("render env %s: %w", k, err)
		}
		resolved.Env = append(resolved.Env, rendered)
	}

	if len(entry.Labels) > 0 {
		resolved.Labels = make(map[string]string, len(entry.Labels))
		for k, v := range entry.Labels {
			rendered, err := renderTemplate(v, ctx)
			if err != nil {
				return nil, fmt.Errorf("render label %s: %w", k, err)
			}
			resolved.Labels[k] = rendered
		}
	}

	for _, cmd := range entry.PostStart {
		rendered, err := renderExecCommand(cmd, ctx)
		if err != nil {
			return nil, fmt.Errorf("render post_start command: %w", err)
		}
		resolved.PostStart = append(resolved.PostStart, rendered)
	}

	for _, cmd := range entry.PreStop {
		rendered, err := renderExecCommand(cmd, ctx)
		if err != nil {
			return nil, fmt.Errorf("render pre_stop command: %w", err)
		}
		resolved.PreStop = append(resolved.PreStop, rendered)
	}

	resolved.SSH = entry.SSH

	for _, ls := range entry.Logs {
		rendered := LogSource{
			Key:   ls.Key,
			Label: ls.Label,
			Type:  ls.Type,
		}
		if ls.Path != "" {
			p, err := renderTemplate(ls.Path, ctx)
			if err != nil {
				return nil, fmt.Errorf("render log path %s: %w", ls.Key, err)
			}
			rendered.Path = p
		}
		resolved.Logs = append(resolved.Logs, rendered)
	}

	return resolved, nil
}

func renderExecCommand(cmd ExecCommand, ctx TemplateContext) (ExecCommand, error) {
	rendered := ExecCommand{
		Delay:      cmd.Delay,
		Timeout:    cmd.Timeout,
		Retries:    cmd.Retries,
		RetryDelay: cmd.RetryDelay,
	}
	for _, arg := range cmd.Command {
		r, err := renderTemplate(arg, ctx)
		if err != nil {
			return rendered, err
		}
		rendered.Command = append(rendered.Command, r)
	}
	return rendered, nil
}

func renderTemplate(text string, ctx TemplateContext) (string, error) {
	tmpl, err := template.New("").Option("missingkey=error").Parse(text)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
}
