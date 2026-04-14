import type { DisplayItem, FieldItem, MetadataContext, MetadataGroup, MetadataItem } from '@/types'

export function itemsForContext(
  metadata: MetadataItem[] | null | undefined,
  ctx: MetadataContext,
): MetadataItem[] {
  if (!metadata) return []
  return metadata.filter((item) => {
    const contexts = item.visibility?.contexts
    if (!contexts || contexts.length === 0) return true
    return contexts.includes(ctx)
  })
}

export function evaluateDependsOn(item: MetadataItem, values: Record<string, string>): boolean {
  const dep = item.visibility?.dependsOn
  if (!dep) return true
  return values[dep.field] === dep.value
}

export function evaluateCondition(
  condition: string | undefined,
  runtime: { ready?: boolean },
): boolean {
  if (!condition) return true
  if (condition === 'ready') return runtime.ready === true
  return true
}

export function groupItems(
  items: MetadataItem[],
  groups: MetadataGroup[] | undefined,
): Array<{ group: MetadataGroup | null; items: MetadataItem[] }> {
  const orderedKeys = (groups ?? []).map((g) => g.key)
  const bucketByKey = new Map<string, MetadataItem[]>()
  for (const key of orderedKeys) bucketByKey.set(key, [])
  const defaults: MetadataItem[] = []

  for (const item of items) {
    if (item.group && bucketByKey.has(item.group)) {
      bucketByKey.get(item.group)!.push(item)
    } else {
      defaults.push(item)
    }
  }

  const result: Array<{ group: MetadataGroup | null; items: MetadataItem[] }> = []
  for (const g of groups ?? []) {
    const bucket = bucketByKey.get(g.key) ?? []
    if (bucket.length > 0) result.push({ group: g, items: bucket })
  }
  if (defaults.length > 0) result.push({ group: null, items: defaults })
  return result
}

export const isFieldItem = (i: MetadataItem): i is FieldItem => i.type === 'field'

export function extractFieldValues(
  metadata: MetadataItem[] | null | undefined,
): Record<string, string> {
  const out: Record<string, string> = {}
  if (!metadata) return out
  for (const item of metadata) {
    if (!isFieldItem(item)) continue
    if (item.field.default !== undefined) out[item.key] = item.field.default
  }
  return out
}

export function stripHiddenValues(
  metadata: MetadataItem[] | null | undefined,
  values: Record<string, string>,
): Record<string, string> {
  if (!metadata) return { ...values }
  const out: Record<string, string> = {}
  for (const item of metadata) {
    if (!isFieldItem(item)) continue
    if (!evaluateDependsOn(item, values)) continue
    if (item.key in values) out[item.key] = values[item.key]
  }
  return out
}

export function isSecretItem(item: MetadataItem): boolean {
  if (item.type === 'field') return item.field.input === 'password'
  if (item.type === 'display') return item.display.format === 'password'
  return false
}

export function fieldAsDisplay(item: FieldItem, copyable = true): DisplayItem {
  return {
    key: item.key,
    label: item.label,
    icon: item.icon,
    group: item.group,
    visibility: item.visibility,
    type: 'display',
    display: {
      value: item.field.default ?? '',
      format: item.field.input === 'password' ? 'password' : 'text',
      copyable,
    },
  }
}

export function maskSecret(value: string): string {
  const len = value?.length ?? 0
  if (len === 0) return '••••••••'
  return '•'.repeat(Math.min(Math.max(len, 4), 12))
}
