package registry

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

var keyRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type pathErrors struct {
	errs []string
}

func (p *pathErrors) at(path, format string, args ...any) {
	p.errs = append(p.errs, fmt.Sprintf("%s: %s", path, fmt.Sprintf(format, args...)))
}

func (p *pathErrors) err() error {
	if len(p.errs) == 0 {
		return nil
	}
	return fmt.Errorf("registry validation failed:\n  - %s", strings.Join(p.errs, "\n  - "))
}

func validateRegistry(reg *ImageRegistry) error {
	pe := &pathErrors{}
	if len(reg.Images) == 0 {
		pe.at("registry", "must contain at least one image entry")
		return pe.err()
	}

	for i := range reg.Images {
		img := &reg.Images[i]
		base := fmt.Sprintf("images[%d]", i)

		if img.Match == "" {
			pe.at(base, "match pattern is required")
		}

		for j, cmd := range img.PostStart {
			if len(cmd.Command) == 0 {
				pe.at(fmt.Sprintf("%s.post_start[%d]", base, j), "command must not be empty")
			}
		}
		for j, cmd := range img.PreStop {
			if len(cmd.Command) == 0 {
				pe.at(fmt.Sprintf("%s.pre_stop[%d]", base, j), "command must not be empty")
			}
		}

		validateMetadata(base+".metadata", &img.Metadata, pe)
		validateLogs(base, img.Logs, pe)
	}

	return pe.err()
}

func ValidateMetadata(metadata []MetadataItem, registrySchema *MetadataSchema) error {
	pe := &pathErrors{}
	groupKeys := map[string]bool{}
	if registrySchema != nil {
		for _, g := range registrySchema.Groups {
			groupKeys[g.Key] = true
		}
	}

	fieldOptions := map[string][]string{}
	seenItems := map[string]bool{}
	for i := range metadata {
		it := &metadata[i]
		if it.Key != "" && keyRE.MatchString(it.Key) && it.Type == "field" && it.Field != nil {
			if it.Field.Input == "select" || it.Field.Input == "multiselect" {
				vals := make([]string, 0, len(it.Field.Options))
				for _, o := range it.Field.Options {
					vals = append(vals, o.Value)
				}
				fieldOptions[it.Key] = vals
			} else {
				fieldOptions[it.Key] = nil
			}
		}
	}

	for i := range metadata {
		it := &metadata[i]
		p := fmt.Sprintf("metadata[%d] (key=%q)", i, it.Key)
		if it.Key == "" || !keyRE.MatchString(it.Key) {
			pe.at(p, "item key must match ^[a-z][a-z0-9_]{0,62}$")
		}
		if seenItems[it.Key] {
			pe.at(p, "duplicate item key %q", it.Key)
		}
		seenItems[it.Key] = true
		if it.Label == "" {
			pe.at(p, "label is required")
		}
		if !ValidItemTypes[it.Type] {
			pe.at(p, "type must be one of field|action|display, got %q", it.Type)
		}
		setCount := 0
		if it.Field != nil {
			setCount++
		}
		if it.Action != nil {
			setCount++
		}
		if it.Display != nil {
			setCount++
		}
		if setCount != 1 {
			pe.at(p, "exactly one of field/action/display must be set (found %d)", setCount)
		}
		if it.Type == "field" && it.Field == nil {
			pe.at(p, "type=field requires field sub-object")
		}
		if it.Type == "action" && it.Action == nil {
			pe.at(p, "type=action requires action sub-object")
		}
		if it.Type == "display" && it.Display == nil {
			pe.at(p, "type=display requires display sub-object")
		}
		if it.Group != "" && !groupKeys[it.Group] {
			pe.at(p, "group %q is not declared in the registry schema", it.Group)
		}
		validateVisibility(p, it.Visibility, fieldOptions, pe)
		if it.Field != nil {
			validateFieldSpec(p+".field", it.Field, pe)
		}
		if it.Action != nil {
			validateActionSpec(p+".action", it.Action, pe)
		}
		if it.Display != nil {
			validateDisplaySpec(p+".display", it.Display, pe)
		}
	}
	return pe.err()
}

func validateMetadata(base string, schema *MetadataSchema, pe *pathErrors) {
	groupKeys := map[string]bool{}
	seenGroups := map[string]bool{}
	for i, g := range schema.Groups {
		p := fmt.Sprintf("%s.groups[%d] (key=%q)", base, i, g.Key)
		if g.Key == "" || !keyRE.MatchString(g.Key) {
			pe.at(p, "group key must match ^[a-z][a-z0-9_]{0,62}$")
			continue
		}
		if g.Key == "_default" {
			pe.at(p, "group key %q is reserved", g.Key)
		}
		if g.Label == "" {
			pe.at(p, "group label is required")
		}
		if seenGroups[g.Key] {
			pe.at(p, "duplicate group key %q", g.Key)
		}
		seenGroups[g.Key] = true
		groupKeys[g.Key] = true
	}

	fieldOptions := map[string][]string{}
	seenItems := map[string]bool{}
	for i := range schema.Items {
		it := &schema.Items[i]
		if it.Key != "" && keyRE.MatchString(it.Key) && it.Type == "field" && it.Field != nil {
			if it.Field.Input == "select" || it.Field.Input == "multiselect" {
				vals := make([]string, 0, len(it.Field.Options))
				for _, o := range it.Field.Options {
					vals = append(vals, o.Value)
				}
				fieldOptions[it.Key] = vals
			} else {
				fieldOptions[it.Key] = nil
			}
		}
	}

	for i := range schema.Items {
		it := &schema.Items[i]
		p := fmt.Sprintf("%s.items[%d] (key=%q)", base, i, it.Key)

		if it.Key == "" || !keyRE.MatchString(it.Key) {
			pe.at(p, "item key must match ^[a-z][a-z0-9_]{0,62}$")
		}
		if it.Label == "" {
			pe.at(p, "label is required")
		}
		if seenItems[it.Key] {
			pe.at(p, "duplicate item key %q", it.Key)
		}
		seenItems[it.Key] = true

		if !ValidItemTypes[it.Type] {
			pe.at(p, "type must be one of field|action|display, got %q", it.Type)
		}

		setCount := 0
		if it.Field != nil {
			setCount++
		}
		if it.Action != nil {
			setCount++
		}
		if it.Display != nil {
			setCount++
		}
		if setCount != 1 {
			pe.at(p, "exactly one of field/action/display must be set (found %d)", setCount)
		}
		if it.Type == "field" && it.Field == nil {
			pe.at(p, "type=field requires field sub-object")
		}
		if it.Type == "action" && it.Action == nil {
			pe.at(p, "type=action requires action sub-object")
		}
		if it.Type == "display" && it.Display == nil {
			pe.at(p, "type=display requires display sub-object")
		}

		if it.Group != "" && !groupKeys[it.Group] {
			pe.at(p, "group %q is not declared in metadata.groups", it.Group)
		}

		validateVisibility(p, it.Visibility, fieldOptions, pe)

		if it.Field != nil {
			validateFieldSpec(p+".field", it.Field, pe)
		}
		if it.Action != nil {
			validateActionSpec(p+".action", it.Action, pe)
		}
		if it.Display != nil {
			validateDisplaySpec(p+".display", it.Display, pe)
		}
	}
}

func validateVisibility(base string, v *VisibilityRule, fieldOptions map[string][]string, pe *pathErrors) {
	if v == nil {
		return
	}
	for i, ctx := range v.Contexts {
		if !ValidContexts[ctx] {
			pe.at(fmt.Sprintf("%s.visibility.contexts[%d]", base, i), "unknown context %q", ctx)
		}
	}
	if v.DependsOn != nil {
		p := base + ".visibility.depends_on"
		if v.DependsOn.Field == "" {
			pe.at(p, "field is required")
		} else {
			opts, ok := fieldOptions[v.DependsOn.Field]
			if !ok {
				pe.at(p, "field %q does not reference a defined field item", v.DependsOn.Field)
			} else if len(opts) > 0 {
				found := false
				for _, o := range opts {
					if o == v.DependsOn.Value {
						found = true
						break
					}
				}
				if !found {
					pe.at(p, "value %q is not one of field %q options", v.DependsOn.Value, v.DependsOn.Field)
				}
			}
		}
	}
}

func validateFieldSpec(base string, f *FieldSpec, pe *pathErrors) {
	if !ValidFieldInputs[f.Input] {
		pe.at(base, "input must be one of text|password|number|email|url|select|multiselect|toggle|textarea, got %q", f.Input)
	}
	if f.Input == "select" || f.Input == "multiselect" {
		if len(f.Options) == 0 {
			pe.at(base+".options", "must be non-empty when input=%s", f.Input)
		}
		seen := map[string]bool{}
		for i, o := range f.Options {
			if o.Value == "" {
				pe.at(fmt.Sprintf("%s.options[%d]", base, i), "value is required")
			}
			if o.Label == "" {
				pe.at(fmt.Sprintf("%s.options[%d]", base, i), "label is required")
			}
			if seen[o.Value] {
				pe.at(fmt.Sprintf("%s.options[%d]", base, i), "duplicate option value %q", o.Value)
			}
			seen[o.Value] = true
		}
	} else if len(f.Options) > 0 {
		pe.at(base+".options", "must be empty unless input=select|multiselect")
	}
}

func validateActionSpec(base string, a *ActionSpec, pe *pathErrors) {
	if a.URL == "" {
		pe.at(base+".url", "is required")
	} else {
		tmpl, err := template.New("url").Option("missingkey=error").Parse(a.URL)
		if err != nil {
			pe.at(base+".url", "template parse error: %v", err)
		} else {
			a.urlTmpl = tmpl
		}
	}
	if a.Variant != "" && !ValidActionVariants[a.Variant] {
		pe.at(base+".variant", "must be one of default|outline|destructive, got %q", a.Variant)
	}
	if a.Size != "" && !ValidActionSizes[a.Size] {
		pe.at(base+".size", "must be one of default|icon, got %q", a.Size)
	}
	if a.Target != "" && !ValidActionTargets[a.Target] {
		pe.at(base+".target", "must be one of _blank|_self, got %q", a.Target)
	}
	if a.Variant == "destructive" && a.Confirm == "" {
		pe.at(base+".confirm", "is required when variant=destructive")
	}
}

func validateDisplaySpec(base string, d *DisplaySpec, pe *pathErrors) {
	if d.Value == "" {
		pe.at(base+".value", "is required")
		return
	}
	if d.Format != "" && !ValidDisplayFormats[d.Format] {
		pe.at(base+".format", "must be one of text|code|badge|link|password, got %q", d.Format)
	}
	tmpl, err := template.New("display").Option("missingkey=error").Parse(d.Value)
	if err != nil {
		pe.at(base+".value", "template parse error: %v", err)
	} else {
		d.valueTmpl = tmpl
	}
}

func validateLogs(base string, logs []LogSource, pe *pathErrors) {
	seen := map[string]bool{}
	for j, ls := range logs {
		p := fmt.Sprintf("%s.logs[%d]", base, j)
		if ls.Key == "" {
			pe.at(p, "key is required")
		}
		if ls.Label == "" {
			pe.at(p, "label is required")
		}
		if ls.Type != LogSourceTypeDocker && ls.Type != LogSourceTypeFile && ls.Type != LogSourceTypeLifecycle {
			pe.at(p, "type must be docker|file|lifecycle, got %q", ls.Type)
		}
		if ls.Type == LogSourceTypeFile && ls.Path == "" {
			pe.at(p, "path is required for file log sources")
		}
		if (ls.Type == LogSourceTypeDocker || ls.Type == LogSourceTypeLifecycle) && ls.Path != "" {
			pe.at(p, "path must not be set for %s log sources", ls.Type)
		}
		if seen[ls.Key] {
			pe.at(p, "duplicate key %q", ls.Key)
		}
		seen[ls.Key] = true
	}
}
