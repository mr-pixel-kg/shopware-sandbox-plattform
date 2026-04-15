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
	for _, it := range entry.Metadata.Items {
		if it.Type != "field" || it.Field == nil || it.Field.Default == "" {
			continue
		}
		if _, ok := ctx.Meta[it.Key]; !ok {
			ctx.Meta[it.Key] = it.Field.Default
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
		Label:      cmd.Label,
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

func (r *Resolver) RenderMetadata(imageName string, values map[string]string, metadata []MetadataItem, ctx TemplateContext) ([]MetadataItem, error) {
	return RenderMetadata(r.SchemaFor(imageName), values, metadata, ctx)
}

func (r *Resolver) SchemaFor(imageName string) *MetadataSchema {
	if r == nil {
		return nil
	}
	entry := r.ResolveEntry(imageName)
	if entry == nil {
		return nil
	}
	return &entry.Metadata
}

func (r *Resolver) MergeMetadata(imageName string, metadata []MetadataItem) []MetadataItem {
	return MergeWithRegistry(r.SchemaFor(imageName), metadata)
}

func RenderMetadata(registrySchema *MetadataSchema, values map[string]string, metadata []MetadataItem, ctx TemplateContext) ([]MetadataItem, error) {
	merged := MergeWithRegistry(registrySchema, metadata)

	meta := make(map[string]string, len(merged))
	for _, it := range merged {
		if it.Type == "field" && it.Field != nil {
			meta[it.Key] = it.Field.Default
		}
	}
	for k, v := range values {
		if _, ok := meta[k]; ok {
			meta[k] = v
		}
	}
	ctx.Meta = meta

	out := make([]MetadataItem, len(merged))
	for i, it := range merged {
		clone := it
		switch it.Type {
		case "field":
			if it.Field != nil {
				f := *it.Field
				if v, ok := values[it.Key]; ok {
					f.Default = v
				}
				clone.Field = &f
			}
		case "action":
			if it.Action != nil {
				a := *it.Action
				rendered, err := renderInlineTemplate(it.Action.urlTmpl, a.URL, ctx)
				if err != nil {
					return nil, fmt.Errorf("render action.url for %q: %w", it.Key, err)
				}
				a.URL = rendered
				clone.Action = &a
			}
		case "display":
			if it.Display != nil {
				d := *it.Display
				rendered, err := renderInlineTemplate(it.Display.valueTmpl, d.Value, ctx)
				if err != nil {
					return nil, fmt.Errorf("render display.value for %q: %w", it.Key, err)
				}
				d.Value = rendered
				clone.Display = &d
			}
		}
		out[i] = clone
	}

	return out, nil
}

func renderInlineTemplate(cached *template.Template, source string, ctx TemplateContext) (string, error) {
	if cached != nil {
		return execTemplate(cached, ctx)
	}
	if source == "" {
		return "", nil
	}
	return renderTemplate(source, ctx)
}

func execTemplate(t *template.Template, ctx TemplateContext) (string, error) {
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", err
	}
	return buf.String(), nil
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

func MergeWithRegistry(registrySchema *MetadataSchema, userItems []MetadataItem) []MetadataItem {
	userByKey := map[string]MetadataItem{}
	for _, it := range userItems {
		userByKey[it.Key] = it
	}
	var registryItems []MetadataItem
	if registrySchema != nil {
		registryItems = registrySchema.Items
	}
	out := make([]MetadataItem, 0, len(registryItems)+len(userItems))
	for _, it := range registryItems {
		if patch, ok := userByKey[it.Key]; ok {
			out = append(out, mergeItem(it, patch))
			delete(userByKey, it.Key)
			continue
		}
		out = append(out, it)
	}
	for _, it := range userItems {
		if _, ok := userByKey[it.Key]; ok {
			out = append(out, it)
		}
	}
	return out
}

func StripRegistryDuplicates(userItems []MetadataItem, registrySchema *MetadataSchema) []MetadataItem {
	registryByKey := map[string]MetadataItem{}
	if registrySchema != nil {
		for _, it := range registrySchema.Items {
			registryByKey[it.Key] = it
		}
	}
	out := make([]MetadataItem, 0, len(userItems))
	for _, it := range userItems {
		ref, hasMatch := registryByKey[it.Key]
		if !hasMatch {
			out = append(out, it)
			continue
		}
		if patch, ok := extractOverride(it, ref); ok {
			out = append(out, patch)
		}
	}
	return out
}

func extractOverride(user, ref MetadataItem) (MetadataItem, bool) {
	if user.Type != "field" || user.Field == nil || ref.Field == nil {
		return MetadataItem{}, false
	}
	if user.Field.Default == ref.Field.Default {
		return MetadataItem{}, false
	}
	return MetadataItem{
		Key:   user.Key,
		Type:  "field",
		Field: &FieldSpec{Default: user.Field.Default},
	}, true
}

func mergeItem(base, patch MetadataItem) MetadataItem {
	if base.Type != "field" || base.Field == nil || patch.Field == nil {
		return base
	}
	merged := *base.Field
	merged.Default = patch.Field.Default
	out := base
	out.Field = &merged
	return out
}
